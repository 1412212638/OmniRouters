package controller

import (
	"net/http"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"

	"github.com/gin-gonic/gin"
)

var adminEmailSettingOptionKeys = []string{
	"EmailDomainRestrictionEnabled",
	"EmailAliasRestrictionEnabled",
	"EmailDomainWhitelist",
	"SMTPServer",
	"SMTPFrom",
	"SMTPPort",
	"SMTPAccount",
	"SMTPToken",
	"SMTPSSLEnabled",
	"SMTPForceAuthLogin",
	"EmailLanguage",
	"TopUpSuccessEmailEnabled",
	"EmailVerificationSubjectTemplate",
	"EmailVerificationContentTemplate",
	"PasswordResetSubjectTemplate",
	"PasswordResetContentTemplate",
	"QuotaWarningSubjectTemplate",
	"QuotaWarningContentTemplate",
	"SubscriptionQuotaWarningSubjectTemplate",
	"SubscriptionQuotaWarningContentTemplate",
	"TopUpSuccessSubjectTemplate",
	"TopUpSuccessContentTemplate",
	"MarketingEmailSubjectTemplate",
	"MarketingEmailContentTemplate",
}

var adminEmailSettingOptionKeySet = buildOptionKeySet(adminEmailSettingOptionKeys)

func buildOptionKeySet(keys []string) map[string]struct{} {
	set := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		set[key] = struct{}{}
	}
	return set
}

func isAdminEmailSettingOptionKey(key string) bool {
	_, ok := adminEmailSettingOptionKeySet[key]
	return ok
}

func GetEmailSettingsOptions(c *gin.Context) {
	options := make([]*model.Option, 0, len(adminEmailSettingOptionKeys))

	common.OptionMapRWMutex.RLock()
	for _, key := range adminEmailSettingOptionKeys {
		if isSensitiveOptionKey(key) && !isVisiblePublicKeyOption(key) {
			continue
		}
		value, ok := common.OptionMap[key]
		if !ok {
			continue
		}
		options = append(options, &model.Option{
			Key:   key,
			Value: common.Interface2String(value),
		})
	}
	common.OptionMapRWMutex.RUnlock()

	common.ApiSuccess(c, options)
}

func UpdateEmailSettingOption(c *gin.Context) {
	var option OptionUpdateRequest
	if err := common.DecodeJson(c.Request.Body, &option); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "invalid request parameters",
		})
		return
	}

	if !isAdminEmailSettingOptionKey(option.Key) {
		common.ApiErrorMsg(c, "unsupported email setting")
		return
	}

	option.Value = normalizeOptionUpdateValue(option.Value)
	if option.Key == "EmailDomainRestrictionEnabled" && option.Value == "true" && len(common.EmailDomainWhitelist) == 0 {
		common.ApiErrorMsg(c, "cannot enable email domain restriction before adding allowed domains")
		return
	}

	if err := model.UpdateOption(option.Key, option.Value.(string)); err != nil {
		common.ApiError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
}
