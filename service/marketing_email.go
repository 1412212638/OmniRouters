package service

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"html"
	"net/url"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/system_setting"
)

const (
	MaxMarketingEmailRecipients  = 100000
	marketingEmailCountBatchSize = 500
)

type MarketingEmailResult struct {
	Total       int `json:"total"`
	Sent        int `json:"sent"`
	Skipped     int `json:"skipped"`
	Failed      int `json:"failed"`
	Unconfirmed int `json:"unconfirmed,omitempty"`
}

func GetMarketingEmailAudienceCount(audienceType, group string) (int, error) {
	ids, err := model.GetMarketingEmailAudienceUserIDs(audienceType, group, nil, MaxMarketingEmailRecipients)
	if err != nil {
		return 0, err
	}
	total := 0
	for start := 0; start < len(ids); start += marketingEmailCountBatchSize {
		end := start + marketingEmailCountBatchSize
		if end > len(ids) {
			end = len(ids)
		}
		users, err := model.GetUsersByIds(ids[start:end])
		if err != nil {
			return 0, err
		}
		for i := range users {
			if marketingEmailUserHasDestination(&users[i]) {
				total++
			}
		}
	}
	return total, nil
}

type MarketingEmailSendPayload struct {
	UserIDs          []int  `json:"user_ids"`
	SubjectTemplate  string `json:"subject_template"`
	ContentTemplate  string `json:"content_template"`
	DeliveryRevision string `json:"delivery_revision,omitempty"`
}

func StartMarketingEmailSend(audienceType, group string, userIds []int, subjectTemplate, contentTemplate string) (*model.SystemTask, error) {
	cfg, err := model.GetMarketingSMTPConfig()
	if err != nil {
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if err := validateMarketingUnsubscribeBase(cfg); err != nil {
		return nil, err
	}
	ids := uniquePositiveUserIds(userIds)
	if audienceType == "manual" && len(ids) > MaxMarketingEmailRecipients {
		return nil, fmt.Errorf("单次最多发送 %d 个用户", MaxMarketingEmailRecipients)
	}
	ids, err = model.GetMarketingEmailAudienceUserIDs(audienceType, group, ids, MaxMarketingEmailRecipients)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, errors.New("没有符合条件的收件人")
	}
	task, created, err := EnqueueSystemTask(model.SystemTaskTypeMarketingEmailSend, MarketingEmailSendPayload{
		UserIDs: ids, SubjectTemplate: subjectTemplate, ContentTemplate: contentTemplate, DeliveryRevision: marketingDeliveryRevision(cfg),
	})
	if err != nil {
		return nil, err
	}
	if !created {
		return nil, errors.New("a marketing email campaign is already in progress")
	}
	return task, nil
}

func validateMarketingUnsubscribeBase(cfg common.MarketingSMTPConfig) error {
	if cfg.Mode != "smtp" {
		return nil
	}
	base, err := url.Parse(strings.TrimSpace(system_setting.ServerAddress))
	if err != nil || base.Scheme != "https" || base.Hostname() == "" || base.User != nil || base.RawQuery != "" || base.Fragment != "" {
		return errors.New("configure a public HTTPS server address for marketing unsubscribe links")
	}
	return nil
}

func marketingDeliveryRevision(cfg common.MarketingSMTPConfig) string {
	raw, _ := common.Marshal(cfg)
	return common.GenerateHMAC("marketing-delivery:" + string(raw))
}

func MarketingSESProof(messageID, from, receiver string) string {
	messageID = strings.Trim(strings.TrimSpace(messageID), "<>")
	return common.GenerateHMAC("marketing-ses:" + messageID + "\n" + strings.ToLower(from) + "\n" + strings.ToLower(receiver))
}

func SendMarketingMessage(ctx context.Context, cfg common.MarketingSMTPConfig, subject, receiver, content, unsubscribe string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if cfg.Mode == "inherit" {
		return common.SendEmail(subject, receiver, content)
	}
	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		return err
	}
	_, domain, _ := strings.Cut(cfg.From, "@")
	messageID := hex.EncodeToString(id[:]) + "@" + domain
	return cfg.Send(ctx, subject, receiver, content, unsubscribe, messageID, MarketingSESProof(messageID, cfg.From, receiver))
}

func marketingEmailDestination(user *model.User) string {
	if address := strings.TrimSpace(user.GetSetting().NotificationEmail); address != "" {
		return address
	}
	return strings.TrimSpace(user.Email)
}

func marketingEmailUserHasDestination(user *model.User) bool {
	if !model.IsMarketingEmailUserEligible(user) {
		return false
	}
	return marketingEmailDestination(user) != ""
}

func MarketingEmailUnsubscribeURL(userID int) string {
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

func VerifyMarketingEmailUnsubscribeToken(userID int, token string) bool {
	if userID <= 0 || token == "" {
		return false
	}
	expected := common.GenerateHMAC("marketing-unsubscribe:" + strconv.Itoa(userID))
	return hmac.Equal([]byte(strings.ToLower(token)), []byte(expected))
}

func SendMarketingEmailToUser(user *model.User, subjectTemplate, contentTemplate string, send func(string, string, string) error) error {
	if !model.IsMarketingEmailUserEligible(user) {
		return errors.New("marketing email preference is disabled or account is unavailable")
	}
	emailToUse := marketingEmailDestination(user)
	if emailToUse == "" {
		return errors.New("user has no email")
	}
	if send == nil {
		cfg, err := model.GetMarketingSMTPConfig()
		if err != nil {
			return err
		}
		if err := cfg.Validate(); err != nil {
			return err
		}
		send = func(subject, receiver, content string) error {
			return SendMarketingMessage(context.Background(), cfg, subject, receiver, content, MarketingEmailUnsubscribeURL(user.Id))
		}
	}
	unsubscribeURL := MarketingEmailUnsubscribeURL(user.Id)
	subject, content := common.BuildMarketingMail(common.MarketingMailParams{
		SystemName:      common.SystemName,
		Username:        user.Username,
		DisplayName:     user.DisplayName,
		Email:           emailToUse,
		UserId:          strconv.Itoa(user.Id),
		UnsubscribeURL:  unsubscribeURL,
		SubjectTemplate: subjectTemplate,
		ContentTemplate: contentTemplate,
	})
	if unsubscribeURL != "" && !strings.Contains(content, unsubscribeURL) {
		content += `<p style="margin-top:24px;color:#6b7280;font-size:12px">To stop receiving marketing emails, <a href="` + html.EscapeString(unsubscribeURL) + `">unsubscribe here</a>.</p>`
	}
	return send(subject, emailToUse, content)
}

func uniquePositiveUserIds(userIds []int) []int {
	seen := make(map[int]struct{}, len(userIds))
	ids := make([]int, 0, len(userIds))
	for _, id := range userIds {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	return ids
}
