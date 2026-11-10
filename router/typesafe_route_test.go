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

func TestDoubaoNativeRoutesRegistered(t *testing.T) {
	router := gin.New()
	SetTaskPluginProtocolRouter(router)
	want := map[string]map[string]bool{
		http.MethodPost: {
			"/doubao/api/v3/contents/generations/tasks": false,
		},
		http.MethodGet: {
			"/doubao/api/v3/contents/generations/tasks/:task_id": false,
		},
	}
	for _, route := range router.Routes() {
		if paths, ok := want[route.Method]; ok {
			if _, exists := paths[route.Path]; exists {
				paths[route.Path] = true
			}
		}
	}
	for method, paths := range want {
		for path, found := range paths {
			if !found {
				t.Fatalf("Doubao native %s route %s must be registered", method, path)
			}
		}
	}
}

func TestAstraFlowNativeRoutesRegistered(t *testing.T) {
	router := gin.New()
	SetTaskRouter(router)
	want := map[string]map[string]bool{
		http.MethodPost: {"/v1/tasks/submit": false},
		http.MethodGet:  {"/v1/tasks/status": false, "/v1beta/models/interactions/:id": false},
	}
	for _, route := range router.Routes() {
		if paths, ok := want[route.Method]; ok {
			if _, exists := paths[route.Path]; exists {
				paths[route.Path] = true
			}
		}
	}
	for method, paths := range want {
		for path, found := range paths {
			if !found {
				t.Fatalf("AstraFlow native %s route %s must be registered", method, path)
			}
		}
	}
}

func TestAstraFlowInteractionsCoexistWithGeminiWildcard(t *testing.T) {
	router := gin.New()
	SetRelayRouter(router)
	SetTaskRouter(router)
	request := httptest.NewRequest(http.MethodPost, "/v1beta/models/interactions", nil)
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = request
	if !isAstraFlowInteraction(context) {
		t.Fatal("native interaction submit must enter the task-plugin pipeline")
	}
	context.Request = httptest.NewRequest(http.MethodPost, "/v1beta/models/gemini-3-pro:generateContent", nil)
	if isAstraFlowInteraction(context) {
		t.Fatal("ordinary Gemini requests must retain the Gemini relay")
	}
}

func TestDoubaoNativeAPIRoutesRegisteredInAPIGroup(t *testing.T) {
	router := gin.New()
	SetApiRouter(router)
	want := map[string]map[string]bool{
		http.MethodPost: {"/api/v3/contents/generations/tasks": false},
		http.MethodGet:  {"/api/v3/contents/generations/tasks/:task_id": false},
	}
	for _, route := range router.Routes() {
		if paths, ok := want[route.Method]; ok {
			if _, exists := paths[route.Path]; exists {
				paths[route.Path] = true
			}
		}
	}
	for method, paths := range want {
		for path, found := range paths {
			if !found {
				t.Fatalf("Doubao native API-group %s route %s must be registered", method, path)
			}
		}
	}
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
