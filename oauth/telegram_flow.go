package oauth

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/coreos/go-oidc/v3/oidc"
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

func ExchangeTelegramCode(ctx context.Context, client *http.Client, flow *TelegramOAuthFlow, code string) (*OAuthToken, error) {
	settings := system_setting.GetTelegramSettings()
	if flow == nil || !settings.IsConfigured() || strings.TrimSpace(code) == "" ||
		flow.ClientID != strings.TrimSpace(settings.ClientID) || flow.CodeVerifier == "" {
		return nil, ErrTelegramOAuthNotReady
	}
	if client == nil { client = &http.Client{Timeout: 20 * time.Second} }
	values := url.Values{"grant_type": {"authorization_code"}, "code": {code}, "client_id": {flow.ClientID}, "redirect_uri": {flow.RedirectURI}, "code_verifier": {flow.CodeVerifier}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, TelegramOAuthIssuer+"/token", strings.NewReader(values.Encode()))
	if err != nil { return nil, fmt.Errorf("telegram token request: %w", err) }
	req.SetBasicAuth(flow.ClientID, strings.TrimSpace(settings.ClientSecret)); req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req); if err != nil { return nil, fmt.Errorf("telegram token request: %w", err) }
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK { return nil, fmt.Errorf("telegram token endpoint returned %d", resp.StatusCode) }
	var token OAuthToken
	if err := common.DecodeJson(io.LimitReader(resp.Body, 1<<20), &token); err != nil { return nil, fmt.Errorf("telegram token response: %w", err) }
	if token.IDToken == "" { return nil, errors.New("telegram token response has no id token") }
	token.ClientID = flow.ClientID
	return &token, nil
}

// VerifyTelegramIDToken validates the provider-issued JWT before any identity
// claim is used. The verifier fetches and caches Telegram's JWKS keys.
func VerifyTelegramIDToken(ctx context.Context, token *OAuthToken, keys oidc.KeySet) (*OAuthUser, error) {
	settings := system_setting.GetTelegramSettings()
	if token == nil || token.IDToken == "" || token.ClientID != strings.TrimSpace(settings.ClientID) || keys == nil {
		return nil, ErrTelegramOAuthNotReady
	}
	verifier := oidc.NewVerifier(TelegramOAuthIssuer, keys, &oidc.Config{ClientID: token.ClientID, SupportedSigningAlgs: []string{oidc.RS256, oidc.ES256}})
	verified, err := verifier.Verify(ctx, token.IDToken)
	if err != nil { return nil, fmt.Errorf("telegram id token verification: %w", err) }
	var claims struct { ID string `json:"id"`; Username string `json:"preferred_username"`; Name string `json:"name"` }
	if err := verified.Claims(&claims); err != nil || strings.TrimSpace(claims.ID) == "" || verified.Subject == "" { return nil, errors.New("invalid telegram identity claims") }
	return &OAuthUser{ProviderUserID: strings.TrimSpace(claims.ID), Username: claims.Username, DisplayName: claims.Name}, nil
}
