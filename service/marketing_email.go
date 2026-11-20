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
	htmlparser "golang.org/x/net/html"
)

const (
	MaxMarketingEmailRecipients  = 100000
	maxCustomMarketingEmails     = 100
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
	UserIDs          []int    `json:"user_ids"`
	Emails           []string `json:"emails,omitempty"`
	SubjectTemplate  string   `json:"subject_template"`
	ContentTemplate  string   `json:"content_template"`
	DeliveryRevision string   `json:"delivery_revision,omitempty"`
}

func StartMarketingEmailSend(audienceType, group string, userIds []int, emails []string, subjectTemplate, contentTemplate string) (*model.SystemTask, error) {
	cfg, err := model.GetMarketingSMTPConfig()
	if err != nil {
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if cfg.Mode == "smtp" || audienceType == "custom" {
		if err := validateMarketingUnsubscribeBase(); err != nil {
			return nil, err
		}
	}
	payload := MarketingEmailSendPayload{
		SubjectTemplate: subjectTemplate, ContentTemplate: contentTemplate, DeliveryRevision: marketingDeliveryRevision(cfg),
	}
	if audienceType == "custom" {
		payload.Emails, err = normalizeCustomMarketingEmails(emails)
		if err != nil {
			return nil, err
		}
	} else {
		ids := uniquePositiveUserIds(userIds)
		if audienceType == "manual" && len(ids) > MaxMarketingEmailRecipients {
			return nil, fmt.Errorf("单次最多发送 %d 个用户", MaxMarketingEmailRecipients)
		}
		payload.UserIDs, err = model.GetMarketingEmailAudienceUserIDs(audienceType, group, ids, MaxMarketingEmailRecipients)
		if err != nil {
			return nil, err
		}
		if len(payload.UserIDs) == 0 {
			return nil, errors.New("没有符合条件的收件人")
		}
	}
	task, created, err := EnqueueSystemTask(model.SystemTaskTypeMarketingEmailSend, payload)
	if err != nil {
		return nil, err
	}
	if !created {
		return nil, errors.New("a marketing email campaign is already in progress")
	}
	return task, nil
}

func validateMarketingUnsubscribeBase() error {
	base, err := url.Parse(strings.TrimSpace(system_setting.ServerAddress))
	if err != nil || base.Scheme != "https" || base.Hostname() == "" || base.User != nil || base.RawQuery != "" || base.Fragment != "" {
		return errors.New("configure a public HTTPS server address for marketing unsubscribe links")
	}
	return nil
}

func normalizeCustomMarketingEmails(emails []string) ([]string, error) {
	if len(emails) == 0 || len(emails) > maxCustomMarketingEmails {
		return nil, fmt.Errorf("enter between 1 and %d email addresses", maxCustomMarketingEmails)
	}
	unique := make([]string, 0, len(emails))
	seen := make(map[string]struct{}, len(emails))
	for _, raw := range emails {
		address := strings.TrimSpace(raw)
		if len(address) > 254 || !common.IsPlainEmailAddress(address) {
			return nil, errors.New("invalid custom recipient email address")
		}
		key := strings.ToLower(address)
		if _, exists := seen[key]; !exists {
			seen[key] = struct{}{}
			unique = append(unique, address)
		}
	}
	return unique, nil
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
	token := common.GenerateHMAC("marketing-unsubscribe:" + strconv.Itoa(userID))
	return marketingUnsubscribeURL("user_id=" + strconv.Itoa(userID) + "&token=" + token)
}

func marketingAddressUnsubscribeURL(address string) string {
	hash := model.MarketingEmailAddressHash(address)
	query := url.Values{
		"email_hash": {hash}, "token": {common.GenerateHMAC("marketing-address-unsubscribe:" + hash)},
	}
	return marketingUnsubscribeURL(query.Encode())
}

func marketingUnsubscribeURL(query string) string {
	base := strings.TrimRight(strings.TrimSpace(system_setting.ServerAddress), "/")
	if base == "" {
		base = "http://localhost:3000"
	}
	return base + "/api/marketing_email/unsubscribe?" + query
}

func VerifyMarketingAddressUnsubscribeToken(hash, token string) bool {
	if len(hash) != 64 {
		return false
	}
	if _, err := hex.DecodeString(hash); err != nil || hash != strings.ToLower(hash) {
		return false
	}
	expected := common.GenerateHMAC("marketing-address-unsubscribe:" + hash)
	return hmac.Equal([]byte(token), []byte(expected))
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
	subject, content := buildMarketingRecipientMail(user, emailToUse, unsubscribeURL, subjectTemplate, contentTemplate)
	return send(subject, emailToUse, content)
}

func buildMarketingRecipientMail(user *model.User, address, unsubscribeURL, subjectTemplate, contentTemplate string) (string, string) {
	params := common.MarketingMailParams{
		Email: address, Username: address, UnsubscribeURL: unsubscribeURL,
		SubjectTemplate: subjectTemplate, ContentTemplate: contentTemplate,
	}
	if user != nil {
		params.Username, params.DisplayName, params.UserId = user.Username, user.DisplayName, strconv.Itoa(user.Id)
	}
	subject, content := common.BuildMarketingMail(params)
	return subject, addMarketingUnsubscribeFooter(content, unsubscribeURL)
}

func addMarketingUnsubscribeFooter(content, unsubscribeURL string) string {
	if unsubscribeURL == "" {
		return content
	}
	tokens := htmlparser.NewTokenizer(strings.NewReader(content))
	offset, bodyEnd, documentEnd := 0, -1, -1
	for kind := tokens.Next(); kind != htmlparser.ErrorToken; kind = tokens.Next() {
		start := offset
		offset += len(tokens.Raw())
		switch kind {
		case htmlparser.StartTagToken:
			tag := tokens.Token()
			if tag.Data != "a" {
				continue
			}
			for _, attr := range tag.Attr {
				if attr.Key == "href" && attr.Val == unsubscribeURL {
					return content
				}
			}
		case htmlparser.EndTagToken:
			switch tokens.Token().Data {
			case "body":
				bodyEnd = start
			case "html":
				documentEnd = start
			}
		}
	}
	insertAt := len(content)
	if bodyEnd >= 0 {
		insertAt = bodyEnd
	} else if documentEnd >= 0 {
		insertAt = documentEnd
	}
	label := "退订营销邮件"
	if common.IsEmailLanguageEnglish() {
		label = "Unsubscribe from marketing emails"
	}
	footer := `<p style="max-width:560px;margin:20px auto;padding:0 16px;text-align:center;font-size:13px;line-height:1.5"><a style="color:#1d4ed8;text-decoration:underline" href="` +
		html.EscapeString(unsubscribeURL) + `">` + label + `</a></p>`
	return content[:insertAt] + footer + content[insertAt:]
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
