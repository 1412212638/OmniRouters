package router

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAPIRouterRegistersTelegramRoutesOnce(t *testing.T) {
	engine := gin.New()
	// Register the real API tree: duplicate routes panic before serving traffic.
	SetApiRouter(engine)
	counts := map[string]int{}
	for _, route := range engine.Routes() {
		if route.Method == http.MethodGet {
			counts[route.Path]++
		}
	}
	for _, path := range []string{
		"/api/oauth/telegram/start",
		"/api/oauth/telegram",
		"/api/oauth/telegram/login",
		"/api/oauth/telegram/bind",
		"/api/oauth/:provider",
	} {
		if counts[path] != 1 {
			t.Errorf("GET %s registered %d times, want 1", path, counts[path])
		}
	}
}
