package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestH3NativeRoutesRequireTokenAuthentication(t *testing.T) {
	engine := gin.New()
	SetTaskPluginProtocolRouter(engine)
	for _, tc := range []struct{ method, path string }{{http.MethodPost, "/v2/video_generation"}, {http.MethodGet, "/v2/query/video_generation/task_public"}} {
		response := httptest.NewRecorder()
		engine.ServeHTTP(response, httptest.NewRequest(tc.method, tc.path, nil))
		require.Equal(t, http.StatusUnauthorized, response.Code, response.Body.String())
	}
}
