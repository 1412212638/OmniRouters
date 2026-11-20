package service

import (
	"html"
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
	assert.Contains(t, html.UnescapeString(sentContent), "https://platform.example/api/marketing_email/unsubscribe?user_id=42&token=")
	_, tokenAndRest, found := strings.Cut(sentContent, "token=")
	require.True(t, found)
	token, _, _ := strings.Cut(tokenAndRest, "\"")
	assert.True(t, VerifyMarketingEmailUnsubscribeToken(42, token))
	assert.False(t, VerifyMarketingEmailUnsubscribeToken(43, token))

	enabled := false
	user.SetSetting(dto.UserSetting{MarketingEmailEnabled: &enabled})
	sentTo, sentContent = "", ""
	err := SendMarketingEmailToUser(user, "subject", "body", send)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "preference")
	assert.Empty(t, sentTo)
	assert.Empty(t, sentContent)
}
