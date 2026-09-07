package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
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
	redirect, err := telegramRedirectURI()
	if err != nil { return nil, err }
	return &TelegramOAuthFlow{CodeVerifier: oauth2.GenerateVerifier(), ClientID: strings.TrimSpace(settings.ClientID), RedirectURI: redirect}, nil
}

func telegramRedirectURI() (string, error) {
	redirect := strings.TrimRight(system_setting.ServerAddress, "/") + "/oauth/telegram"
	if parsed, err := url.Parse(redirect); err != nil || parsed.Host == "" || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" { return "", ErrTelegramOAuthNotReady }
	return redirect, nil
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
		flow.ClientID != strings.TrimSpace(settings.ClientID) || flow.CodeVerifier == "" ||
		flow.RedirectURI != strings.TrimRight(system_setting.ServerAddress, "/")+"/oauth/telegram" {
		return nil, ErrTelegramOAuthNotReady
	}
	if _, err := telegramRedirectURI(); err != nil { return nil, err }
	if client == nil { client = &http.Client{Timeout: 20 * time.Second} }
	values := url.Values{"grant_type": {"authorization_code"}, "code": {code}, "client_id": {flow.ClientID}, "redirect_uri": {flow.RedirectURI}, "code_verifier": {flow.CodeVerifier}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, TelegramOAuthIssuer+"/token", strings.NewReader(values.Encode()))
	if err != nil { return nil, fmt.Errorf("telegram token request: %w", err) }
	req.SetBasicAuth(flow.ClientID, strings.TrimSpace(settings.ClientSecret)); req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req); if err != nil { return nil, fmt.Errorf("telegram token request: %w", err) }
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK { return nil, fmt.Errorf("telegram token endpoint returned %d", resp.StatusCode) }
	var token OAuthToken
	body, err := io.ReadAll(io.LimitReader(resp.Body, (1<<20)+1))
	if err != nil || len(body) > 1<<20 { return nil, errors.New("invalid telegram token response size") }
	if err := common.Unmarshal(body, &token); err != nil { return nil, fmt.Errorf("telegram token response: %w", err) }
	if token.IDToken == "" { return nil, errors.New("telegram token response has no id token") }
	token.ClientID = flow.ClientID
	return &token, nil
}

// VerifyTelegramIDToken validates the provider-issued JWT before any identity
// claim is used. The verifier fetches and caches Telegram's JWKS keys.
func VerifyTelegramIDToken(ctx context.Context, token *OAuthToken, keys oidc.KeySet) (*OAuthUser, error) {
	settings := system_setting.GetTelegramSettings()
	if !settings.IsConfigured() || token == nil || token.IDToken == "" || token.ClientID != strings.TrimSpace(settings.ClientID) || keys == nil {
		return nil, ErrTelegramOAuthNotReady
	}
	verifier := oidc.NewVerifier(TelegramOAuthIssuer, keys, &oidc.Config{ClientID: token.ClientID, SupportedSigningAlgs: []string{oidc.RS256, oidc.ES256}})
	verified, err := verifier.Verify(ctx, token.IDToken)
	if err != nil { return nil, fmt.Errorf("telegram id token verification: %w", err) }
	var raw json.RawMessage
	if err := verified.Claims(&raw); err != nil { return nil, errors.New("invalid telegram identity claims") }
	return parseTelegramIdentity(raw, verified.Subject)
}

func parseTelegramIdentity(raw []byte, subject string) (*OAuthUser, error) {
	var claims struct { ID json.Number `json:"id"`; Username string `json:"preferred_username"`; Name string `json:"name"` }
	if err := common.Unmarshal(raw, &claims); err != nil { return nil, errors.New("invalid telegram identity claims") }
	id, err := strconv.ParseUint(claims.ID.String(), 10, 64)
	if err != nil || id == 0 || strings.TrimSpace(subject) == "" { return nil, errors.New("invalid telegram identity claims") }
	return &OAuthUser{ProviderUserID: strconv.FormatUint(id, 10), Username: claims.Username, DisplayName: claims.Name}, nil
}

const TelegramOAuthFlowContextKey = "telegram_oauth_flow"

type TelegramOAuthProvider struct {
	client *http.Client
	keys oidc.KeySet
}

var _ Provider = (*TelegramOAuthProvider)(nil)

func NewTelegramOAuthProvider(client *http.Client) *TelegramOAuthProvider {
	if client == nil { client = &http.Client{Timeout: 20 * time.Second} }
	return &TelegramOAuthProvider{client: client, keys: oidc.NewRemoteKeySet(oidc.ClientContext(context.Background(), client), TelegramOAuthIssuer+"/.well-known/jwks.json")}
}

func (TelegramOAuthProvider) GetName() string { return "Telegram" }
// IsEnabled remains false until the unified callback/session integration is
// complete; this prevents the scaffold from competing with the legacy route.
func (TelegramOAuthProvider) IsEnabled() bool { return false }
func (p TelegramOAuthProvider) ExchangeToken(ctx context.Context, code string, c *gin.Context) (*OAuthToken, error) {
	if c == nil { return nil, ErrTelegramOAuthNotReady }
	value, _ := c.Get(TelegramOAuthFlowContextKey)
	flow, ok := value.(*TelegramOAuthFlow)
	if !ok { return nil, ErrTelegramOAuthNotReady }
	return ExchangeTelegramCode(ctx, p.client, flow, code)
}
func (p TelegramOAuthProvider) GetUserInfo(ctx context.Context, token *OAuthToken) (*OAuthUser, error) {
	return VerifyTelegramIDToken(ctx, token, p.keys)
}
func (TelegramOAuthProvider) IsUserIDTaken(id string) bool { return model.IsTelegramIdAlreadyTaken(id) }
func (TelegramOAuthProvider) FillUserByProviderID(user *model.User, id string) error {
	if user == nil || id == "" { return errors.New("invalid telegram identity") }
	stored := model.User{TelegramId: id}
	if err := model.DB.Where("telegram_id = ?", id).First(&stored).Error; err != nil { return err }
	*user = stored
	return nil
}
func (TelegramOAuthProvider) SetProviderUserID(user *model.User, id string) { user.TelegramId = id }
func (TelegramOAuthProvider) GetProviderPrefix() string { return "telegram_" }

// Deliberately not registered until single-use authorization flows and
// session-bound callbacks are integrated. Legacy Telegram routes are unchanged.
