package common

import (
	"bytes"
	"io"
	"mime/quotedprintable"
	"net/mail"
	"net/textproto"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMarketingSMTPMessageUsesIndependentSender(t *testing.T) {
	cfg := MarketingSMTPConfig{
		Mode: "smtp", Server: "smtp.example.com", Port: 465,
		Account: "account@example.com", Password: "secret",
		From: "campaign@example.com", DisplayName: "Campaigns", ReplyTo: "support@example.org",
		PerSecond: 5,
	}
	require.NoError(t, cfg.Validate())
	content := "<p>" + strings.Repeat("测试", 1200) + "</p>"
	url := "https://example.com/api/marketing_email/unsubscribe?user_id=7&token=abc"
	var raw bytes.Buffer
	require.NoError(t, cfg.writeMessage(&raw, "产品动态", "reader@example.org", content, url, "message@example.com", "proof"))
	message, err := mail.ReadMessage(&raw)
	require.NoError(t, err)
	require.Equal(t, `"Campaigns" <campaign@example.com>`, message.Header.Get("From"))
	require.Equal(t, "<support@example.org>", message.Header.Get("Reply-To"))
	require.Equal(t, "<"+url+">", message.Header.Get("List-Unsubscribe"))
	require.Equal(t, "List-Unsubscribe=One-Click", message.Header.Get("List-Unsubscribe-Post"))
	require.Equal(t, "proof", message.Header.Get("X-Tencentcloudses-Cb-Proof"))
	decoded, err := io.ReadAll(quotedprintable.NewReader(message.Body))
	require.NoError(t, err)
	require.Equal(t, content, string(decoded))
}

func TestMarketingSMTPClassifiesSESLimitAndUnconfirmedSubmission(t *testing.T) {
	classified := func(err error, stage string) *MarketingSMTPError {
		t.Helper()
		var result *MarketingSMTPError
		require.ErrorAs(t, marketingSMTPError(err, stage), &result)
		return result
	}
	require.True(t, classified(&textproto.Error{Code: 554, Msg: "Failed: RequestLimitExceeded.SmtpRateLimit"}, "submit").Retryable)
	require.False(t, classified(&textproto.Error{Code: 550, Msg: "No mailbox"}, "recipient").Retryable)
	require.True(t, classified(&textproto.Error{Code: 535, Msg: "Invalid login"}, "login").Fatal)
	require.True(t, classified(io.EOF, "submit").Unknown)
}
