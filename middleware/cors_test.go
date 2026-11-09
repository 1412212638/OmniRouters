package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/gin-gonic/gin"
)

func TestDashboardCORSLoginAndRefreshDeploymentOrigins(t *testing.T) {
	previousAddress := system_setting.ServerAddress
	previousSecure := common.SessionCookieSecure
	previousTrusted := common.SessionCookieTrustedURLs
	system_setting.ServerAddress = "https://console.example.org"
	common.SessionCookieSecure = true
	common.SessionCookieTrustedURLs = []string{"https://admin.example.org"}
	t.Cleanup(func() {
		system_setting.ServerAddress = previousAddress
		common.SessionCookieSecure = previousSecure
		common.SessionCookieTrustedURLs = previousTrusted
	})
	// Public API settings must not break dashboard authentication.
	setTestCORSOptions(t, "all", `[]`)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CORS())
	reached := func(c *gin.Context) { c.Status(http.StatusNoContent) }
	router.POST("/api/user/login", reached)
	router.POST("/api/user/auth/refresh", SessionCookieOriginGuard(), reached)

	for _, tc := range []struct {
		name       string
		target     string
		origin     string
		forwarded  string
		wantStatus int
	}{
		{"direct HTTPS", "https://tenant.example.org", "https://tenant.example.org", "", http.StatusNoContent},
		{"TLS terminates at proxy", "http://tenant.example.org", "https://tenant.example.org", "", http.StatusNoContent},
		{"configured trusted origin with internal host", "http://gateway:3000", "https://admin.example.org", "", http.StatusNoContent},
		{"configured site origin with internal host", "http://gateway:3000", "https://console.example.org", "", http.StatusNoContent},
		{"foreign origin", "http://tenant.example.org", "https://evil.example.org", "", http.StatusForbidden},
		{"lookalike host", "http://tenant.example.org", "https://tenant.example.org.evil.test", "", http.StatusForbidden},
		{"different port", "http://tenant.example.org", "https://tenant.example.org:9443", "", http.StatusForbidden},
		{"HTTP downgrade", "https://tenant.example.org", "http://tenant.example.org", "", http.StatusForbidden},
		{"forged forwarded host", "http://gateway:3000", "https://evil.example.org", "evil.example.org", http.StatusForbidden},
		{"null origin", "https://tenant.example.org", "null", "", http.StatusForbidden},
		{"malformed origin", "https://tenant.example.org", "https://tenant.example.org/path", "", http.StatusForbidden},
	} {
		for _, path := range []string{"/api/user/login", "/api/user/auth/refresh"} {
			t.Run(tc.name+path, func(t *testing.T) {
				request := httptest.NewRequest(http.MethodPost, tc.target+path, nil)
				request.Header.Set("Origin", tc.origin)
				request.Header.Set("X-Forwarded-Host", tc.forwarded)
				recorder := httptest.NewRecorder()
				router.ServeHTTP(recorder, request)
				if recorder.Code != tc.wantStatus {
					t.Fatalf("status=%d, want=%d, body=%s", recorder.Code, tc.wantStatus, recorder.Body.String())
				}
				if tc.wantStatus == http.StatusNoContent {
					if recorder.Header().Get("Access-Control-Allow-Origin") != tc.origin ||
						recorder.Header().Get("Access-Control-Allow-Credentials") != "true" {
						t.Fatal("dashboard must retain exact-origin credential behavior")
					}
				} else if recorder.Header().Get("Access-Control-Allow-Origin") != "" {
					t.Fatal("rejected origins must not receive CORS access")
				}
			})
		}
	}
}

func setTestCORSOptions(t *testing.T, mode, origins string) {
	t.Helper()
	common.OptionMapRWMutex.Lock()
	if common.OptionMap == nil {
		common.OptionMap = make(map[string]string)
	}
	oldMode, hadMode := common.OptionMap["cors_setting.mode"]
	oldOrigins, hadOrigins := common.OptionMap["cors_setting.allowed_origins"]
	common.OptionMap["cors_setting.mode"] = mode
	common.OptionMap["cors_setting.allowed_origins"] = origins
	common.OptionMapRWMutex.Unlock()
	t.Cleanup(func() {
		common.OptionMapRWMutex.Lock()
		if hadMode {
			common.OptionMap["cors_setting.mode"] = oldMode
		} else {
			delete(common.OptionMap, "cors_setting.mode")
		}
		if hadOrigins {
			common.OptionMap["cors_setting.allowed_origins"] = oldOrigins
		} else {
			delete(common.OptionMap, "cors_setting.allowed_origins")
		}
		common.OptionMapRWMutex.Unlock()
	})
}

func TestCORSAllowsTokenPlanWithCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CORS())
	router.GET("/api/status", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	request.Header.Set("Origin", "https://tokenplan.omnirouters.com")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, recorder.Code)
	}
	if origin := recorder.Header().Get("Access-Control-Allow-Origin"); origin != "https://tokenplan.omnirouters.com" {
		t.Fatalf("unexpected allowed origin %q", origin)
	}
	if credentials := recorder.Header().Get("Access-Control-Allow-Credentials"); credentials != "true" {
		t.Fatalf("expected credentials to be allowed, got %q", credentials)
	}
}

func TestCORSAllowsOmniRoutersOriginsWithCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CORS())
	router.PUT("/api/user/invoice-profile", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	for _, origin := range []string{
		"https://omnirouters.com.cn",
		"https://www.omnirouters.com.cn",
	} {
		t.Run(origin, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPut, "/api/user/invoice-profile", nil)
			request.Header.Set("Origin", origin)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusNoContent {
				t.Fatalf("expected status %d, got %d", http.StatusNoContent, recorder.Code)
			}
			if allowedOrigin := recorder.Header().Get("Access-Control-Allow-Origin"); allowedOrigin != origin {
				t.Fatalf("unexpected allowed origin %q", allowedOrigin)
			}
			if credentials := recorder.Header().Get("Access-Control-Allow-Credentials"); credentials != "true" {
				t.Fatalf("expected credentials to be allowed, got %q", credentials)
			}
		})
	}
}

func TestCORSHandlesOmniRoutersInvoiceProfilePreflight(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CORS())
	router.PUT("/api/user/invoice-profile", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodOptions, "/api/user/invoice-profile", nil)
	request.Header.Set("Origin", "https://omnirouters.com.cn")
	request.Header.Set("Access-Control-Request-Method", http.MethodPut)
	request.Header.Set("Access-Control-Request-Headers", "content-type,new-api-user")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, recorder.Code)
	}
	if origin := recorder.Header().Get("Access-Control-Allow-Origin"); origin != "https://omnirouters.com.cn" {
		t.Fatalf("unexpected allowed origin %q", origin)
	}
	if headers := recorder.Header().Get("Access-Control-Allow-Headers"); headers != "Content-Type,New-Api-User" {
		t.Fatalf("unexpected allowed headers %q", headers)
	}
}

func TestCORSHandlesTokenPlanPreflight(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CORS())
	router.GET("/api/status", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodOptions, "/api/status", nil)
	request.Header.Set("Origin", "https://tokenplan.omnirouters.com")
	request.Header.Set("Access-Control-Request-Method", http.MethodGet)
	request.Header.Set("Access-Control-Request-Headers", "content-type,new-api-user")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, recorder.Code)
	}
	if origin := recorder.Header().Get("Access-Control-Allow-Origin"); origin != "https://tokenplan.omnirouters.com" {
		t.Fatalf("unexpected allowed origin %q", origin)
	}
	if headers := recorder.Header().Get("Access-Control-Allow-Headers"); headers != "Content-Type,New-Api-User" {
		t.Fatalf("unexpected allowed headers %q", headers)
	}
}

func TestCORSRejectsUntrustedOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CORS())
	router.GET("/api/status", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	request.Header.Set("Origin", "https://example.com")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, recorder.Code)
	}
}

func TestCORSAllowsConfiguredPublicAPIOrigin(t *testing.T) {
	setTestCORSOptions(t, "allowlist", `["https://client.example.com"]`)
	gIn := gin.New()
	gIn.Use(CORS())
	gIn.POST("/v1/chat/completions", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	request := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	request.Header.Set("Origin", "https://client.example.com")
	recorder := httptest.NewRecorder()
	gIn.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent || recorder.Header().Get("Access-Control-Allow-Origin") != "https://client.example.com" {
		t.Fatalf("expected configured public API origin to be allowed, status=%d origin=%q", recorder.Code, recorder.Header().Get("Access-Control-Allow-Origin"))
	}
	if recorder.Header().Get("Access-Control-Allow-Credentials") != "" {
		t.Fatal("public API CORS must not allow credentials")
	}
}

func TestCORSAllowsAllPublicAPIOriginsWithoutCredentials(t *testing.T) {
	setTestCORSOptions(t, "all", `[]`)
	gIn := gin.New()
	gIn.Use(CORS())
	gIn.GET("/v1/models", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	request := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	request.Header.Set("Origin", "https://any.example.com")
	recorder := httptest.NewRecorder()
	gIn.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent || recorder.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatalf("expected all public API origins to be allowed, status=%d origin=%q", recorder.Code, recorder.Header().Get("Access-Control-Allow-Origin"))
	}
	if recorder.Header().Get("Access-Control-Allow-Credentials") != "" {
		t.Fatal("public API wildcard CORS must not allow credentials")
	}
}

func TestCORSDisablesPublicAPIAccess(t *testing.T) {
	setTestCORSOptions(t, "disabled", `[]`)
	gIn := gin.New()
	gIn.Use(CORS())
	gIn.GET("/v1/models", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	request := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	request.Header.Set("Origin", "https://client.example.com")
	recorder := httptest.NewRecorder()
	gIn.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent || recorder.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatalf("expected public API CORS to be disabled, status=%d origin=%q", recorder.Code, recorder.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestPublicCORSAllowAllDoesNotOpenDashboardRoutes(t *testing.T) {
	setTestCORSOptions(t, "all", `[]`)
	router := gin.New()
	router.Use(CORS())
	for _, path := range []string{"/api/user/login", "/api/user/auth/refresh", "/v1/dashboard/test"} {
		router.POST(path, func(c *gin.Context) { c.Status(http.StatusNoContent) })
	}
	for _, path := range []string{"/api/user/login", "/api/user/auth/refresh", "/v1/dashboard/test"} {
		request := httptest.NewRequest(http.MethodPost, "https://tenant.example.org"+path, nil)
		request.Header.Set("Origin", "https://foreign.example.org")
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusForbidden || recorder.Header().Get("Access-Control-Allow-Origin") != "" {
			t.Fatalf("public API wildcard must not grant dashboard access: path=%s status=%d", path, recorder.Code)
		}
		var response struct {
			Code string `json:"code"`
		}
		if err := common.Unmarshal(recorder.Body.Bytes(), &response); err != nil || response.Code != "CORS_ORIGIN_FORBIDDEN" {
			t.Fatalf("expected an actionable CORS rejection: body=%s err=%v", recorder.Body.String(), err)
		}
	}
}
