package service

import (
	"context"
	"crypto/hmac"
	"errors"
	"fmt"
	"html"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"gorm.io/gorm"
)

const (
	MaxMarketingDigestIntervalDays = 365
	MaxMarketingDigestRecipients   = 100000
	MarketingDigestSendBatchSize   = 100
)

type MarketingDigestGenerationPayload struct {
	PeriodStart int64 `json:"period_start"`
	PeriodEnd   int64 `json:"period_end"`
}

type MarketingDigestSendPayload struct {
	DigestID int64 `json:"digest_id"`
}

type MarketingDigestPromotionInput struct {
	ID          int64  `json:"id"`
	ModelName   string `json:"model_name"`
	Description string `json:"description"`
	Discount    string `json:"discount"`
	StartsAt    int64  `json:"starts_at"`
	EndsAt      int64  `json:"ends_at"`
	URL         string `json:"url"`
	Enabled     bool   `json:"enabled"`
}

func GenerateMarketingDigest(periodStart, periodEnd int64) (*model.MarketingDigest, bool, error) {
	if periodEnd <= 0 {
		periodEnd = common.GetTimestamp()
	}
	if periodStart <= 0 || periodStart >= periodEnd {
		periodStart = periodEnd - 10*24*60*60
	}
	if existing, err := model.GetMarketingDigestByPeriodEnd(periodEnd); err == nil {
		return existing, false, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, err
	}

	var models []model.Model
	if err := model.DB.Select("model_name", "description").
		Where("status = ? AND created_time > ? AND created_time <= ?", 1, periodStart, periodEnd).
		Order("created_time ASC, id ASC").Find(&models).Error; err != nil {
		return nil, false, err
	}
	promotions, err := model.GetActiveMarketingDigestPromotions(periodStart, periodEnd)
	if err != nil {
		return nil, false, err
	}

	modelNames := make([]string, 0, len(models))
	modelDetails := make([]model.Model, 0, len(models))
	for _, item := range models {
		modelNames = append(modelNames, item.ModelName)
		modelDetails = append(modelDetails, item)
	}
	modelNamesJSON, err := common.Marshal(modelNames)
	if err != nil {
		return nil, false, err
	}
	promotionsJSON, err := common.Marshal(promotions)
	if err != nil {
		return nil, false, err
	}
	digest := &model.MarketingDigest{
		PeriodStart:       periodStart,
		PeriodEnd:         periodEnd,
		Subject:           defaultMarketingDigestSubject(),
		Content:           buildMarketingDigestContent(periodStart, periodEnd, modelDetails, promotions),
		ModelNames:        string(modelNamesJSON),
		PromotionSnapshot: string(promotionsJSON),
		AudienceType:      "all",
		AudienceUserIDs:   "[]",
		Status:            model.MarketingDigestStatusPendingReview,
	}
	if err := model.CreateMarketingDigest(digest); err != nil {
		if existing, findErr := model.GetMarketingDigestByPeriodEnd(periodEnd); findErr == nil {
			return existing, false, nil
		}
		return nil, false, err
	}
	notifyMarketingDigestReviewer(digest)
	return digest, true, nil
}

func defaultMarketingDigestSubject() string {
	if common.IsEmailLanguageEnglish() {
		return "{{system_name}} product updates and offers"
	}
	return "{{system_name}} 新模型与优惠"
}

func buildMarketingDigestContent(periodStart, periodEnd int64, models []model.Model, promotions []model.MarketingDigestPromotion) string {
	start := time.Unix(periodStart, 0).Format("2006-01-02")
	end := time.Unix(periodEnd, 0).Format("2006-01-02")
	var b strings.Builder
	if common.IsEmailLanguageEnglish() {
		b.WriteString("<h1>Product updates</h1>")
		b.WriteString("<p>Updates from " + start + " to " + end + ".</p>")
		if len(models) > 0 {
			b.WriteString("<h2>New models</h2><ul>")
			for _, item := range models {
				b.WriteString("<li><strong>" + html.EscapeString(item.ModelName) + "</strong>")
				if item.Description != "" {
					b.WriteString("<br>" + html.EscapeString(item.Description))
				}
				b.WriteString("</li>")
			}
			b.WriteString("</ul>")
		}
		if len(promotions) > 0 {
			b.WriteString("<h2>Current offers</h2><ul>")
			for _, item := range promotions {
				b.WriteString("<li><strong>" + html.EscapeString(item.ModelName) + "</strong>")
				if item.Discount != "" {
					b.WriteString(" — " + html.EscapeString(item.Discount))
				}
				if item.Description != "" {
					b.WriteString("<br>" + html.EscapeString(item.Description))
				}
				if safeMarketingURL(item.URL) != "" {
					b.WriteString("<br><a href=\"" + html.EscapeString(item.URL) + "\">View offer</a>")
				}
				b.WriteString("</li>")
			}
			b.WriteString("</ul>")
		}
	} else {
		b.WriteString("<h1>产品动态</h1>")
		b.WriteString("<p>本期内容统计范围：" + start + " 至 " + end + "。</p>")
		if len(models) > 0 {
			b.WriteString("<h2>近期上新模型</h2><ul>")
			for _, item := range models {
				b.WriteString("<li><strong>" + html.EscapeString(item.ModelName) + "</strong>")
				if item.Description != "" {
					b.WriteString("<br>" + html.EscapeString(item.Description))
				}
				b.WriteString("</li>")
			}
			b.WriteString("</ul>")
		}
		if len(promotions) > 0 {
			b.WriteString("<h2>模型优惠</h2><ul>")
			for _, item := range promotions {
				b.WriteString("<li><strong>" + html.EscapeString(item.ModelName) + "</strong>")
				if item.Discount != "" {
					b.WriteString("：" + html.EscapeString(item.Discount))
				}
				if item.Description != "" {
					b.WriteString("<br>" + html.EscapeString(item.Description))
				}
				if safeMarketingURL(item.URL) != "" {
					b.WriteString("<br><a href=\"" + html.EscapeString(item.URL) + "\">查看活动</a>")
				}
				b.WriteString("</li>")
			}
			b.WriteString("</ul>")
		}
	}
	return b.String()
}

func safeMarketingURL(raw string) string {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" {
		return ""
	}
	return parsed.String()
}

func notifyMarketingDigestReviewer(digest *model.MarketingDigest) {
	if digest == nil {
		return
	}
	reviewEmail := strings.TrimSpace(common.MarketingDigestReviewEmail)
	if reviewEmail == "" {
		if root := model.GetRootUser(); root != nil {
			reviewEmail = strings.TrimSpace(root.Email)
			if reviewEmail == "" {
				reviewEmail = strings.TrimSpace(root.GetSetting().NotificationEmail)
			}
		}
	}
	if reviewEmail == "" {
		common.SysLog("marketing digest draft created without reviewer email")
		return
	}
	base := strings.TrimRight(strings.TrimSpace(system_setting.ServerAddress), "/")
	if base == "" {
		base = "http://localhost:3000"
	}
	link := base + "/email-settings"
	subject := fmt.Sprintf("%s newsletter draft is ready", common.SystemName)
	body := fmt.Sprintf("<p>A new product update email is ready for review.</p><p>Period: %s to %s</p><p><a href=\"%s\">Open email settings</a> and review draft #%d before sending.</p>", time.Unix(digest.PeriodStart, 0).Format("2006-01-02"), time.Unix(digest.PeriodEnd, 0).Format("2006-01-02"), html.EscapeString(link), digest.ID)
	if err := common.SendEmail(subject, reviewEmail, body); err != nil {
		common.SysLog(fmt.Sprintf("failed to notify newsletter reviewer for draft %d: %v", digest.ID, err))
	}
}

func MarketingDigestUnsubscribeURL(userID int) string {
	if userID <= 0 {
		return ""
	}
	base := strings.TrimRight(strings.TrimSpace(system_setting.ServerAddress), "/")
	if base == "" {
		base = "http://localhost:3000"
	}
	token := common.GenerateHMAC("marketing-unsubscribe:" + strconv.Itoa(userID))
	return fmt.Sprintf("%s/api/marketing_email/unsubscribe?user_id=%d&token=%s", base, userID, token)
}

func VerifyMarketingDigestUnsubscribeToken(userID int, token string) bool {
	if userID <= 0 || token == "" {
		return false
	}
	expected := common.GenerateHMAC("marketing-unsubscribe:" + strconv.Itoa(userID))
	return hmac.Equal([]byte(strings.ToLower(token)), []byte(expected))
}

func SendMarketingDigestEmail(user *model.User, subjectTemplate, contentTemplate string, send func(string, string, string) error) error {
	if !model.IsMarketingDigestUserEligible(user) {
		return errors.New("marketing email preference is disabled or account is unavailable")
	}
	userSetting := user.GetSetting()
	emailToUse := strings.TrimSpace(userSetting.NotificationEmail)
	if emailToUse == "" {
		emailToUse = strings.TrimSpace(user.Email)
	}
	if emailToUse == "" {
		return errors.New("user has no email")
	}
	if send == nil {
		send = common.SendEmail
	}
	subject, content := common.BuildMarketingMail(common.MarketingMailParams{
		SystemName:      common.SystemName,
		Username:        user.Username,
		DisplayName:     user.DisplayName,
		Email:           emailToUse,
		UserId:          strconv.Itoa(user.Id),
		UnsubscribeURL:  MarketingDigestUnsubscribeURL(user.Id),
		SubjectTemplate: subjectTemplate,
		ContentTemplate: contentTemplate,
	})
	unsubscribeURL := MarketingDigestUnsubscribeURL(user.Id)
	if unsubscribeURL != "" && !strings.Contains(content, unsubscribeURL) {
		content += `<p style="margin-top:24px;color:#6b7280;font-size:12px">To stop receiving product updates, <a href="` + html.EscapeString(unsubscribeURL) + `">unsubscribe here</a>.</p>`
	}
	return send(subject, emailToUse, content)
}

func GenerateManualMarketingDigest(periodDays int) (*model.MarketingDigest, bool, error) {
	if periodDays < 1 || periodDays > MaxMarketingDigestIntervalDays {
		periodDays = common.MarketingDigestIntervalDays
	}
	if periodDays < 1 || periodDays > MaxMarketingDigestIntervalDays {
		periodDays = 10
	}
	end := common.GetTimestamp()
	return GenerateMarketingDigest(end-int64(periodDays*24*60*60), end)
}

func MarketingDigestScheduleEnabled() bool {
	return common.MarketingDigestIntervalDays > 0 && common.MarketingDigestIntervalDays <= MaxMarketingDigestIntervalDays
}

func MarketingDigestGenerationTaskPayload() MarketingDigestGenerationPayload {
	end := common.GetTimestamp()
	start := end - int64(common.MarketingDigestIntervalDays*24*60*60)
	if latest, err := model.GetLatestMarketingDigest(); err == nil && latest.PeriodEnd < end {
		start = latest.PeriodEnd
	}
	return MarketingDigestGenerationPayload{PeriodStart: start, PeriodEnd: end}
}

func RunMarketingDigestGenerationTask(ctx context.Context, task *model.SystemTask, runnerID string) {
	var payload MarketingDigestGenerationPayload
	if err := task.DecodePayload(&payload); err != nil {
		failSystemTask(task, runnerID, err)
		return
	}
	if err := ctx.Err(); err != nil {
		failSystemTask(task, runnerID, err)
		return
	}
	_, _, err := GenerateMarketingDigest(payload.PeriodStart, payload.PeriodEnd)
	if err != nil {
		failSystemTask(task, runnerID, err)
		return
	}
	if err := model.FinishSystemTask(task.TaskID, runnerID, model.SystemTaskStatusSucceeded, map[string]any{"period_start": payload.PeriodStart, "period_end": payload.PeriodEnd}, ""); err != nil {
		common.SysLog(fmt.Sprintf("failed to finish marketing digest generation task: %v", err))
	}
}

func RunMarketingDigestSendTask(ctx context.Context, task *model.SystemTask, runnerID string) {
	var payload MarketingDigestSendPayload
	if err := task.DecodePayload(&payload); err != nil {
		failSystemTask(task, runnerID, err)
		return
	}
	digest, err := model.GetMarketingDigest(payload.DigestID)
	if err != nil {
		failSystemTask(task, runnerID, err)
		return
	}
	if err := runMarketingDigestSend(ctx, digest); err != nil {
		if ctx.Err() != nil {
			return
		}
		_ = model.FinishStaleMarketingDigestRecipients(digest.ID)
		_ = model.SetMarketingDigestStatus(digest.ID, model.MarketingDigestStatusFailed, 0, 0)
		failSystemTask(task, runnerID, err)
		return
	}
	counts, err := model.GetMarketingDigestRecipientCounts(digest.ID)
	if err != nil {
		failSystemTask(task, runnerID, err)
		return
	}
	status := model.MarketingDigestStatusSent
	if counts.Failed > 0 {
		if counts.Sent == 0 {
			status = model.MarketingDigestStatusFailed
		} else {
			status = model.MarketingDigestStatusSentWithError
		}
	}
	if counts.Pending > 0 {
		status = model.MarketingDigestStatusFailed
	}
	if err := model.SetMarketingDigestStatus(digest.ID, status, 0, common.GetTimestamp()); err != nil {
		failSystemTask(task, runnerID, err)
		return
	}
	if err := model.FinishSystemTask(task.TaskID, runnerID, model.SystemTaskStatusSucceeded, counts, ""); err != nil {
		common.SysLog(fmt.Sprintf("failed to finish marketing digest send task: %v", err))
	}
}

func runMarketingDigestSend(ctx context.Context, digest *model.MarketingDigest) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		batch, err := model.ClaimMarketingDigestRecipients(digest.ID, MarketingDigestSendBatchSize)
		if err != nil {
			return err
		}
		if len(batch) == 0 {
			return nil
		}
		for _, recipient := range batch {
			if err := ctx.Err(); err != nil {
				return err
			}
			var user model.User
			err := model.DB.Unscoped().Where("id = ?", recipient.UserID).First(&user).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if finishErr := model.FinishMarketingDigestRecipient(recipient.ID, model.MarketingDigestRecipientSkipped, "", "user not found"); finishErr != nil {
					return finishErr
				}
				continue
			}
			if err != nil {
				return err
			}
			if !model.IsMarketingDigestUserEligible(&user) {
				if finishErr := model.FinishMarketingDigestRecipient(recipient.ID, model.MarketingDigestRecipientSkipped, "", "user opted out or account is unavailable"); finishErr != nil {
					return finishErr
				}
				continue
			}
			userSetting := user.GetSetting()
			emailToUse := strings.TrimSpace(userSetting.NotificationEmail)
			if emailToUse == "" {
				emailToUse = strings.TrimSpace(user.Email)
			}
			if emailToUse == "" {
				if finishErr := model.FinishMarketingDigestRecipient(recipient.ID, model.MarketingDigestRecipientSkipped, "", "user has no email"); finishErr != nil {
					return finishErr
				}
				continue
			}
			err = SendMarketingDigestEmail(&user, digest.Subject, digest.Content, common.SendEmail)
			if err != nil {
				if finishErr := model.FinishMarketingDigestRecipient(recipient.ID, model.MarketingDigestRecipientFailed, common.MaskEmail(emailToUse), err.Error()); finishErr != nil {
					return finishErr
				}
				continue
			}
			if err := model.FinishMarketingDigestRecipient(recipient.ID, model.MarketingDigestRecipientSent, common.MaskEmail(emailToUse), ""); err != nil {
				return err
			}
		}
	}
}

func StartMarketingDigestSend(digestID int64, reviewerID int, retryFailed bool) (*model.SystemTask, error) {
	digest, err := model.GetMarketingDigest(digestID)
	if err != nil {
		return nil, err
	}
	active, err := model.GetActiveSystemTask(model.SystemTaskTypeMarketingDigestSend)
	if err != nil {
		return nil, err
	}
	if active != nil {
		var activePayload MarketingDigestSendPayload
		if err := active.DecodePayload(&activePayload); err == nil && activePayload.DigestID == digestID {
			return active, nil
		}
		return nil, errors.New("another newsletter is currently being sent")
	}
	previousStatus := digest.Status

	if retryFailed {
		if digest.Status != model.MarketingDigestStatusSending && digest.Status != model.MarketingDigestStatusSentWithError && digest.Status != model.MarketingDigestStatusFailed {
			return nil, errors.New("newsletter has no failed recipients to retry")
		}
		if digest.Status == model.MarketingDigestStatusSending {
			if err := model.FinishStaleMarketingDigestRecipients(digestID); err != nil {
				return nil, err
			}
		}
		if _, err := model.RetryFailedMarketingDigestRecipients(digestID); err != nil {
			return nil, err
		}
	} else {
		if digest.Status != model.MarketingDigestStatusPendingReview {
			return nil, errors.New("newsletter is not awaiting review")
		}
		var userIDs []int
		if err := common.Unmarshal([]byte(digest.AudienceUserIDs), &userIDs); err != nil {
			return nil, fmt.Errorf("invalid recipient selection: %w", err)
		}
		ids, err := model.GetMarketingDigestAudienceUserIDs(digest.AudienceType, digest.AudienceGroup, userIDs, model.MarketingDigestActiveSince(time.Now()), MaxMarketingDigestRecipients)
		if err != nil {
			return nil, err
		}
		if len(ids) == 0 {
			return nil, errors.New("selected audience has no eligible users")
		}
		if err := model.CreateMarketingDigestRecipients(digestID, ids); err != nil {
			return nil, err
		}
	}
	if err := model.SetMarketingDigestStatus(digestID, model.MarketingDigestStatusSending, reviewerID, 0); err != nil {
		if !retryFailed {
			_ = model.DeleteMarketingDigestRecipients(digestID)
		}
		return nil, err
	}
	task, _, err := EnqueueSystemTask(model.SystemTaskTypeMarketingDigestSend, MarketingDigestSendPayload{DigestID: digestID})
	if err != nil {
		_ = model.SetMarketingDigestStatus(digestID, previousStatus, 0, 0)
		if !retryFailed {
			_ = model.DeleteMarketingDigestRecipients(digestID)
		}
		return nil, err
	}
	var taskPayload MarketingDigestSendPayload
	if err := task.DecodePayload(&taskPayload); err != nil {
		_ = model.SetMarketingDigestStatus(digestID, previousStatus, 0, 0)
		if !retryFailed {
			_ = model.DeleteMarketingDigestRecipients(digestID)
		}
		return nil, fmt.Errorf("existing newsletter task has invalid payload: %w", err)
	}
	if taskPayload.DigestID != digestID {
		_ = model.SetMarketingDigestStatus(digestID, previousStatus, 0, 0)
		if !retryFailed {
			_ = model.DeleteMarketingDigestRecipients(digestID)
		}
		return nil, errors.New("another newsletter is currently being sent")
	}
	return task, nil
}

func MarketingDigestIntervalOptionValue() string {
	return strconv.Itoa(common.MarketingDigestIntervalDays)
}
