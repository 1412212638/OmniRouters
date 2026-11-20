package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"gorm.io/gorm"
)

type marketingEmailSendHandler struct{}

func init() {
	RegisterSystemTaskHandler(marketingEmailSendHandler{})
}

func (marketingEmailSendHandler) Type() string { return model.SystemTaskTypeMarketingEmailSend }

func (marketingEmailSendHandler) Run(ctx context.Context, task *model.SystemTask, runnerID string) {
	payload := MarketingEmailSendPayload{}
	if err := task.DecodePayload(&payload); err != nil {
		failSystemTask(task, runnerID, err)
		return
	}
	if len(payload.UserIDs) == 0 {
		failSystemTask(task, runnerID, errors.New("marketing email audience is empty"))
		return
	}

	cfg, err := model.GetMarketingSMTPConfig()
	if err == nil {
		err = cfg.Validate()
	}
	if err == nil && payload.DeliveryRevision != "" && payload.DeliveryRevision != marketingDeliveryRevision(cfg) {
		err = errors.New("marketing delivery settings changed after queueing; create a new campaign")
	}
	if err == nil && payload.DeliveryRevision == "" && cfg.Mode != "inherit" {
		err = errors.New("campaign was queued before independent marketing delivery was configured")
	}
	if err == nil {
		err = validateMarketingUnsubscribeBase(cfg)
	}
	if err != nil {
		failSystemTask(task, runnerID, err)
		return
	}

	result := &MarketingEmailResult{Total: len(payload.UserIDs)}
	interval := time.Second / time.Duration(cfg.PerSecond)
	var nextSend time.Time
	lastProgress := time.Now()
	for i, userID := range payload.UserIDs {
		if err := deliverCampaignEmail(ctx, cfg, interval, &nextSend, userID, payload, result); err != nil {
			if finishErr := model.FinishSystemTask(task.TaskID, runnerID, model.SystemTaskStatusFailed, result, err.Error()); finishErr != nil {
				logSystemTaskLockError(ctx, task, finishErr)
			}
			return
		}
		processed := i + 1
		if processed != result.Total && time.Since(lastProgress) < 5*time.Second {
			continue
		}
		state := SystemTaskProgress{Total: result.Total, Processed: processed, Progress: processed * 100 / result.Total}
		if err := model.UpdateSystemTaskState(task.TaskID, runnerID, state); err != nil {
			logSystemTaskLockError(ctx, task, err)
			return
		}
		lastProgress = time.Now()
	}
	common.SysLog(fmt.Sprintf("marketing email campaign completed: total=%d submitted=%d skipped=%d failed=%d unconfirmed=%d", result.Total, result.Sent, result.Skipped, result.Failed, result.Unconfirmed))
	if err := model.FinishSystemTask(task.TaskID, runnerID, model.SystemTaskStatusSucceeded, result, ""); err != nil {
		logSystemTaskLockError(ctx, task, err)
	}
}

func deliverCampaignEmail(ctx context.Context, cfg common.MarketingSMTPConfig, interval time.Duration, nextSend *time.Time, userID int, payload MarketingEmailSendPayload, result *MarketingEmailResult) error {
	for attempt := 0; ; attempt++ {
		if err := waitMarketingEmail(ctx, time.Until(*nextSend)); err != nil {
			return err
		}
		user, err := model.GetUserById(userID, false)
		if errors.Is(err, gorm.ErrRecordNotFound) || (err == nil && !marketingEmailUserHasDestination(user)) {
			result.Skipped++
			return nil
		}
		if err != nil {
			return err
		}
		suppressed, err := model.IsMarketingEmailSuppressed(marketingEmailDestination(user))
		if err != nil {
			return err
		}
		if suppressed {
			result.Skipped++
			return nil
		}
		err = SendMarketingEmailToUser(user, payload.SubjectTemplate, payload.ContentTemplate, func(subject, recipient, content string) error {
			*nextSend = time.Now().Add(interval)
			return SendMarketingMessage(ctx, cfg, subject, recipient, content, MarketingEmailUnsubscribeURL(userID))
		})
		if err == nil {
			result.Sent++
			return nil
		}
		var smtpErr *common.MarketingSMTPError
		if !errors.As(err, &smtpErr) {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			result.Failed++
			return nil
		}
		switch {
		case smtpErr.Fatal:
			result.Failed++
			return err
		case smtpErr.Unknown:
			result.Unconfirmed++
			return nil
		case !smtpErr.Retryable || attempt >= 2:
			result.Failed++
			return nil
		}
		if err := waitMarketingEmail(ctx, time.Second<<attempt); err != nil {
			return err
		}
	}
}

func waitMarketingEmail(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return ctx.Err()
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return ctx.Err()
	}
}
