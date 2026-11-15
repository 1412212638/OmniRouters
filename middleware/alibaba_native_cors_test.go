package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestAlibabaNativeCORSIsLimitedToOfficialTaskAliases(t *testing.T) {
	for _, mode := range []string{"all", "allowlist", "disabled"} {
		t.Run(mode, func(t *testing.T) {
			setTestCORSOptions(t, mode, `["https://customer.example"]`)
			engine := gin.New()
			engine.Use(CORS())
			for _, path := range []string{
				"/api/v1/services/aigc/video-generation/video-synthesis",
				"/api/v1/services/aigc/image2video/video-synthesis",
				"/api/v1/tasks/task_public",
			} {
				for _, origin := range []string{"https://customer.example", "https://foreign.example"} {
					request := httptest.NewRequest(http.MethodOptions, path, nil)
					request.Header.Set("Origin", origin)
					request.Header.Set("Access-Control-Request-Method", "POST")
					request.Header.Set("Access-Control-Request-Headers", "authorization,content-type,x-dashscope-async")
					response := httptest.NewRecorder()
					engine.ServeHTTP(response, request)
					require.Empty(t, response.Header().Get("Access-Control-Allow-Credentials"))
					switch {
					case mode == "disabled":
						require.Empty(t, response.Header().Get("Access-Control-Allow-Origin"))
					case mode == "allowlist" && origin == "https://foreign.example":
						require.Equal(t, http.StatusForbidden, response.Code)
						require.Empty(t, response.Header().Get("Access-Control-Allow-Origin"))
					default:
						require.Equal(t, http.StatusNoContent, response.Code)
						require.Contains(t, response.Header().Get("Access-Control-Allow-Headers"), "X-DashScope-Async")
						if mode == "all" {
							require.Equal(t, "*", response.Header().Get("Access-Control-Allow-Origin"))
						} else {
							require.Equal(t, origin, response.Header().Get("Access-Control-Allow-Origin"))
						}
					}
				}
			}
		})
	}
	for _, path := range []string{
		"/api/user/login", "/api/option", "/api/v1/admin", "/api/v1/tasks/",
		"/api/v1/tasks/id/content", "/api/v1/services/aigc/video-generation/video-synthesis/admin",
	} {
		require.False(t, isPublicAPIPath(path), path)
	}
}
