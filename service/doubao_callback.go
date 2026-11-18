package service

import (
	"bytes"
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	"github.com/bytedance/gopkg/util/gopool"
	"gorm.io/gorm"
)

const DoubaoCallbackPath = "/api/v3/task-notifications/"

// Activate only after initial reservation settlement and durable task creation.
// An early vendor notification must not race submission-time accounting.
func ActivateDoubaoCallback(task *model.Task) {
	if task == nil || string(task.Platform) != "doubao" {
		return
	}
	if err := model.DB.Model(&model.TaskCallback{}).Where("task_id = ?", task.TaskID).Update("ready_at", time.Now().Unix()).Error; err != nil {
		common.SysError("failed to activate task notification task=" + task.TaskID)
	}
}

// PrepareDoubaoCallback persists the destination before contacting Ark. Only
// an unguessable per-task token is sent upstream; customer URLs stay private.
func PrepareDoubaoCallback(taskID, target string) (string, error) {
	if err := validateTaskCallbackURL(target); err != nil {
		return "", err
	}
	base := strings.TrimRight(GetCallbackAddress(), "/")
	u, err := url.Parse(base)
	if err != nil || u.Host == "" || (u.Scheme != "https" && u.Scheme != "http") || u.RawQuery != "" || u.Fragment != "" || u.User != nil {
		return "", fmt.Errorf("configure the public server address before using callback_url")
	}
	var row model.TaskCallback
	err = model.DB.Where("task_id = ?", taskID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		token, tokenErr := common.GenerateRandomCharsKey(48)
		if tokenErr != nil {
			return "", tokenErr
		}
		row = model.TaskCallback{TaskID: taskID, TargetURL: target, Token: token, CreatedAt: time.Now().Unix()}
		err = model.DB.Create(&row).Error
	}
	if err != nil {
		return "", err
	}
	if row.TargetURL != target {
		return "", fmt.Errorf("callback_url changed during task submission")
	}
	return base + DoubaoCallbackPath + url.PathEscape(taskID) + "?token=" + url.QueryEscape(row.Token), nil
}

func validateTaskCallbackURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || u.User != nil || u.Fragment != "" || (u.Scheme != "http" && u.Scheme != "https") || len(raw) > 4096 {
		return fmt.Errorf("callback_url must be an HTTP or HTTPS URL without credentials or fragment")
	}
	return ValidateSSRFProtectedFetchURL(raw)
}

// ReceiveDoubaoCallback treats the notification as a wake-up signal. Even a
// leaked/replayed token cannot forge usage: fetch authenticated provider state.
func ReceiveDoubaoCallback(ctx context.Context, taskID, token, upstreamID string) error {
	var row model.TaskCallback
	if err := model.DB.WithContext(ctx).Where("task_id = ?", taskID).First(&row).Error; err != nil {
		return err
	}
	if token == "" || subtle.ConstantTimeCompare([]byte(token), []byte(row.Token)) != 1 {
		return fmt.Errorf("invalid notification token")
	}
	if row.ReadyAt == 0 {
		return fmt.Errorf("task submission is not ready")
	}
	var task model.Task
	if err := model.DB.WithContext(ctx).Where("task_id = ? AND platform = ?", taskID, "doubao").First(&task).Error; err != nil {
		return err
	}
	if upstreamID != task.GetUpstreamTaskID() {
		return fmt.Errorf("notification task ID mismatch")
	}
	if isNonTerminalPollStatus(task.Status) {
		if GetTaskAdaptorFunc == nil {
			return fmt.Errorf("task polling is unavailable")
		}
		now := time.Now().Unix()
		claim := model.DB.Model(&model.TaskCallback{}).Where("id = ? AND refresh_after <= ?", row.ID, now).Update("refresh_after", now+30)
		if claim.Error != nil {
			return claim.Error
		}
		if claim.RowsAffected == 1 {
			gopool.Go(func() {
				refreshCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
				defer cancel()
				if err := updateVideoTasks(refreshCtx, constant.TaskPlatform("doubao"), task.ChannelId, []string{upstreamID}, map[string]*model.Task{upstreamID: &task}); err != nil {
					common.SysError(fmt.Sprintf("task notification refresh failed task=%s", task.TaskID))
				}
				dispatchDoubaoCallback(refreshCtx, row)
			})
		}
	} else if row.DeliveredStatus != string(task.Status) && row.NextAttemptAt <= time.Now().Unix() {
		gopool.Go(func() { dispatchDoubaoCallback(context.Background(), row) })
	}
	// Delivery is decoupled from the vendor's five-second acknowledgement.
	return nil
}

// DispatchDoubaoCallbacks runs with the existing polling system task. The DB
// lease coordinates instances; failed deliveries retain their retry schedule.
func DispatchDoubaoCallbacks(ctx context.Context) {
	// Slow customer endpoints cannot stall the shared billing polling task.
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	var rows []model.TaskCallback
	now := time.Now().Unix()
	model.DB.Where("created_at < ?", now-16*86400).Delete(&model.TaskCallback{})
	if model.DB.Where("ready_at > 0 AND next_attempt_at <= ? AND lease_until <= ? AND created_at >= ?", now, now, now-15*86400).
		Order("next_attempt_at ASC, id ASC").Limit(1000).Find(&rows).Error != nil {
		return
	}
	var workers sync.WaitGroup
	jobs := make(chan model.TaskCallback)
	for range 8 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for row := range jobs {
				if ctx.Err() == nil {
					dispatchDoubaoCallback(ctx, row)
				}
			}
		}()
	}
	for _, row := range rows {
		select {
		case jobs <- row:
		case <-ctx.Done():
		}
		if ctx.Err() != nil {
			break
		}
	}
	close(jobs)
	workers.Wait()
}

func dispatchDoubaoCallback(ctx context.Context, row model.TaskCallback) {
	now := time.Now().Unix()
	claim := model.DB.Model(&model.TaskCallback{}).Where("id = ? AND lease_until <= ?", row.ID, now).
		Where("ready_at > 0 AND next_attempt_at <= ?", now).
		Update("lease_until", now+30)
	if claim.Error != nil || claim.RowsAffected != 1 {
		return
	}
	updates := map[string]any{"lease_until": int64(0), "next_attempt_at": now + 30}
	defer func() {
		model.DB.Model(&model.TaskCallback{}).Where("id = ? AND lease_until = ?", row.ID, now+30).Updates(updates)
	}()
	// The list snapshot may predate another worker's successful delivery.
	if model.DB.First(&row, row.ID).Error != nil {
		return
	}
	var task model.Task
	if model.DB.Where("task_id = ? AND platform = ?", row.TaskID, "doubao").First(&task).Error != nil {
		return // May arrive before the task's durable submission barrier.
	}
	status := string(task.Status)
	if status == row.DeliveredStatus {
		return
	}
	payload, err := doubaoCallbackPayload(&task)
	if err == nil {
		err = sendTaskCallback(ctx, row.TargetURL, payload)
	}
	if err != nil {
		attempts := row.Attempts + 1
		if attempts > 8 {
			attempts = 8
		}
		updates["attempts"] = attempts
		updates["next_attempt_at"] = now + int64(15*(1<<attempts))
		common.SysError(fmt.Sprintf("task notification delivery failed task=%s attempt=%d", row.TaskID, attempts))
		return
	}
	updates["delivered_status"] = status
	updates["attempts"] = 0
	if task.Status == model.TaskStatusSuccess || task.Status == model.TaskStatusFailure {
		updates["next_attempt_at"] = now + 16*86400
	}
}

func doubaoCallbackPayload(task *model.Task) ([]byte, error) {
	var data map[string]any
	if err := common.Unmarshal(task.Data, &data); err != nil || data == nil {
		return nil, fmt.Errorf("task notification data is unavailable")
	}
	status := map[model.TaskStatus]string{model.TaskStatusSubmitted: "queued", model.TaskStatusQueued: "queued", model.TaskStatusInProgress: "running", model.TaskStatusSuccess: "succeeded", model.TaskStatusFailure: "failed"}[task.Status]
	if status == "" {
		return nil, fmt.Errorf("task notification status is unavailable")
	}
	// Preserve upstream failure variants, e.g. expired/cancelled.
	if task.Status == model.TaskStatusFailure && (data["status"] == "expired" || data["status"] == "cancelled") {
		status = data["status"].(string)
	}
	data["id"], data["status"] = task.TaskID, status
	if task.Properties.OriginModelName != "" {
		data["model"] = task.Properties.OriginModelName
	}
	delete(data, "callback_url")
	return common.Marshal(data)
}

func sendTaskCallback(ctx context.Context, target string, body []byte) error {
	if err := validateTaskCallbackURL(target); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	// Never redirect customer notifications or attach gateway/provider credentials.
	client := *GetSSRFProtectedHTTPClient()
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("notification HTTP %d", resp.StatusCode)
	}
	return nil
}
