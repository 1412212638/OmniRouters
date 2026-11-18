package taskcallback_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/controller"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/pkg/billingexpr"
	pluginruntime "github.com/QuantumNous/new-api/pkg/jsplugin"
	builtinplugins "github.com/QuantumNous/new-api/plugins"
	"github.com/QuantumNous/new-api/relay"
	jsadaptor "github.com/QuantumNous/new-api/relay/channel/task/jsplugin"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setup(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "callbacks.db")), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	oldDB, oldLog := model.DB, model.LOG_DB
	oldFetch := *system_setting.GetFetchSetting()
	oldAddress, oldCustom := system_setting.ServerAddress, operation_setting.CustomCallbackAddress
	oldMemory, oldRedis, oldBatch := common.MemoryCacheEnabled, common.RedisEnabled, common.BatchUpdateEnabled
	oldFactory := service.GetTaskAdaptorFunc
	model.DB, model.LOG_DB = db, db
	common.SetDatabaseTypes(common.DatabaseTypeSQLite, common.DatabaseTypeSQLite)
	common.MemoryCacheEnabled, common.RedisEnabled, common.BatchUpdateEnabled = false, false, false
	system_setting.GetFetchSetting().EnableSSRFProtection = false // Loopback test servers only.
	system_setting.ServerAddress, operation_setting.CustomCallbackAddress = "https://gateway.example", ""
	service.InitHttpClient()
	require.NoError(t, db.AutoMigrate(&model.TaskCallback{}, &model.Task{}, &model.Channel{}, &model.User{}, &model.Token{}, &model.Log{}))
	t.Cleanup(func() {
		model.DB, model.LOG_DB = oldDB, oldLog
		*system_setting.GetFetchSetting() = oldFetch
		system_setting.ServerAddress, operation_setting.CustomCallbackAddress = oldAddress, oldCustom
		common.MemoryCacheEnabled, common.RedisEnabled, common.BatchUpdateEnabled = oldMemory, oldRedis, oldBatch
		service.GetTaskAdaptorFunc = oldFactory
		sqlDB.Close()
	})
}

func loadPlugin(t *testing.T) *pluginruntime.LoadedPlugin {
	t.Helper()
	source, err := builtinplugins.Source("doubao")
	require.NoError(t, err)
	plugin, err := pluginruntime.NewRegistry().RegisterFactory(source, pluginruntime.Options{Key: "doubao"})
	require.NoError(t, err)
	return plugin
}

func TestDoubaoCallbackRewritesUpstreamDestinationAndKeepsClientBody(t *testing.T) {
	setup(t)
	adaptor := jsadaptor.New(loadPlugin(t))
	info := &relaycommon.RelayInfo{OriginModelName: "doubao-seedance-2.5", TaskRelayInfo: &relaycommon.TaskRelayInfo{PublicTaskID: "task_submit"},
		ChannelMeta: &relaycommon.ChannelMeta{UpstreamModelName: "doubao-seedance-2.5", ChannelType: constant.ChannelTypeVolcEngine, ChannelBaseUrl: "https://ark.example", ApiKey: "upstream-secret"}}
	adaptor.Init(info)
	body := map[string]any{"model": "doubao-seedance-2.5", "resolution": "720p", "callback_url": "https://customer.example/hook?customer=1",
		"content": []any{map[string]any{"type": "text", "text": "a cat"}}, "priority": 0, "service_tier": "default"}
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v3/contents/generations/tasks", nil)
	c.Set("task_request", map[string]any{"model": "doubao-seedance-2.5", "metadata": body, "__doubao_native_body": body})
	reader, err := adaptor.BuildRequestBody(c, info)
	require.NoError(t, err)
	var upstream map[string]any
	require.NoError(t, common.DecodeJson(reader, &upstream))
	u, err := url.Parse(upstream["callback_url"].(string))
	require.NoError(t, err)
	require.Equal(t, "gateway.example", u.Host)
	require.Equal(t, service.DoubaoCallbackPath+"task_submit", u.Path)
	require.Len(t, u.Query().Get("token"), 48)
	require.Equal(t, "https://customer.example/hook?customer=1", body["callback_url"])
	require.Equal(t, float64(0), upstream["priority"])
	require.Equal(t, "default", upstream["service_tier"])
	var row model.TaskCallback
	require.NoError(t, model.DB.First(&row).Error)
	require.Equal(t, body["callback_url"], row.TargetURL)
	require.Zero(t, row.ReadyAt)
}

func TestDoubaoDraftBillingFactsReachPluginFromPersistedOrigin(t *testing.T) {
	setup(t)
	for _, source := range []string{"expression_snapshot", "legacy_plugin_state"} {
		t.Run(source, func(t *testing.T) {
			origin := &model.Task{TaskID: "task_sample", PrivateData: model.TaskPrivateData{UpstreamTaskID: "cgt-sample"}}
			if source == "expression_snapshot" {
				origin.PrivateData.BillingContext = &model.TaskBillingContext{TieredSnapshot: &billingexpr.BillingSnapshot{UsageFacts: map[string]any{"video_input": "video"}}}
			} else {
				origin.PrivateData.PluginState = []byte(`{"seedance_video_input":"video"}`)
			}
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodPost, "/api/v3/contents/generations/tasks", nil)
			common.SetContextKey(c, constant.ContextKeyOriginTasks, []*model.Task{origin})
			info := &relaycommon.RelayInfo{OriginModelName: "doubao-seedance-2.5", ChannelMeta: &relaycommon.ChannelMeta{UpstreamModelName: "doubao-seedance-2.5"}}
			require.Nil(t, relay.ApplyChannelPin(c, info))
			adaptor := jsadaptor.New(loadPlugin(t))
			c.Set("task_request", map[string]any{"metadata": map[string]any{"resolution": "4k", "content": []any{
				map[string]any{"type": "draft_task", "draft_task": map[string]any{"id": "task_sample"}},
			}}})
			facts, err := adaptor.ExtractUsageFactsValidated(c, info)
			require.NoError(t, err)
			require.Equal(t, "video", facts["video_input"])
			require.Equal(t, "4k", facts["resolution"])
		})
	}
}

func TestDoubaoCallbackConfirmedUsageSettlesOnceAndPublishesPublicID(t *testing.T) {
	setup(t)
	plugin := loadPlugin(t)
	var fetches atomic.Int32
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fetches.Add(1)
		if r.Header.Get("Authorization") != "Bearer provider-secret" || r.URL.Path != "/api/v3/contents/generations/tasks/cgt-real" {
			w.WriteHeader(403)
			return
		}
		io.WriteString(w, `{"id":"cgt-real","model":"vendor-model","status":"succeeded","resolution":"4k","content":{"video_url":"https://cdn.example/result.mp4"},"usage":{"completion_tokens":2000,"total_tokens":2500}}`)
	}))
	defer provider.Close()
	service.GetTaskAdaptorFunc = func(constant.TaskPlatform) service.TaskPollingAdaptor { return jsadaptor.New(plugin) }
	ch := &model.Channel{Id: 1, Type: constant.ChannelTypeVolcEngine, Key: "provider-secret", BaseURL: &provider.URL, Status: common.ChannelStatusEnabled}
	require.NoError(t, model.DB.Create(ch).Error)
	require.NoError(t, model.DB.Create(&model.User{Id: 1, Username: "callback-test", Quota: 10000}).Error)
	expr := `u("video_input") == "video" ? tier("video", u("tokens") * 0.000002) : tier("none", u("tokens") * 0.000005)`
	task := &model.Task{TaskID: "task_public", Platform: "doubao", ChannelId: 1, UserId: 1, Status: model.TaskStatusInProgress, Quota: 1000,
		Properties: model.Properties{OriginModelName: "doubao-seedance-2.5"},
		PrivateData: model.TaskPrivateData{UpstreamTaskID: "cgt-real", BillingContext: &model.TaskBillingContext{
			TieredSnapshot: &billingexpr.BillingSnapshot{ExprString: expr, ExprVersion: 1, TaskUsageBilling: true, QuotaPerUnit: 1000000, GroupRatio: 0.6,
				UsageFacts: map[string]any{"video_input": "video", "resolution": "4k", "tokens": 1000}},
		}}, Data: []byte(`{"id":"cgt-real","status":"running"}`)}
	require.NoError(t, model.DB.Create(task).Error)
	var deliveries atomic.Int32
	received := make(chan map[string]any, 2)
	customer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		deliveries.Add(1)
		if r.Header.Get("Authorization") != "" || r.Header.Get("Cookie") != "" {
			w.WriteHeader(403)
			return
		}
		var payload map[string]any
		if common.DecodeJson(r.Body, &payload) != nil {
			w.WriteHeader(400)
			return
		}
		received <- payload
		w.WriteHeader(204)
	}))
	defer customer.Close()
	callbackURL, err := service.PrepareDoubaoCallback(task.TaskID, customer.URL)
	require.NoError(t, err)
	u, err := url.Parse(callbackURL)
	require.NoError(t, err)
	token := u.Query().Get("token")
	require.Error(t, service.ReceiveDoubaoCallback(t.Context(), task.TaskID, token, "cgt-real"), "must wait for initial settlement")
	service.ActivateDoubaoCallback(task)
	require.Error(t, service.ReceiveDoubaoCallback(t.Context(), task.TaskID, "bad-token", "cgt-real"))
	require.Error(t, service.ReceiveDoubaoCallback(t.Context(), task.TaskID, token, "another-task"))
	require.Zero(t, fetches.Load())
	// Forged status/usage in the body are ignored. The real provider reports
	// 2000 completion tokens, which settle at 2400 quota including the 0.6 ratio.
	r := gin.New()
	r.POST(service.DoubaoCallbackPath+":task_id", controller.DoubaoTaskNotification)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, u.RequestURI(), strings.NewReader(`{"id":"cgt-real","status":"failed","usage":{"completion_tokens":0}}`)))
	require.Equal(t, 204, w.Code)
	select {
	case payload := <-received:
		require.Equal(t, "task_public", payload["id"])
		require.Equal(t, "succeeded", payload["status"])
		require.Equal(t, "doubao-seedance-2.5", payload["model"])
		require.Equal(t, float64(2000), payload["usage"].(map[string]any)["completion_tokens"])
	case <-time.After(5 * time.Second):
		t.Fatal("callback not delivered")
	}
	require.Eventually(t, func() bool {
		var row model.TaskCallback
		return model.DB.Where("task_id = ?", task.TaskID).First(&row).Error == nil && row.DeliveredStatus == string(model.TaskStatusSuccess) && row.LeaseUntil == 0
	}, time.Second, 10*time.Millisecond)
	var user model.User
	require.NoError(t, model.DB.First(&user, 1).Error)
	require.Equal(t, 8600, user.Quota)
	var saved model.Task
	require.NoError(t, model.DB.First(&saved, task.ID).Error)
	require.Equal(t, 2400, saved.Quota)
	// Replayed notifications and a polling pass must neither charge nor send again.
	for range 3 {
		require.NoError(t, service.ReceiveDoubaoCallback(t.Context(), task.TaskID, token, "cgt-real"))
	}
	service.DispatchDoubaoCallbacks(context.Background())
	require.NoError(t, model.DB.First(&user, 1).Error)
	require.Equal(t, 8600, user.Quota)
	require.Equal(t, int32(1), fetches.Load())
	require.Equal(t, int32(1), deliveries.Load())
}

func TestDoubaoCallbackRetryLeaseAndPollingFallback(t *testing.T) {
	setup(t)
	var calls atomic.Int32
	customer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.WriteHeader(503)
			return
		}
		w.WriteHeader(204)
	}))
	defer customer.Close()
	task := &model.Task{TaskID: "task_retry", Platform: "doubao", Status: model.TaskStatusSuccess,
		Data: []byte(`{"id":"cgt-retry","status":"succeeded","usage":{"completion_tokens":123}}`)}
	require.NoError(t, model.DB.Create(task).Error)
	_, err := service.PrepareDoubaoCallback(task.TaskID, customer.URL)
	require.NoError(t, err)
	service.ActivateDoubaoCallback(task)
	// No vendor notification: scheduled dispatch still sends completed task data.
	service.DispatchDoubaoCallbacks(t.Context())
	var row model.TaskCallback
	require.NoError(t, model.DB.First(&row).Error)
	require.Equal(t, 1, row.Attempts)
	require.Empty(t, row.DeliveredStatus)
	service.DispatchDoubaoCallbacks(t.Context())
	require.Equal(t, int32(1), calls.Load(), "backoff must survive another dispatch")
	require.NoError(t, model.DB.Model(&row).Updates(map[string]any{"next_attempt_at": 0, "lease_until": time.Now().Unix() + 30}).Error)
	service.DispatchDoubaoCallbacks(t.Context())
	require.Equal(t, int32(1), calls.Load(), "another node owns the lease")
	require.NoError(t, model.DB.Model(&row).Update("lease_until", 0).Error)
	service.DispatchDoubaoCallbacks(t.Context())
	require.Equal(t, int32(2), calls.Load())
	require.NoError(t, model.DB.First(&row).Error)
	require.Equal(t, string(model.TaskStatusSuccess), row.DeliveredStatus)
	// Callback data and credentials must never appear in public model JSON.
	data, err := common.Marshal(row)
	require.NoError(t, err)
	require.JSONEq(t, `{}`, string(data))
	for _, target := range []string{"file:///etc/passwd", "https://user:secret@example.com/hook", "https://example.com/hook#secret"} {
		_, err := service.PrepareDoubaoCallback("invalid", target)
		require.Error(t, err)
	}
	system_setting.GetFetchSetting().EnableSSRFProtection = true
	system_setting.GetFetchSetting().AllowPrivateIp = false
	_, err = service.PrepareDoubaoCallback("private", "http://127.0.0.1/hook")
	require.Error(t, err, "callback must honor existing SSRF protections")
}
