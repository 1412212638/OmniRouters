package common

import (
	"fmt"
	"strings"
)

const (
	EmailLanguageChinese = "zh"
	EmailLanguageEnglish = "en"
)

func NormalizeEmailLanguage(value string) string {
	switch value {
	case EmailLanguageEnglish, "en-us", "en-US", "en-gb", "en-GB":
		return EmailLanguageEnglish
	default:
		return EmailLanguageChinese
	}
}

func IsEmailLanguageEnglish() bool {
	return NormalizeEmailLanguage(EmailLanguage) == EmailLanguageEnglish
}

func renderEmailTemplate(template string, values map[string]string) string {
	rendered := template
	for key, value := range values {
		rendered = strings.ReplaceAll(rendered, "{{"+key+"}}", value)
	}
	return rendered
}

func BuildEmailVerificationMail(systemName string, code string, validMinutes int) (subject string, content string) {
	values := map[string]string{
		"system_name":   systemName,
		"code":          code,
		"valid_minutes": fmt.Sprintf("%d", validMinutes),
		"valid_time":    formatEmailMinutes(validMinutes),
	}
	if strings.TrimSpace(EmailVerificationSubjectTemplate) != "" {
		subject = renderEmailTemplate(EmailVerificationSubjectTemplate, values)
	}
	if strings.TrimSpace(EmailVerificationContentTemplate) != "" {
		content = renderEmailTemplate(EmailVerificationContentTemplate, values)
	}
	if subject != "" && content != "" {
		return
	}

	if IsEmailLanguageEnglish() {
		if subject == "" {
			subject = fmt.Sprintf("%s Email Verification", systemName)
		}
		if content == "" {
			content = fmt.Sprintf(
				"<p>Hello,</p><p>You are verifying the email address for %s.</p><p>Your verification code is: <strong>%s</strong></p><p>This code is valid for %s. If this was not you, please ignore this email.</p>",
				systemName,
				code,
				formatEmailMinutes(validMinutes),
			)
		}
		return
	}

	if subject == "" {
		subject = fmt.Sprintf("%s邮箱验证邮件", systemName)
	}
	if content == "" {
		content = fmt.Sprintf(
			"<p>您好，您正在进行 %s 邮箱验证。</p><p>您的验证码为：<strong>%s</strong></p><p>验证码在 %s 内有效，如果不是本人操作，请忽略此邮件。</p>",
			systemName,
			code,
			formatEmailMinutes(validMinutes),
		)
	}
	return
}

func BuildPasswordResetMail(systemName string, link string, validMinutes int) (subject string, content string) {
	values := map[string]string{
		"system_name":   systemName,
		"link":          link,
		"valid_minutes": fmt.Sprintf("%d", validMinutes),
		"valid_time":    formatEmailMinutes(validMinutes),
	}
	if strings.TrimSpace(PasswordResetSubjectTemplate) != "" {
		subject = renderEmailTemplate(PasswordResetSubjectTemplate, values)
	}
	if strings.TrimSpace(PasswordResetContentTemplate) != "" {
		content = renderEmailTemplate(PasswordResetContentTemplate, values)
	}
	if subject != "" && content != "" {
		return
	}

	if IsEmailLanguageEnglish() {
		if subject == "" {
			subject = fmt.Sprintf("%s Password Reset", systemName)
		}
		if content == "" {
			content = fmt.Sprintf(
				"<p>Hello,</p><p>You requested a password reset for %s.</p><p>Click <a href='%s'>here</a> to reset your password.</p><p>If the link does not open, copy and paste this URL into your browser:<br>%s</p><p>This reset link is valid for %s. If this was not you, please ignore this email.</p>",
				systemName,
				link,
				link,
				formatEmailMinutes(validMinutes),
			)
		}
		return
	}

	if subject == "" {
		subject = fmt.Sprintf("%s密码重置", systemName)
	}
	if content == "" {
		content = fmt.Sprintf(
			"<p>您好，您正在进行 %s 密码重置。</p><p>点击 <a href='%s'>此处</a> 进行密码重置。</p><p>如果链接无法点击，请尝试将下面的链接复制到浏览器中打开：<br>%s</p><p>重置链接在 %s 内有效，如果不是本人操作，请忽略此邮件。</p>",
			systemName,
			link,
			link,
			formatEmailMinutes(validMinutes),
		)
	}
	return
}

func BuildQuotaWarningMail(remainingQuota string, topUpLink string, subscription bool) (subject string, content string) {
	subjectTemplate := QuotaWarningSubjectTemplate
	contentTemplate := QuotaWarningContentTemplate
	if subscription {
		subjectTemplate = SubscriptionQuotaWarningSubjectTemplate
		contentTemplate = SubscriptionQuotaWarningContentTemplate
	}
	values := map[string]string{
		"system_name":     SystemName,
		"remaining_quota": remainingQuota,
		"top_up_link":     topUpLink,
	}
	if strings.TrimSpace(subjectTemplate) != "" {
		subject = renderEmailTemplate(subjectTemplate, values)
	}
	if strings.TrimSpace(contentTemplate) != "" {
		content = renderEmailTemplate(contentTemplate, values)
	}
	if subject != "" && content != "" {
		return
	}

	if IsEmailLanguageEnglish() {
		if subscription {
			if subject == "" {
				subject = "Your subscription quota is running low"
			}
		} else if subject == "" {
			subject = "Your quota is running low"
		}
		if content == "" {
			content = fmt.Sprintf(
				"<p>Hello,</p><p>%s</p><p>Your remaining quota is <strong>%s</strong>.</p><p>To avoid interruption, please top up in time:</p><p><a href='%s'>%s</a></p>",
				subject,
				remainingQuota,
				topUpLink,
				topUpLink,
			)
		}
		return
	}

	if subscription {
		if subject == "" {
			subject = "您的订阅额度即将用尽"
		}
	} else if subject == "" {
		subject = "您的额度即将用尽"
	}
	if content == "" {
		content = fmt.Sprintf(
			"<p>您好，</p><p>%s</p><p>当前剩余额度为 <strong>%s</strong>。</p><p>为了不影响您的使用，请及时充值：</p><p><a href='%s'>%s</a></p>",
			subject,
			remainingQuota,
			topUpLink,
			topUpLink,
		)
	}
	return
}

type TopUpSuccessMailParams struct {
	SystemName      string
	Username        string
	TradeNo         string
	PaymentMethod   string
	PaymentProvider string
	Amount          string
	Quota           string
	Balance         string
	PaidAt          string
}

func BuildTopUpSuccessMail(params TopUpSuccessMailParams) (subject string, content string) {
	systemName := params.SystemName
	if systemName == "" {
		systemName = SystemName
	}
	values := map[string]string{
		"system_name":      systemName,
		"username":         params.Username,
		"trade_no":         params.TradeNo,
		"payment_method":   params.PaymentMethod,
		"payment_provider": params.PaymentProvider,
		"amount":           params.Amount,
		"quota":            params.Quota,
		"balance":          params.Balance,
		"paid_at":          params.PaidAt,
	}
	if strings.TrimSpace(TopUpSuccessSubjectTemplate) != "" {
		subject = renderEmailTemplate(TopUpSuccessSubjectTemplate, values)
	}
	if strings.TrimSpace(TopUpSuccessContentTemplate) != "" {
		content = renderEmailTemplate(TopUpSuccessContentTemplate, values)
	}
	if subject != "" && content != "" {
		return
	}

	if IsEmailLanguageEnglish() {
		if subject == "" {
			subject = fmt.Sprintf("%s Top-up Successful", systemName)
		}
		if content == "" {
			content = fmt.Sprintf(
				"<p>Hello%s,</p><p>Your top-up has been completed.</p><p>Amount: <strong>%s</strong></p><p>Quota added: <strong>%s</strong></p><p>Current balance: <strong>%s</strong></p><p>Payment method: %s</p><p>Trade no: %s</p><p>Paid at: %s</p>",
				formatEmailUsername(params.Username),
				params.Amount,
				params.Quota,
				params.Balance,
				params.PaymentMethod,
				params.TradeNo,
				params.PaidAt,
			)
		}
		return
	}

	if subject == "" {
		subject = fmt.Sprintf("%s充值成功通知", systemName)
	}
	if content == "" {
		content = fmt.Sprintf(
			"<p>您好%s，</p><p>您的充值已经到账。</p><p>支付金额：<strong>%s</strong></p><p>到账额度：<strong>%s</strong></p><p>当前余额：<strong>%s</strong></p><p>支付方式：%s</p><p>订单号：%s</p><p>到账时间：%s</p>",
			formatEmailUsername(params.Username),
			params.Amount,
			params.Quota,
			params.Balance,
			params.PaymentMethod,
			params.TradeNo,
			params.PaidAt,
		)
	}
	return
}

type MarketingMailParams struct {
	SystemName      string
	Username        string
	DisplayName     string
	Email           string
	UserId          string
	SubjectTemplate string
	ContentTemplate string
}

func BuildMarketingMail(params MarketingMailParams) (subject string, content string) {
	systemName := params.SystemName
	if systemName == "" {
		systemName = SystemName
	}
	displayName := params.DisplayName
	if strings.TrimSpace(displayName) == "" {
		displayName = params.Username
	}
	values := map[string]string{
		"system_name":  systemName,
		"username":     params.Username,
		"display_name": displayName,
		"email":        params.Email,
		"user_id":      params.UserId,
	}

	subjectTemplate := params.SubjectTemplate
	if strings.TrimSpace(subjectTemplate) == "" {
		subjectTemplate = MarketingEmailSubjectTemplate
	}
	contentTemplate := params.ContentTemplate
	if strings.TrimSpace(contentTemplate) == "" {
		contentTemplate = MarketingEmailContentTemplate
	}

	if strings.TrimSpace(subjectTemplate) != "" {
		subject = renderEmailTemplate(subjectTemplate, values)
	}
	if strings.TrimSpace(contentTemplate) != "" {
		content = renderEmailTemplate(contentTemplate, values)
	}
	if subject != "" && content != "" {
		return
	}

	if IsEmailLanguageEnglish() {
		if subject == "" {
			subject = fmt.Sprintf("%s Update", systemName)
		}
		if content == "" {
			content = fmt.Sprintf("<p>Hello%s,</p><p>Here is the latest update from %s.</p>", formatEmailUsername(displayName), systemName)
		}
		return
	}

	if subject == "" {
		subject = fmt.Sprintf("%s最新通知", systemName)
	}
	if content == "" {
		content = fmt.Sprintf("<p>您好%s，</p><p>这里是 %s 的最新通知。</p>", formatEmailUsername(displayName), systemName)
	}
	return
}

func formatEmailUsername(username string) string {
	if strings.TrimSpace(username) == "" {
		return ""
	}
	if IsEmailLanguageEnglish() {
		return " " + username
	}
	return username
}

func formatEmailMinutes(validMinutes int) string {
	if IsEmailLanguageEnglish() {
		if validMinutes == 1 {
			return "1 minute"
		}
		return fmt.Sprintf("%d minutes", validMinutes)
	}
	return fmt.Sprintf("%d 分钟", validMinutes)
}
