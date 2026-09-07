package oauth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/golang-jwt/jwt/v5"
)

type telegramTestTransport func(*http.Request) (*http.Response, error)

func (f telegramTestTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func configureTelegramTest(t *testing.T) {
	t.Helper()
	settings := system_setting.GetTelegramSettings()
	previous, address := *settings, system_setting.ServerAddress
	t.Cleanup(func() { *settings = previous; system_setting.ServerAddress = address })
	*settings = system_setting.TelegramSettings{Enabled: true, ClientID: "123", ClientSecret: "secret"}
	system_setting.ServerAddress = "https://gateway.example"
}

func TestTelegramIdentityNumbers(t *testing.T) {
	for _, id := range []string{"1", "9007199254740993", "18446744073709551615"} {
		user, err := parseTelegramIdentity([]byte(`{"id":`+id+`}`), "subject")
		if err != nil || user.ProviderUserID != id { t.Fatalf("id %s: user=%v err=%v", id, user, err) }
	}
	for _, id := range []string{"0", "-1", "1.5", "1e3", "18446744073709551616", "null", `"abc"`} {
		if _, err := parseTelegramIdentity([]byte(`{"id":`+id+`}`), "subject"); err == nil { t.Fatalf("accepted id %s", id) }
	}
	if _, err := parseTelegramIdentity([]byte(`{"id":1}`), ""); err == nil { t.Fatal("accepted empty subject") }
}

func TestTelegramRedirectAndConfigurationBinding(t *testing.T) {
	configureTelegramTest(t)
	for _, address := range []string{"ftp://gateway.example", "//gateway.example", "https://u:p@gateway.example", "https://gateway.example?x=1", "https://gateway.example#x"} {
		system_setting.ServerAddress = address
		if _, err := NewTelegramOAuthFlow(); err == nil { t.Fatalf("accepted %s", address) }
	}
	system_setting.ServerAddress = "https://gateway.example"
	flow, err := NewTelegramOAuthFlow()
	if err != nil { t.Fatal(err) }
	client := &http.Client{Transport: telegramTestTransport(func(r *http.Request) (*http.Response, error) {
		t.Fatal("invalid flow reached token endpoint")
		return nil, nil
	})}
	flow.RedirectURI = "https://other.example/oauth/telegram"
	if _, err := ExchangeTelegramCode(context.Background(), client, flow, "code"); err == nil { t.Fatal("accepted changed redirect") }
	flow.RedirectURI = "https://gateway.example/oauth/telegram"
	flow.ClientID = "other"
	if _, err := ExchangeTelegramCode(context.Background(), client, flow, "code"); err == nil { t.Fatal("accepted changed client") }
}

func TestTelegramExchangeProtocolAndPrivateClientID(t *testing.T) {
	configureTelegramTest(t)
	flow, err := NewTelegramOAuthFlow()
	if err != nil { t.Fatal(err) }
	client := &http.Client{Transport: telegramTestTransport(func(r *http.Request) (*http.Response, error) {
		id, secret, ok := r.BasicAuth()
		if !ok || id != "123" || secret != "secret" || r.URL.String() != TelegramOAuthIssuer+"/token" || r.Method != http.MethodPost { t.Fatal("invalid token request") }
		if err := r.ParseForm(); err != nil { t.Fatal(err) }
		if r.Form.Get("code_verifier") != flow.CodeVerifier || r.Form.Get("redirect_uri") != flow.RedirectURI { t.Fatal("lost PKCE/configuration binding") }
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"id_token":"jwt","client_id":"attacker"}`))}, nil
	})}
	token, err := ExchangeTelegramCode(context.Background(), client, flow, "code")
	if err != nil || token.ClientID != "123" { t.Fatalf("token=%v err=%v", token, err) }
	encoded, err := common.Marshal(token)
	if err != nil || strings.Contains(string(encoded), "client_id") { t.Fatal("client binding leaked into JSON") }
}

func TestTelegramMigrationIsNotRegistered(t *testing.T) {
	configureTelegramTest(t)
	if GetProvider("telegram_oauth") != nil { t.Fatal("unfinished provider registered") }
	provider := NewTelegramOAuthProvider(nil)
	if provider.IsEnabled() { t.Fatal("unfinished callback enabled") }
	if _, err := provider.ExchangeToken(context.Background(), "code", nil); err == nil { t.Fatal("accepted missing flow") }
}

func TestTelegramIDTokenVerification(t *testing.T) {
	configureTelegramTest(t)
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil { t.Fatal(err) }
	// Serve a real RSA JWKS through the provider's shared transport, without network access.
	provider := NewTelegramOAuthProvider(&http.Client{Transport: telegramTestTransport(func(r *http.Request) (*http.Response, error) {
		if r.URL.String() != TelegramOAuthIssuer+"/.well-known/jwks.json" { t.Fatal("unexpected JWKS URL") }
		body, err := common.Marshal(map[string]any{"keys": []any{map[string]any{
			"kty": "RSA", "kid": "test", "use": "sig", "alg": "RS256",
			"n": base64.RawURLEncoding.EncodeToString(key.N.Bytes()), "e": "AQAB",
		}}})
		if err != nil { t.Fatal(err) }
		return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(string(body)))}, nil
	})})
	for _, variant := range []string{"valid", "issuer", "audience", "expired", "subject", "signature"} {
		t.Run(variant, func(t *testing.T) {
			claims := jwt.MapClaims{"iss": TelegramOAuthIssuer, "aud": "123", "exp": time.Now().Add(time.Hour).Unix(), "sub": "subject", "id": 123}
			switch variant {
			case "issuer": claims["iss"] = "https://attacker.example"
			case "audience": claims["aud"] = "other"
			case "expired": claims["exp"] = time.Now().Add(-time.Hour).Unix()
			case "subject": claims["sub"] = ""
			}
			signer := key
			if variant == "signature" {
				signer, err = rsa.GenerateKey(rand.Reader, 2048)
				if err != nil { t.Fatal(err) }
			}
			token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
			token.Header["kid"] = "test"
			raw, err := token.SignedString(signer)
			if err != nil { t.Fatal(err) }
			user, err := provider.GetUserInfo(context.Background(), &OAuthToken{IDToken: raw, ClientID: "123"})
			if variant == "valid" {
				if err != nil || user.ProviderUserID != "123" { t.Fatalf("user=%v err=%v", user, err) }
			} else if err == nil { t.Fatal("accepted invalid token") }
		})
	}
}
