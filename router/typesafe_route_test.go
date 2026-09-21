package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestTypeSafeNativeRouteRegistered(t *testing.T) {
	router := gin.New()
	SetTaskPluginProtocolRouter(router)
	for _, route := range router.Routes() {
		if route.Method == http.MethodPost && route.Path == "/v1/systemone" {
			return
		}
	}
	t.Fatal("TypeSafe native POST route must not fall through to the web frontend")
}

func TestTypeSafeMissingPluginReturnsJSON(t *testing.T) {
	router := gin.New()
	router.POST("/v1/systemone", pinTypeSafeRoute, func(c *gin.Context) {
		t.Error("missing plugin must stop before submission and billing")
	})
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/v1/systemone", nil))
	if response.Code != http.StatusNotFound || !strings.Contains(response.Body.String(), "task_plugin_route_unavailable") {
		t.Fatalf("expected explicit missing-plugin JSON error, got %d: %s", response.Code, response.Body.String())
	}
}
