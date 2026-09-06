package oauth

import (
	"errors"
	"net/url"
	"strings"

	"github.com/QuantumNous/new-api/setting/system_setting"
	"golang.org/x/oauth2"
)

const TelegramOAuthIssuer = "https://oauth.telegram.org"

var ErrTelegramOAuthNotReady = errors.New("Telegram OAuth is not ready")

// TelegramOAuthFlow keeps the PKCE verifier server-side; only the derived
// authorization URL should be exposed to the browser.
type TelegramOAuthFlow struct {
	CodeVerifier string `json:"code_verifier"`
	ClientID     string `json:"client_id"`
	RedirectURI  string `json:"redirect_uri"`
}

func NewTelegramOAuthFlow() (*TelegramOAuthFlow, error) {
	settings := system_setting.GetTelegramSettings()
	if !settings.IsConfigured() { return nil, ErrTelegramOAuthNotReady }
	redirect := strings.TrimRight(system_setting.ServerAddress, "/") + "/oauth/telegram"
	if parsed, err := url.Parse(redirect); err != nil || parsed.Host == "" { return nil, ErrTelegramOAuthNotReady }
	return &TelegramOAuthFlow{CodeVerifier: oauth2.GenerateVerifier(), ClientID: strings.TrimSpace(settings.ClientID), RedirectURI: redirect}, nil
}

func (flow *TelegramOAuthFlow) AuthorizationURL(state string) string {
	values := url.Values{
		"client_id": {flow.ClientID}, "redirect_uri": {flow.RedirectURI},
		"response_type": {"code"}, "scope": {"openid profile"}, "state": {state},
		"code_challenge": {oauth2.S256ChallengeFromVerifier(flow.CodeVerifier)},
		"code_challenge_method": {"S256"},
	}
	return TelegramOAuthIssuer + "/auth?" + values.Encode()
}
