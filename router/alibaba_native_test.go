package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/QuantumNous/new-api/model"
	pluginruntime "github.com/QuantumNous/new-api/pkg/jsplugin"
	builtinplugins "github.com/QuantumNous/new-api/plugins"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestAlibabaNativeAPIRoutesRequireToken(t *testing.T) {
	engine := gin.New()
	SetApiRouter(engine)
	for _, tc := range []struct{ method, path string }{
		{http.MethodPost, "/api/v1/services/aigc/video-generation/video-synthesis"},
		{http.MethodPost, "/api/v1/services/aigc/image2video/video-synthesis"},
		{http.MethodGet, "/api/v1/tasks/task_public"},
	} {
		response := httptest.NewRecorder()
		engine.ServeHTTP(response, httptest.NewRequest(tc.method, tc.path, nil))
		require.Equal(t, http.StatusUnauthorized, response.Code, response.Body.String())
	}
}

func TestAlibabaNativeAliasesReusePluginPipeline(t *testing.T) {
	source, err := builtinplugins.Source("alibaba")
	require.NoError(t, err)
	previousRegistry := pluginruntime.DefaultRegistry
	pluginruntime.DefaultRegistry = pluginruntime.NewRegistry()
	t.Cleanup(func() { pluginruntime.DefaultRegistry = previousRegistry })
	availability := gin.New()
	availability.POST("/api/v1/services/aigc/video-generation/video-synthesis",
		pinDeclaredTaskPluginRoute(http.MethodPost, "/ali/api/v1/services/aigc/video-generation/video-synthesis"),
		func(c *gin.Context) { c.Status(http.StatusNoContent) },
	)
	response := httptest.NewRecorder()
	availability.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/v1/services/aigc/video-generation/video-synthesis", nil))
	require.Equal(t, http.StatusNotFound, response.Code)
	require.Contains(t, response.Body.String(), "task_plugin_route_unavailable")
	_, err = pluginruntime.DefaultRegistry.RegisterFactory(source, pluginruntime.Options{Key: "alibaba"})
	require.NoError(t, err)
	response = httptest.NewRecorder()
	availability.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/v1/services/aigc/video-generation/video-synthesis", nil))
	require.Equal(t, http.StatusNoContent, response.Code)

	// Stop before channel selection: compare the actual prepared billing input,
	// rather than issuing a paid upstream request.
	engine := gin.New()
	var prepared []map[string]any
	for _, service := range []string{"video-generation", "image2video"} {
		path := "/api/v1/services/aigc/" + service + "/video-synthesis"
		for _, prefix := range []string{"", "/ali"} {
			engine.POST(prefix+path, pinDeclaredTaskPluginRoute(http.MethodPost, "/ali"+path), middleware.PrepareTaskPluginRoute(), func(c *gin.Context) {
				require.Equal(t, "alibaba", c.GetString("expected_task_plugin_key"))
				require.Equal(t, "wan3.0-video", c.GetString("resolved_task_model"))
				prepared = append(prepared, c.MustGet("task_request").(map[string]any))
				c.Status(http.StatusNoContent)
			})
			request := httptest.NewRequest(http.MethodPost, prefix+path, strings.NewReader(`{"model":"wan3.0-video","input":{"prompt":"a cat"},"parameters":{"duration":5,"resolution":"720P","watermark":false}}`))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			engine.ServeHTTP(response, request)
			require.Equal(t, http.StatusNoContent, response.Code, response.Body.String())
		}
	}
	require.Len(t, prepared, 4)
	for _, body := range prepared[1:] {
		require.Equal(t, prepared[0], body)
	}

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	previousDB := model.DB
	model.DB = db
	t.Cleanup(func() {
		model.DB = previousDB
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	})
	require.NoError(t, db.AutoMigrate(&model.Task{}))
	data, err := common.Marshal(map[string]any{"output": map[string]any{
		"task_id": "private-upstream-id", "task_status": "SUCCEEDED", "video_url": "https://cdn.example/video.mp4",
	}})
	require.NoError(t, err)
	for _, task := range []model.Task{
		{TaskID: "own", Platform: "alibaba", UserId: 1, Status: model.TaskStatusSuccess, Data: data, Quota: 123},
		{TaskID: "other-user", Platform: "alibaba", UserId: 2, Status: model.TaskStatusSuccess, Data: data},
		{TaskID: "other-org", Platform: "alibaba", UserId: 1, OrganizationId: 7, Status: model.TaskStatusSuccess, Data: data},
		{TaskID: "other-plugin", Platform: "doubao", UserId: 1, Status: model.TaskStatusSuccess, Data: data},
	} {
		require.NoError(t, db.Create(&task).Error)
	}
	queries := gin.New()
	for _, prefix := range []string{"", "/ali"} {
		queries.GET(prefix+"/api/v1/tasks/:task_id", func(c *gin.Context) {
			common.SetContextKey(c, constant.ContextKeyUserId, 1)
		}, pinDeclaredTaskPluginRoute(http.MethodGet, "/ali/api/v1/tasks/:task_id"), func(c *gin.Context) {
			pinned := c.MustGet(pluginruntime.ContextKeyPinnedRoute).(pluginruntime.PinnedRoute)
			require.Equal(t, "task_id", pinned.Route.TaskIDParam)
			require.NotEmpty(t, c.Param(pinned.Route.TaskIDParam))
		}, middleware.PrepareTaskPluginRoute(), func(c *gin.Context) {
			t.Error("query must terminate before channel distribution and billing")
		})
	}
	for _, prefix := range []string{"", "/ali"} {
		for _, tc := range []struct {
			id     string
			status int
		}{{"own", 200}, {"other-user", 404}, {"other-org", 404}, {"other-plugin", 404}, {"missing", 404}} {
			response := httptest.NewRecorder()
			queries.ServeHTTP(response, httptest.NewRequest(http.MethodGet, prefix+"/api/v1/tasks/"+tc.id, nil))
			require.Equal(t, tc.status, response.Code, response.Body.String())
			if tc.status == 200 {
				var result struct {
					Output struct {
						TaskID string `json:"task_id"`
						Status string `json:"task_status"`
					} `json:"output"`
				}
				require.NoError(t, common.Unmarshal(response.Body.Bytes(), &result))
				require.Equal(t, "own", result.Output.TaskID)
				require.Equal(t, "SUCCEEDED", result.Output.Status)
			}
		}
	}
	var task model.Task
	require.NoError(t, db.Where("task_id = ?", "own").First(&task).Error)
	require.EqualValues(t, 123, task.Quota)
}
