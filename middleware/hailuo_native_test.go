package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	pluginruntime "github.com/QuantumNous/new-api/pkg/jsplugin"
	builtinplugins "github.com/QuantumNous/new-api/plugins"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestH3NativeTaskOwnershipAndSubmitPreparation(t *testing.T) {
	source, err := builtinplugins.Source("hailuo")
	require.NoError(t, err)
	registry := pluginruntime.NewRegistry()
	plugin, err := registry.RegisterFactory(source, pluginruntime.Options{Key: "hailuo"})
	require.NoError(t, err)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	previous := model.DB
	model.DB = db
	t.Cleanup(func() {
		model.DB = previous
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	})
	require.NoError(t, db.AutoMigrate(&model.Task{}))
	data, err := common.Marshal(map[string]any{"task": map[string]any{
		"id": "private-id", "model": "UPSTREAM-MODEL", "status": "succeeded", "usage": map[string]any{"output_seconds": 5},
	}})
	require.NoError(t, err)
	for _, task := range []model.Task{
		{TaskID: "own", Platform: "hailuo", UserId: 1, Status: model.TaskStatusSuccess, Data: data, Properties: model.Properties{OriginModelName: "minimax-h3"}},
		{TaskID: "other-user", Platform: "hailuo", UserId: 2, Status: model.TaskStatusSuccess, Data: data},
		{TaskID: "other-org", Platform: "hailuo", UserId: 1, OrganizationId: 7, Status: model.TaskStatusSuccess, Data: data},
		{TaskID: "other-plugin", Platform: "doubao", UserId: 1, Status: model.TaskStatusSuccess, Data: data},
	} {
		require.NoError(t, db.Create(&task).Error)
	}
	pin := func(route pluginruntime.Route) gin.HandlerFunc {
		return func(c *gin.Context) {
			common.SetContextKey(c, constant.ContextKeyUserId, 1)
			c.Set(pluginruntime.ContextKeyPinnedRoute, pluginruntime.PinnedRoute{Generation: registry.Generation(), Plugin: plugin, Route: route})
		}
	}
	router := gin.New()
	router.GET("/v2/query/video_generation/:task_id", pin(plugin.Meta.Routes[1]), PrepareTaskPluginRoute(), func(c *gin.Context) {
		t.Error("queries must not enter distribution or billing")
	})
	for _, tc := range []struct {
		id   string
		code int
	}{{"own", 200}, {"other-user", 404}, {"other-org", 404}, {"other-plugin", 404}, {"missing", 404}} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v2/query/video_generation/"+tc.id, nil))
		require.Equal(t, tc.code, response.Code, response.Body.String())
		if tc.code == 200 {
			var result map[string]any
			require.NoError(t, common.Unmarshal(response.Body.Bytes(), &result))
			output := result["task"].(map[string]any)
			require.Equal(t, "own", output["id"])
			require.Equal(t, "minimax-h3", output["model"])
			require.Contains(t, output, "usage")
			require.NotContains(t, response.Body.String(), "UPSTREAM-MODEL")
		}
	}
	submits := 0
	router.POST("/v2/video_generation", pin(plugin.Meta.Routes[0]), PrepareTaskPluginRoute(), func(c *gin.Context) {
		submits++
		require.Equal(t, "hailuo", c.GetString("expected_task_plugin_key"))
		require.Equal(t, "minimax-h3", c.GetString("resolved_task_model"))
		body := c.MustGet("task_request").(map[string]any)
		require.Contains(t, body, "__hailuo_h3_native_body")
		c.Status(http.StatusNoContent)
	})
	for _, tc := range []struct {
		model string
		code  int
	}{{"minimax-h3", 204}, {"MiniMax-Hailuo-2.3", 400}} {
		request := httptest.NewRequest(http.MethodPost, "/v2/video_generation", strings.NewReader(`{"model":"`+tc.model+`","duration":5,"content":[{"type":"text","text":"a cat"}]}`))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		require.Equal(t, tc.code, response.Code, response.Body.String())
	}
	require.Equal(t, 1, submits)
}

func TestH3NativeCORSUsesPublicAPIPolicyOnlyForNewEndpoints(t *testing.T) {
	setTestCORSOptions(t, "all", `[]`)
	router := gin.New()
	router.Use(CORS())
	for _, tc := range []struct{ path, method string }{{"/v2/video_generation", "POST"}, {"/v2/query/video_generation/task_public", "GET"}} {
		request := httptest.NewRequest(http.MethodOptions, tc.path, nil)
		request.Header.Set("Origin", "https://customer.example")
		request.Header.Set("Access-Control-Request-Method", tc.method)
		request.Header.Set("Access-Control-Request-Headers", "authorization,content-type")
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		require.Equal(t, http.StatusNoContent, response.Code)
		require.Equal(t, "*", response.Header().Get("Access-Control-Allow-Origin"))
		require.Empty(t, response.Header().Get("Access-Control-Allow-Credentials"))
	}
	for _, path := range []string{"/v2/admin", "/v2/query/video_generation/", "/v2/query/video_generation/id/content", "/api/user/self"} {
		require.False(t, isPublicAPIPath(path), path)
	}
}
