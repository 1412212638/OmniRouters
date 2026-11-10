package jsplugin

import (
	"encoding/base64"
	"net/http"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	pluginruntime "github.com/QuantumNous/new-api/pkg/jsplugin"
	"github.com/QuantumNous/new-api/relay/channel"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/stretchr/testify/require"
)

func TestOmniResponseLimitIsScoped(t *testing.T) {
	adaptor := &TaskAdaptor{plugin: &pluginruntime.LoadedPlugin{Meta: pluginruntime.Meta{Key: "astraflow"}}, info: &relaycommon.RelayInfo{TaskRelayInfo: &relaycommon.TaskRelayInfo{Action: "omni_video"}}}
	require.Equal(t, maxAstraFlowOmniJSONBytes, adaptor.submitJSONLimit())
	adaptor.info.Action = "video_enhancement"
	require.Equal(t, maxTaskPluginPersistedJSONBytes, adaptor.submitJSONLimit())
	adaptor.info.Action = "omni_video"
	adaptor.plugin.Meta.Key = "other"
	require.Equal(t, maxTaskPluginPersistedJSONBytes, adaptor.submitJSONLimit())
}

func TestOmniInlineContentNeedsNoUpstreamCredentials(t *testing.T) {
	registry := pluginruntime.NewRegistry()
	plugin, err := registry.RegisterFactory(`
export const meta = {apiVersion:1,key:"astraflow",name:"fixture",version:"1.1.0",author:{name:"fixture"},models:["fixture"],fetchMode:"per_task"};
export function buildSubmitRequest() {return {};}
export function parseSubmitResponse() {return {taskId:"fixture"};}
export function buildQueryRequest() {return {};}
export function parseTaskResult() {return {status:"SUCCESS"};}
export function listArtifacts() {return [{key:"video",type:"video"}];}
export function buildContentRequest(ctx) {return ctx.data;}
`, pluginruntime.Options{Key: "astraflow"})
	require.NoError(t, err)
	adaptor := New(plugin)
	for _, test := range []struct {
		name     string
		payload  string
		mimeType string
		valid    bool
	}{
		{name: "mp4", payload: base64.StdEncoding.EncodeToString([]byte("video-bytes")), mimeType: "video/mp4", valid: true},
		{name: "malformed Base64", payload: "%%%", mimeType: "video/mp4"},
		{name: "active content", payload: "eA==", mimeType: "text/html"},
		{name: "empty content", mimeType: "video/mp4"},
	} {
		t.Run(test.name, func(t *testing.T) {
			data, marshalErr := common.Marshal(map[string]any{"inlineBase64": test.payload, "mimeType": test.mimeType})
			require.NoError(t, marshalErr)
			task := &model.Task{Action: "omni_video", Data: data}
			result, contentErr := adaptor.BuildContentRequest(task, "video", channel.TaskArtifactClientRequest{Method: http.MethodGet})
			if test.valid {
				require.NoError(t, contentErr)
				require.Equal(t, []byte("video-bytes"), result.InlineData)
				require.Empty(t, result.URL)
				require.Empty(t, result.Headers)
			} else {
				require.Error(t, contentErr)
			}
		})
	}
}
