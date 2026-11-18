package jsplugin

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	pluginruntime "github.com/QuantumNous/new-api/pkg/jsplugin"
	builtinplugins "github.com/QuantumNous/new-api/plugins"
	"github.com/QuantumNous/new-api/relay/channel"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/stretchr/testify/require"
)

func hailuoContentAdaptor(t *testing.T, baseURL string) *TaskAdaptor {
	t.Helper()
	source, err := builtinplugins.Source("hailuo")
	require.NoError(t, err)
	plugin, err := pluginruntime.NewRegistry().RegisterFactory(source, pluginruntime.Options{Key: "hailuo"})
	require.NoError(t, err)
	a := New(plugin)
	a.Init(&relaycommon.RelayInfo{ChannelMeta: &relaycommon.ChannelMeta{
		ChannelType: constant.ChannelTypeMiniMax, ChannelBaseUrl: baseURL, ApiKey: "task-key",
	}})
	return a
}

func TestHailuoFileLookupSeparatesCredentialsFromVideo(t *testing.T) {
	var hits int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		require.Equal(t, http.MethodGet, r.Method, "signed GET downloads require GET metadata even for a HEAD client")
		require.Equal(t, "/prefix/v1/files/retrieve", r.URL.Path)
		require.Equal(t, "file/with?reserved", r.URL.Query().Get("file_id"))
		require.Equal(t, "Bearer task-key", r.Header.Get("Authorization"))
		require.Equal(t, "application/json", r.Header.Get("Accept"))
		_, _ = fmt.Fprint(w, `{"base_resp":{"status_code":0},"file":{"download_url":"https://cdn.example/video.mp4?signature=signed"}}`)
	}))
	defer server.Close()
	a := hailuoContentAdaptor(t, server.URL+"/prefix/")
	task := &model.Task{TaskID: "public-id", Status: model.TaskStatusSuccess, Quota: 123, Data: []byte(`{"file_id":"file/with?reserved"}`)}
	before, err := common.Marshal(task)
	require.NoError(t, err)
	for _, method := range []string{http.MethodGet, http.MethodHead} {
		result, err := a.BuildContentRequest(task, "video", channel.TaskArtifactClientRequest{Method: method})
		require.NoError(t, err)
		require.Equal(t, "https://cdn.example/video.mp4?signature=signed", result.URL)
		require.Equal(t, http.MethodGet, result.Method)
		require.True(t, result.Credentialless)
		require.Empty(t, result.Headers)
		require.Empty(t, result.Body)
	}
	after, err := common.Marshal(task)
	require.NoError(t, err)
	require.Equal(t, string(before), string(after), "media retrieval must not change task or charge data")
	require.Equal(t, 123, task.Quota)
	require.Equal(t, 2, hits)
}

func TestHailuoFileLookupRejectsInvalidResponses(t *testing.T) {
	for _, test := range []struct {
		name, body string
		status     int
	}{
		{"HTTP failure", `{}`, http.StatusUnauthorized},
		{"provider failure", `{"base_resp":{"status_code":2013},"file":{"download_url":"https://cdn.example/v.mp4"}}`, http.StatusOK},
		{"missing success code", `{"file":{"download_url":"https://cdn.example/v.mp4"}}`, http.StatusOK},
		{"null success code", `{"base_resp":{"status_code":null},"file":{"download_url":"https://cdn.example/v.mp4"}}`, http.StatusOK},
		{"invalid JSON", `{"file":`, http.StatusOK},
		{"oversized JSON", strings.Repeat(" ", maxTaskPluginPersistedJSONBytes+1), http.StatusOK},
		{"missing URL", `{"base_resp":{"status_code":0},"file":{}}`, http.StatusOK},
		{"relative URL", `{"base_resp":{"status_code":0},"file":{"download_url":"/video.mp4"}}`, http.StatusOK},
		{"unsafe scheme", `{"base_resp":{"status_code":0},"file":{"download_url":"file:///private"}}`, http.StatusOK},
		{"credentials in URL", `{"base_resp":{"status_code":0},"file":{"download_url":"https://key@cdn.example/video.mp4"}}`, http.StatusOK},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(test.status)
				_, _ = fmt.Fprint(w, test.body)
			}))
			defer server.Close()
			a := hailuoContentAdaptor(t, server.URL)
			result, err := a.BuildContentRequest(&model.Task{Data: []byte(`{"file_id":"file-id"}`)}, "video", channel.TaskArtifactClientRequest{Method: http.MethodGet})
			require.Error(t, err)
			require.Nil(t, result)
			require.NotContains(t, err.Error(), "task-key")
		})
	}
}

func TestHailuoFileLookupDoesNotFollowRedirects(t *testing.T) {
	var targetHits int
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { targetHits++; w.WriteHeader(http.StatusOK) }))
	defer target.Close()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, target.URL, http.StatusFound) }))
	defer server.Close()
	a := hailuoContentAdaptor(t, server.URL)
	_, err := a.BuildContentRequest(&model.Task{Data: []byte(`{"file_id":"file-id"}`)}, "video", channel.TaskArtifactClientRequest{Method: http.MethodGet})
	require.ErrorContains(t, err, "HTTP 302")
	require.Zero(t, targetHits)
}

func TestHailuoFileLookupScopeAndExistingH3(t *testing.T) {
	a := hailuoContentAdaptor(t, "https://vendor.example")
	base := requestDescriptor{Method: http.MethodGet, URL: "https://vendor.example/v1/files/retrieve?file_id=one"}
	for _, test := range []struct {
		name, value string
	}{
		{"other host", "https://other.example/v1/files/retrieve?file_id=one"},
		{"other endpoint", "https://vendor.example/private?file_id=one"},
		{"duplicate id", "https://vendor.example/v1/files/retrieve?file_id=one&file_id=two"},
		{"extra query", "https://vendor.example/v1/files/retrieve?file_id=one&url=https://other.example"},
	} {
		t.Run(test.name, func(t *testing.T) {
			descriptor := base
			descriptor.URL = test.value
			_, err := a.resolveHailuoFileContent("video", channel.TaskArtifactClientRequest{Method: http.MethodGet}, descriptor)
			require.Error(t, err)
		})
	}
	_, err := a.BuildContentRequest(&model.Task{Data: []byte(`{"file_id":"one"}`)}, "video", channel.TaskArtifactClientRequest{Method: http.MethodPost})
	require.Error(t, err)
	result, err := a.BuildContentRequest(&model.Task{Data: []byte(`{"task":{"content":{"url":"https://cdn.example/h3.mp4"}}}`)}, "video", channel.TaskArtifactClientRequest{Method: http.MethodHead})
	require.NoError(t, err)
	require.Equal(t, "https://cdn.example/h3.mp4", result.URL)
	require.Equal(t, http.MethodHead, result.Method)
	require.True(t, result.Credentialless)
	require.Empty(t, result.Headers)
}

func TestHailuoFileLookupLeavesOtherPluginsUnchanged(t *testing.T) {
	plugin, err := pluginruntime.NewRegistry().RegisterFactory(`
export const meta = {apiVersion:1,key:"other-content",name:"fixture",version:"1.0.0",author:{name:"fixture"},models:["fixture"],fetchMode:"per_task"};
export function buildSubmitRequest() {return {};}
export function parseSubmitResponse() {return {taskId:"fixture"};}
export function buildQueryRequest() {return {};}
export function parseTaskResult() {return {status:"SUCCESS"};}
export function listArtifacts() {return [{key:"video",type:"video"}];}
export function buildContentRequest() {
  return {url:"https://vendor.example/unrelated",method:"GET",headers:{Accept:"application/json"},responseType:"json"};
}`, pluginruntime.Options{Key: "other-content"})
	require.NoError(t, err)
	a := New(plugin)
	a.Init(&relaycommon.RelayInfo{ChannelMeta: &relaycommon.ChannelMeta{
		ChannelType: constant.ChannelTypeOpenAI, ChannelBaseUrl: "https://vendor.example", ApiKey: "task-key",
	}})
	result, err := a.BuildContentRequest(&model.Task{}, "video", channel.TaskArtifactClientRequest{Method: http.MethodGet})
	require.NoError(t, err)
	require.Equal(t, "https://vendor.example/unrelated", result.URL)
	require.Equal(t, http.MethodGet, result.Method)
	require.Equal(t, map[string]string{"Accept": "application/json"}, result.Headers)
	require.False(t, result.Credentialless)
}

func TestCredentiallessArtifactDoesNotRequireChannelInitialization(t *testing.T) {
	plugin, err := pluginruntime.NewRegistry().RegisterFactory(`
export const meta = {apiVersion:1,key:"credentialless-content",name:"fixture",version:"1.0.0",author:{name:"fixture"},models:["fixture"],fetchMode:"per_task"};
export function buildSubmitRequest() {return {};}
export function parseSubmitResponse() {return {taskId:"fixture"};}
export function buildQueryRequest() {return {};}
export function parseTaskResult() {return {status:"SUCCESS"};}
export function listArtifacts() {return [{key:"audio",type:"audio"}];}
export function buildContentRequest(ctx) {return {url:"https://cdn.example/audio.mp3",method:ctx.clientRequest.method,credentialless:true};}
`, pluginruntime.Options{Key: "credentialless-content"})
	require.NoError(t, err)
	result, err := New(plugin).BuildContentRequest(&model.Task{}, "audio", channel.TaskArtifactClientRequest{Method: http.MethodGet})
	require.NoError(t, err)
	require.Equal(t, "https://cdn.example/audio.mp3", result.URL)
	require.Equal(t, http.MethodGet, result.Method)
	require.True(t, result.Credentialless)
}
