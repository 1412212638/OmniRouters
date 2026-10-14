package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

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
