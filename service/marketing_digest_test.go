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

func TestMarketingDigestEmailHonorsOptOutAndIncludesUnsubscribeLink(t *testing.T) {
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
		Id: 42, Username: "newsletter-user", Email: "user@example.com", Status: common.UserStatusEnabled,
	}
	require.NoError(t, SendMarketingDigestEmail(user, "{{system_name}} update", "<p>Hello {{username}}</p>", send))
	assert.Equal(t, "user@example.com", sentTo)
	assert.Contains(t, sentContent, "newsletter-user")
	assert.Contains(t, sentContent, "https://platform.example/api/marketing_email/unsubscribe?user_id=42&token=")
	assert.True(t, VerifyMarketingDigestUnsubscribeToken(42, strings.Split(strings.Split(sentContent, "token=")[1], "\"")[0]))
	assert.False(t, VerifyMarketingDigestUnsubscribeToken(43, strings.Split(strings.Split(sentContent, "token=")[1], "\"")[0]))

	enabled := false
	user.SetSetting(dto.UserSetting{MarketingEmailEnabled: &enabled})
	sentTo, sentContent = "", ""
	err := SendMarketingDigestEmail(user, "subject", "body", send)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "preference")
	assert.Empty(t, sentTo)
	assert.Empty(t, sentContent)
}

