package service

import (
	"context"
	"html"
	"strings"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
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

func TestMarketingUnsubscribeFooterStaysInsideBody(t *testing.T) {
	const unsubscribeURL = "https://platform.example/api/marketing_email/unsubscribe?user_id=42&token=abc"
	const body = `<!doctype html><HTML><BODY><p>News</p><!-- ` + unsubscribeURL + ` --><script>const text = '</body>'</script></BODY></HTML>`
	rendered := addMarketingUnsubscribeFooter(body, unsubscribeURL)
	assert.Contains(t, rendered, `href="`+html.EscapeString(unsubscribeURL)+`"`)
	assert.True(t, strings.HasSuffix(rendered, "</a></p></BODY></HTML>"))

	custom := `<p>News</p><a href="` + html.EscapeString(unsubscribeURL) + `">Manage subscriptions</a>`
	assert.Equal(t, custom, addMarketingUnsubscribeFooter(custom, unsubscribeURL))
}

func TestMarketingCustomAddressRespectsOptOutAndSuppression(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	previous := model.DB
	model.DB = db
	t.Cleanup(func() { model.DB = previous })
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.MarketingEmailSuppression{}))

	email := "customer@example.net"
	notification := "notify_all@example.net"
	enabled := false
	user := model.User{Username: "opted-out", Email: email, Status: common.UserStatusEnabled}
	user.SetSetting(dto.UserSetting{MarketingEmailEnabled: &enabled, NotificationEmail: " " + strings.ToUpper(notification) + " "})
	require.NoError(t, db.Create(&user).Error)

	addresses, err := normalizeCustomMarketingEmails([]string{" external@example.net ", "EXTERNAL@example.net"})
	require.NoError(t, err)
	require.Equal(t, []string{"external@example.net"}, addresses)
	_, err = normalizeCustomMarketingEmails([]string{"external@example.net", "invalid"})
	require.Error(t, err)
	recipient := marketingEmailRecipient{address: addresses[0]}
	account, address, unsubscribe, err := recipient.resolve()
	require.NoError(t, err)
	require.Nil(t, account)
	require.Equal(t, addresses[0], address)
	require.Contains(t, unsubscribe, "email_hash=")
	require.NotContains(t, unsubscribe, address)
	_, content := buildMarketingRecipientMail(account, address, unsubscribe, "Test", "<p>Hello {{email}}</p>")
	require.Contains(t, content, html.EscapeString(unsubscribe))
	require.NoError(t, model.SuppressMarketingEmail(address, "unsubscribe"))

	result := &MarketingEmailResult{}
	var nextSend time.Time
	for _, blocked := range []string{strings.ToUpper(email), notification, address} {
		require.NoError(t, deliverCampaignEmail(context.Background(), common.MarketingSMTPConfig{}, time.Second, &nextSend, marketingEmailRecipient{address: blocked}, MarketingEmailSendPayload{}, result))
	}
	require.Equal(t, 3, result.Skipped)
	require.True(t, nextSend.IsZero())
}
