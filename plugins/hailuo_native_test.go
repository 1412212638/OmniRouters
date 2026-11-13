package plugins_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/pkg/jsplugin"
	builtinplugins "github.com/QuantumNous/new-api/plugins"
	"github.com/QuantumNous/new-api/service"
	"github.com/stretchr/testify/require"
)

func h3NativePlugin(t *testing.T) *jsplugin.LoadedPlugin {
	t.Helper()
	source, err := builtinplugins.Source("hailuo")
	require.NoError(t, err)
	plugin, err := jsplugin.NewRegistry().RegisterFactory(source, jsplugin.Options{Key: "hailuo"})
	require.NoError(t, err)
	return plugin
}

func h3Call(t *testing.T, plugin *jsplugin.LoadedPlugin, export string, path []string, args ...any) map[string]any {
	t.Helper()
	value, err := plugin.Engine.CallPath(t.Context(), export, path, args...)
	require.NoError(t, err)
	encoded, err := common.Marshal(value)
	require.NoError(t, err)
	var result map[string]any
	require.NoError(t, common.Unmarshal(encoded, &result))
	return result
}

func TestH3NativeSubmitAndBillingFacts(t *testing.T) {
	plugin := h3NativePlugin(t)
	require.Equal(t, "/v2/video_generation", plugin.Meta.Routes[0].Path)
	require.Equal(t, "/v2/query/video_generation/:task_id", plugin.Meta.Routes[1].Path)
	for _, name := range plugin.Meta.Routes[0].Models {
		t.Run(name, func(t *testing.T) {
			body := map[string]any{
				"model": name, "duration": 5, "resolution": "768P", "ratio": "16:9",
				"aigc_watermark": false, "seed": 0,
				"callback_url": "https://customer.example/callback",
				"content": []any{
					map[string]any{"type": "text", "text": "a cat"},
					map[string]any{"type": "image_url", "role": "reference_image", "image_url": map[string]any{"url": "https://cdn.example/ref.png"}},
					map[string]any{"type": "video_url", "role": "reference_video", "video_url": map[string]any{"url": "https://cdn.example/ref.mp4"}},
				},
			}
			intent := h3Call(t, plugin, "native", []string{"createH3Task"}, map[string]any{"body": map[string]any{"kind": "json", "value": body}})
			require.Equal(t, name, intent["model"])
			require.Equal(t, "image_to_video", intent["action"])
			ctx := map[string]any{
				"model": name, "upstreamModel": strings.ToUpper(name), "baseUrl": "https://vendor.example", "apiKey": "fixture",
				"requestBody": intent["requestBody"], "usagePurpose": "facts",
			}
			descriptor := h3Call(t, plugin, "buildSubmitRequest", nil, ctx)
			require.Equal(t, "https://vendor.example/v2/video_generation", descriptor["url"])
			outbound := descriptor["body"].(map[string]any)
			require.Equal(t, strings.ToUpper(name), outbound["model"])
			require.Equal(t, false, outbound["aigc_watermark"])
			require.Equal(t, float64(0), outbound["seed"])
			require.Equal(t, float64(5), outbound["duration"])
			require.Equal(t, body["callback_url"], outbound["callback_url"])
			require.Len(t, outbound["content"], 3)
			require.NotContains(t, outbound, "__hailuo_h3_native_body")
			query := h3Call(t, plugin, "buildQueryRequest", nil, map[string]any{
				"model": name, "upstreamModel": "private-provider-alias", "taskId": "upstream/id", "baseUrl": "https://vendor.example",
			})
			require.Equal(t, "https://vendor.example/v2/query/video_generation/upstream%2Fid", query["url"])
			facts := h3Call(t, plugin, "extractUsage", nil, ctx)
			require.Equal(t, map[string]any{
				"seconds": float64(5), "resolution": "768P", "input_images": float64(1), "input_video_seconds": float64(15),
			}, facts)
			settled := h3Call(t, plugin, "extractUsageOnComplete", nil, ctx, nil, map[string]any{
				"task": map[string]any{"model": strings.ToUpper(name), "resolution": "768P", "usage": map[string]any{
					"output_seconds": 6.5, "input_image_count": 1, "input_seconds": 3.2,
				}},
			})
			require.Equal(t, map[string]any{
				"seconds": 6.5, "resolution": "768P", "input_images": float64(1), "input_video_seconds": 3.2,
			}, settled)
		})
	}
}

func TestH3NativeRejectsUnboundedOrInvalidRequests(t *testing.T) {
	plugin := h3NativePlugin(t)
	for _, duration := range []any{-1, 0, 3, 16, 3601, 5.5, "18446744073686646784"} {
		t.Run(fmt.Sprint(duration), func(t *testing.T) {
			_, err := plugin.Engine.CallPath(t.Context(), "native", []string{"createH3Task"}, map[string]any{"body": map[string]any{"kind": "json", "value": map[string]any{
				"model": "minimax-h3", "duration": duration, "content": []any{map[string]any{"type": "text", "text": "a cat"}},
			}}})
			require.ErrorContains(t, err, "duration")
		})
	}
	for _, body := range []map[string]any{
		{"model": "MiniMax-Hailuo-2.3", "content": []any{}},
		{"model": "minimax-h3", "content": "invalid"},
		{"model": "minimax-h3", "content": []any{}},
		{"model": "minimax-h3", "resolution": "4K", "content": []any{map[string]any{"type": "text", "text": "cat"}}},
	} {
		_, err := plugin.Engine.CallPath(t.Context(), "native", []string{"createH3Task"}, map[string]any{"body": map[string]any{"kind": "json", "value": body}})
		require.Error(t, err)
	}
	content := []any{map[string]any{"type": "text", "text": "a cat"}}
	for range 10 {
		content = append(content, map[string]any{"type": "image_url", "role": "reference_image", "image_url": map[string]any{"url": "https://cdn.example/ref.png"}})
	}
	_, err := plugin.Engine.CallPath(t.Context(), "native", []string{"createH3Task"}, map[string]any{"body": map[string]any{"kind": "json", "value": map[string]any{"model": "minimax-h3", "content": content}}})
	require.ErrorContains(t, err, "at most 9")
	// A forged native marker through another protocol must not bypass validation.
	for _, hook := range []string{"buildSubmitRequest", "extractUsage"} {
		_, err := plugin.Engine.Call(t.Context(), hook, map[string]any{
			"requestBody": map[string]any{"__hailuo_h3_native_body": map[string]any{
				"model": "minimax-h3", "duration": -1, "content": []any{map[string]any{"type": "text", "text": "a cat"}},
			}},
		})
		require.ErrorContains(t, err, "duration")
	}
}

func TestH3NativePublicTaskResponse(t *testing.T) {
	plugin := h3NativePlugin(t)
	data, err := common.Marshal(map[string]any{"task": map[string]any{
		"id": "upstream-private", "model": "UPSTREAM-INTERNAL-MODEL", "upstream_model": "hidden", "channel_id": 123,
		"content": map[string]any{"url": "https://cdn.example/result.mp4"},
		"usage":   map[string]any{"input_image_count": 2, "output_seconds": 5}, "duration": 5,
	}})
	require.NoError(t, err)
	task := &model.Task{TaskID: "task_public", Platform: "hailuo", Status: model.TaskStatusSuccess, CreatedAt: 100, UpdatedAt: 110,
		Data: data, Properties: model.Properties{OriginModelName: "minimax-h3", UpstreamModelName: "UPSTREAM-INTERNAL-MODEL"}}
	task.PrivateData.UpstreamTaskID = "upstream-private"
	view, err := service.BuildTaskPluginView(task)
	require.NoError(t, err)
	encoded, err := common.Marshal(view)
	require.NoError(t, err)
	var publicView map[string]any
	require.NoError(t, common.Unmarshal(encoded, &publicView))
	created := h3Call(t, plugin, "native", []string{"h3TaskCreated"}, nil, publicView)
	require.Equal(t, map[string]any{"task_id": "task_public"}, created)
	result := h3Call(t, plugin, "native", []string{"h3TaskStatus"}, nil, publicView)
	output := result["task"].(map[string]any)
	require.Equal(t, "task_public", output["id"])
	require.Equal(t, "minimax-h3", output["model"])
	require.Equal(t, "succeeded", output["status"])
	require.Equal(t, map[string]any{"url": "https://cdn.example/result.mp4"}, output["content"])
	require.Equal(t, map[string]any{"input_image_count": float64(2), "output_seconds": float64(5)}, output["usage"])
	require.NotContains(t, output, "upstream_model")
	require.NotContains(t, output, "channel_id")
	for _, state := range []struct{ host, wire string }{{"SUBMITTED", "queued"}, {"IN_PROGRESS", "running"}, {"FAILURE", "failed"}} {
		publicView["status"], publicView["data"] = state.host, map[string]any{"task_id": "task_public"}
		result = h3Call(t, plugin, "native", []string{"h3TaskStatus"}, nil, publicView)
		require.Equal(t, state.wire, result["task"].(map[string]any)["status"])
	}
}
