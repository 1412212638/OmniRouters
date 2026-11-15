package controller

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	pluginruntime "github.com/QuantumNous/new-api/pkg/jsplugin"
	builtinplugins "github.com/QuantumNous/new-api/plugins"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestHailuoFileTaskScope(t *testing.T) {
	for _, test := range []struct {
		name, platform, data string
		want                 bool
	}{
		{"legacy Hailuo", "hailuo", `{"file_id":"one"}`, true},
		{"numeric platform", strconv.Itoa(constant.ChannelTypeMiniMax), `{"file_id":"one"}`, true},
		{"numeric file id", "hailuo", `{"file_id":123}`, true},
		{"wrapped task", "hailuo", `{"data":{"task_id":"private-id","data":{"file_id":"one"}}}`, true},
		{"H3 direct URL", "hailuo", `{"task":{"content":{"url":"https://cdn.example/h3.mp4"}}}`, false},
		{"other plugin", "sora", `{"file_id":"one"}`, false},
		{"missing id", "hailuo", `{}`, false},
		{"blank id", "hailuo", `{"file_id":" "}`, false},
		{"invalid JSON", "hailuo", `{`, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			task := &model.Task{Platform: constant.TaskPlatform(test.platform), Data: []byte(test.data)}
			require.Equal(t, test.want, isHailuoFileTask(task))
		})
	}
	require.False(t, isHailuoFileTask(nil))
	task := &model.Task{Platform: "hailuo", Data: []byte(`{"file_id":"one"}`), PrivateData: model.TaskPrivateData{
		Execution: &model.TaskExecutionSnapshot{TaskPlugin: &model.TaskPluginSnapshot{Key: "another-plugin"}},
	}}
	require.False(t, isHailuoFileTask(task))
}

func TestHailuoFileContentHTTPRoutes(t *testing.T) {
	previousDB, previousCache, previousRedis := model.DB, common.MemoryCacheEnabled, common.RedisEnabled
	previousRegistry := pluginruntime.DefaultRegistry
	fetch := system_setting.GetFetchSetting()
	previousFetch := *fetch
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	raw, err := db.DB()
	require.NoError(t, err)
	model.DB, common.MemoryCacheEnabled, common.RedisEnabled = db, false, false
	pluginruntime.DefaultRegistry = pluginruntime.NewRegistry()
	pluginruntime.DefaultRegistry.SetEnabled(true)
	fetch.AllowPrivateIp = true
	fetch.AllowedPorts = []string{"1-65535"}
	service.InitHttpClient()
	t.Cleanup(func() {
		model.DB, common.MemoryCacheEnabled, common.RedisEnabled = previousDB, previousCache, previousRedis
		pluginruntime.DefaultRegistry = previousRegistry
		*fetch = previousFetch
		_ = raw.Close()
	})
	require.NoError(t, db.AutoMigrate(&model.Channel{}, &model.Task{}))
	source, err := builtinplugins.Source("hailuo")
	require.NoError(t, err)
	_, err = pluginruntime.DefaultRegistry.RegisterFactory(source, pluginruntime.Options{Key: "hailuo"})
	require.NoError(t, err)

	var metadataHits, videoHits atomic.Int32
	cdn := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		videoHits.Add(1)
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "signed", r.URL.Query().Get("signature"))
		for _, name := range []string{"Authorization", "Cookie", "X-Api-Key", "Proxy-Authorization"} {
			require.Empty(t, r.Header.Get(name), "channel and client credentials must not reach the CDN")
		}
		w.Header().Set("Content-Type", "video/mp4")
		_, _ = fmt.Fprint(w, "video-bytes")
	}))
	defer cdn.Close()
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metadataHits.Add(1)
		require.Equal(t, "/v1/files/retrieve", r.URL.Path)
		require.Equal(t, "file-id", r.URL.Query().Get("file_id"))
		require.Equal(t, "Bearer saved-task-key", r.Header.Get("Authorization"))
		body, marshalErr := common.Marshal(map[string]any{
			"base_resp": map[string]any{"status_code": 0}, "file": map[string]any{"download_url": cdn.URL + "/video.mp4?signature=signed"},
		})
		require.NoError(t, marshalErr)
		_, _ = w.Write(body)
	}))
	defer upstream.Close()
	baseURL := upstream.URL
	ch := &model.Channel{Id: 1, Type: constant.ChannelTypeMiniMax, BaseURL: &baseURL, Key: "different-current-key"}
	require.NoError(t, db.Create(ch).Error)
	task := &model.Task{TaskID: "public-id", Platform: "hailuo", UserId: 11, ChannelId: ch.Id, Status: model.TaskStatusSuccess,
		Quota: 123, Data: []byte(`{"status":"Success","file_id":"file-id"}`), PrivateData: model.TaskPrivateData{Key: "saved-task-key"}}
	require.NoError(t, db.Create(task).Error)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("id", 11)
		if c.GetHeader("X-Test-Other-User") != "" {
			c.Set("id", 22)
		}
		if c.GetHeader("X-Test-Other-Organization") != "" {
			common.SetContextKey(c, constant.ContextKeyOrganizationId, 9)
		}
	})
	for _, method := range []string{http.MethodGet, http.MethodHead} {
		router.Handle(method, "/v1/videos/:task_id/content", VideoProxy)
		router.Handle(method, "/v1/tasks/:key/artifacts/:artifact_key/content", TaskArtifactContent)
	}
	for _, path := range []string{"/v1/videos/public-id/content", "/v1/tasks/public-id/artifacts/video/content"} {
		for _, method := range []string{http.MethodGet, http.MethodHead} {
			response := httptest.NewRecorder()
			req := httptest.NewRequest(method, path, nil)
			req.Header.Set("Authorization", "Bearer client-platform-key")
			req.Header.Set("Cookie", "session=client-cookie")
			router.ServeHTTP(response, req)
			require.Equal(t, http.StatusOK, response.Code, response.Body.String())
			require.Equal(t, "video/mp4", response.Header().Get("Content-Type"))
			if method == http.MethodHead {
				require.Empty(t, response.Body.String())
			} else {
				require.Equal(t, "video-bytes", response.Body.String())
			}
		}
		for _, header := range []string{"X-Test-Other-User", "X-Test-Other-Organization"} {
			before := metadataHits.Load()
			response := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, path, nil)
			req.Header.Set(header, "true")
			router.ServeHTTP(response, req)
			require.Equal(t, http.StatusNotFound, response.Code)
			require.Equal(t, before, metadataHits.Load(), "ownership must be checked before contacting the upstream")
		}
	}
	require.EqualValues(t, 4, metadataHits.Load())
	require.EqualValues(t, 4, videoHits.Load())
	var saved model.Task
	require.NoError(t, db.First(&saved, task.ID).Error)
	require.Equal(t, 123, saved.Quota)
	require.EqualValues(t, model.TaskStatusSuccess, saved.Status)
	require.Equal(t, "saved-task-key", saved.PrivateData.Key)

	// The configured channel is trusted for metadata; its returned CDN URL
	// still passes the normal media SSRF boundary before any download.
	fetch.EnableSSRFProtection, fetch.AllowPrivateIp = true, false
	before := videoHits.Load()
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v1/videos/public-id/content", nil))
	require.Equal(t, http.StatusForbidden, response.Code)
	require.Equal(t, before, videoHits.Load())
}

func TestHailuoFileContentRequiresOriginalMultiKeyCredential(t *testing.T) {
	previous := pluginruntime.DefaultRegistry
	pluginruntime.DefaultRegistry = pluginruntime.NewRegistry()
	t.Cleanup(func() { pluginruntime.DefaultRegistry = previous })
	source, err := builtinplugins.Source("hailuo")
	require.NoError(t, err)
	_, err = pluginruntime.DefaultRegistry.RegisterFactory(source, pluginruntime.Options{Key: "hailuo"})
	require.NoError(t, err)
	_, err = buildHailuoFileContentRequest(&model.Task{Platform: "hailuo"}, &model.Channel{
		Key: "key-one\nkey-two", ChannelInfo: model.ChannelInfo{IsMultiKey: true},
	}, http.MethodGet)
	require.ErrorContains(t, err, "credential is unavailable")
}
