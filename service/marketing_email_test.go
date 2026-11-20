package service

import (
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarketingEmailHonorsOptOutAndIncludesUnsubscribeLink(t *testing.T) {
	originalSecret := common.CryptoSecret
	originalServerAddress := system_setting.ServerAddress
	common.CryptoSecret = "marketing-test-secret"
	system_setting.ServerAddress = "https://platform.example"
	t.Cleanup(func() {
		common.CryptoSecret = originalSecret
		system_setting.ServerAddress = originalServerAddress
	})

	var sentTo, sentContent string
	send := func(_, to, content string) error {
		sentTo = to
		sentContent = content
		return nil
	}
	user := &model.User{
		Id: 42, Username: "marketing-user", Email: "user@example.com", Status: common.UserStatusEnabled,
	}
	require.NoError(t, SendMarketingEmailToUser(user, "{{system_name}} update", "<p>Hello {{username}}</p>", send))
	assert.Equal(t, "user@example.com", sentTo)
	assert.Contains(t, sentContent, "marketing-user")
	assert.Contains(t, sentContent, "https://platform.example/api/marketing_email/unsubscribe?user_id=42&token=")
	assert.True(t, VerifyMarketingEmailUnsubscribeToken(42, strings.Split(strings.Split(sentContent, "token=")[1], "\"")[0]))
	assert.False(t, VerifyMarketingEmailUnsubscribeToken(43, strings.Split(strings.Split(sentContent, "token=")[1], "\"")[0]))

	enabled := false
	user.SetSetting(dto.UserSetting{MarketingEmailEnabled: &enabled})
	sentTo, sentContent = "", ""
	err := SendMarketingEmailToUser(user, "subject", "body", send)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "preference")
	assert.Empty(t, sentTo)
	assert.Empty(t, sentContent)
}

