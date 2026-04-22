package common

import "fmt"

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

func BuildEmailVerificationMail(systemName string, code string, validMinutes int) (subject string, content string) {
	if IsEmailLanguageEnglish() {
		subject = fmt.Sprintf("%s Email Verification", systemName)
		content = fmt.Sprintf(
			"<p>Hello,</p><p>You are verifying the email address for %s.</p><p>Your verification code is: <strong>%s</strong></p><p>This code is valid for %s. If this was not you, please ignore this email.</p>",
			systemName,
			code,
			formatEmailMinutes(validMinutes),
		)
		return
	}

	subject = fmt.Sprintf("%s邮箱验证邮件", systemName)
	content = fmt.Sprintf(
		"<p>您好，您正在进行 %s 邮箱验证。</p><p>您的验证码为：<strong>%s</strong></p><p>验证码在 %s 内有效，如果不是本人操作，请忽略此邮件。</p>",
		systemName,
		code,
		formatEmailMinutes(validMinutes),
	)
	return
}

func BuildPasswordResetMail(systemName string, link string, validMinutes int) (subject string, content string) {
	if IsEmailLanguageEnglish() {
		subject = fmt.Sprintf("%s Password Reset", systemName)
		content = fmt.Sprintf(
			"<p>Hello,</p><p>You requested a password reset for %s.</p><p>Click <a href='%s'>here</a> to reset your password.</p><p>If the link does not open, copy and paste this URL into your browser:<br>%s</p><p>This reset link is valid for %s. If this was not you, please ignore this email.</p>",
			systemName,
			link,
			link,
			formatEmailMinutes(validMinutes),
		)
		return
	}

	subject = fmt.Sprintf("%s密码重置", systemName)
	content = fmt.Sprintf(
		"<p>您好，您正在进行 %s 密码重置。</p><p>点击 <a href='%s'>此处</a> 进行密码重置。</p><p>如果链接无法点击，请尝试将下面的链接复制到浏览器中打开：<br>%s</p><p>重置链接在 %s 内有效，如果不是本人操作，请忽略此邮件。</p>",
		systemName,
		link,
		link,
		formatEmailMinutes(validMinutes),
	)
	return
}

func BuildQuotaWarningMail(remainingQuota string, topUpLink string, subscription bool) (subject string, content string) {
	if IsEmailLanguageEnglish() {
		if subscription {
			subject = "Your subscription quota is running low"
		} else {
			subject = "Your quota is running low"
		}
		content = fmt.Sprintf(
			"<p>Hello,</p><p>%s</p><p>Your remaining quota is <strong>%s</strong>.</p><p>To avoid interruption, please top up in time:</p><p><a href='%s'>%s</a></p>",
			subject,
			remainingQuota,
			topUpLink,
			topUpLink,
		)
		return
	}

	if subscription {
		subject = "您的订阅额度即将用尽"
	} else {
		subject = "您的额度即将用尽"
	}
	content = fmt.Sprintf(
		"<p>您好，</p><p>%s</p><p>当前剩余额度为 <strong>%s</strong>。</p><p>为了不影响您的使用，请及时充值：</p><p><a href='%s'>%s</a></p>",
		subject,
		remainingQuota,
		topUpLink,
		topUpLink,
	)
	return
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
