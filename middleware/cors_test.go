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
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, recorder.Code)
	}
	if origin := recorder.Header().Get("Access-Control-Allow-Origin"); origin != "https://tokenplan.omnirouters.com" {
		t.Fatalf("unexpected allowed origin %q", origin)
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
