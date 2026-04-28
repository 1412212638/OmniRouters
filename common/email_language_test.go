package common

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildTopUpSuccessMailUsesCustomTemplate(t *testing.T) {
	originalLanguage := EmailLanguage
	originalSystemName := SystemName
	originalSubjectTemplate := TopUpSuccessSubjectTemplate
	originalContentTemplate := TopUpSuccessContentTemplate
	defer func() {
		EmailLanguage = originalLanguage
		SystemName = originalSystemName
		TopUpSuccessSubjectTemplate = originalSubjectTemplate
		TopUpSuccessContentTemplate = originalContentTemplate
	}()

	EmailLanguage = EmailLanguageChinese
	SystemName = "OmniRouters"
	TopUpSuccessSubjectTemplate = "{{system_name}} order {{trade_no}} paid"
	TopUpSuccessContentTemplate = "Hi {{username}}, {{amount}} => {{quota}}, balance {{balance}}"

	subject, content := BuildTopUpSuccessMail(TopUpSuccessMailParams{
		Username: "Alex",
		TradeNo:  "topup_001",
		Amount:   "10.00",
		Quota:    "$10.00",
		Balance:  "$18.88",
	})

	require.Equal(t, "OmniRouters order topup_001 paid", subject)
	require.Equal(t, "Hi Alex, 10.00 => $10.00, balance $18.88", content)
}

func TestBuildTopUpSuccessMailChineseDefaultUsername(t *testing.T) {
	originalLanguage := EmailLanguage
	originalSystemName := SystemName
	originalSubjectTemplate := TopUpSuccessSubjectTemplate
	originalContentTemplate := TopUpSuccessContentTemplate
	defer func() {
		EmailLanguage = originalLanguage
		SystemName = originalSystemName
		TopUpSuccessSubjectTemplate = originalSubjectTemplate
		TopUpSuccessContentTemplate = originalContentTemplate
	}()

	EmailLanguage = EmailLanguageChinese
	SystemName = "OmniRouters"
	TopUpSuccessSubjectTemplate = ""
	TopUpSuccessContentTemplate = ""

	subject, content := BuildTopUpSuccessMail(TopUpSuccessMailParams{
		Username:      "张三",
		TradeNo:       "topup_001",
		PaymentMethod: "stripe",
		Amount:        "10.00",
		Quota:         "10.00 美元",
		Balance:       "18.88 美元",
		PaidAt:        "2026-04-27 20:30:00",
	})

	require.Equal(t, "OmniRouters充值成功通知", subject)
	require.Contains(t, content, "您好张三，")
	require.Contains(t, content, "topup_001")
	require.Contains(t, content, "18.88 美元")
}

func TestBuildMarketingMailUsesRequestTemplate(t *testing.T) {
	originalLanguage := EmailLanguage
	originalSystemName := SystemName
	originalSubjectTemplate := MarketingEmailSubjectTemplate
	originalContentTemplate := MarketingEmailContentTemplate
	defer func() {
		EmailLanguage = originalLanguage
		SystemName = originalSystemName
		MarketingEmailSubjectTemplate = originalSubjectTemplate
		MarketingEmailContentTemplate = originalContentTemplate
	}()

	EmailLanguage = EmailLanguageEnglish
	SystemName = "OmniRouters"
	MarketingEmailSubjectTemplate = "global {{username}}"
	MarketingEmailContentTemplate = "global content"

	subject, content := BuildMarketingMail(MarketingMailParams{
		Username:        "alex",
		DisplayName:     "Alex",
		Email:           "alex@example.com",
		UserId:          "42",
		SubjectTemplate: "{{system_name}} hello {{display_name}}",
		ContentTemplate: "id={{user_id}}, email={{email}}, username={{username}}",
	})

	require.Equal(t, "OmniRouters hello Alex", subject)
	require.Equal(t, "id=42, email=alex@example.com, username=alex", content)
}
