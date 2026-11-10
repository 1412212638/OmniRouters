package plugins_test

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	builtinplugins "github.com/QuantumNous/new-api/plugins"
	"github.com/QuantumNous/new-api/pkg/jsplugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAstraFlowTaskPlugin(t *testing.T) {
	source, err := builtinplugins.Source("astraflow")
	require.NoError(t, err)
	registry := jsplugin.NewRegistry()
	plugin, err := registry.RegisterFactory(source, jsplugin.Options{Key: "astraflow"})
	require.NoError(t, err)

	t.Run("declares native routes and usage facts", func(t *testing.T) {
		interaction, found := registry.Generation().LookupDeclaredRoute("POST", "/v1beta/models/interactions")
		require.True(t, found)
		assert.Equal(t, []string{"gemini-omni-1.2-flash"}, interaction.Route.Models)
		query, found := registry.Generation().LookupDeclaredRoute("GET", "/v1beta/models/interactions/:id")
		require.True(t, found)
		assert.Equal(t, "id", query.Route.TaskIDParam)
		omniSchema, _ := plugin.Meta.UsageForModel("gemini-omni-1.2-flash")
		assert.Len(t, omniSchema, 3)
		assert.Equal(t, "token", omniSchema["video_output_tokens"].Unit)
		enhancementSchema, _ := plugin.Meta.UsageForModel("ai-mediakit-video-enhancement-std")
		assert.Len(t, enhancementSchema, 3)
		assert.Equal(t, "second", enhancementSchema["seconds"].Unit)
		binding, found := registry.Generation().LookupDeclaredRoute("POST", "/v1/tasks/submit")
		require.True(t, found)
		assert.Same(t, plugin, binding.Plugin)
		queryBinding, found := registry.Generation().LookupDeclaredRoute("GET", "/v1/tasks/status")
		require.True(t, found)
		assert.True(t, queryBinding.Route.TaskIDQuery)
		assert.Equal(t, "task_id", queryBinding.Route.TaskIDParam)
		value, callErr := plugin.Engine.Call(t.Context(), "extractUsage", map[string]any{
			"requestBody": map[string]any{
				"model": "ai-mediakit-video-enhancement-std",
				"input": map[string]any{"video_url": "https://cdn.example/input.mp4"},
				"parameters": map[string]any{"resolution": "1080p", "fps": 30},
			},
		})
		require.NoError(t, callErr)
		encoded, marshalErr := common.Marshal(value)
		require.NoError(t, marshalErr)
		var facts map[string]any
		require.NoError(t, common.Unmarshal(encoded, &facts))
		assert.Equal(t, map[string]any{"seconds": float64(10), "resolution": "1080p", "fps": "30"}, facts)
		_, callErr = plugin.Engine.Call(t.Context(), "extractUsage", map[string]any{
			"requestBody": map[string]any{
				"model":      "ai-mediakit-video-enhancement-std",
				"input":      map[string]any{"video_url": "https://cdn.example/input.mp4"},
				"parameters": map[string]any{"resolution": "1080p", "fps": 21},
			},
		})
		require.Error(t, callErr)
	})

	t.Run("builds native submit and query requests", func(t *testing.T) {
		value, callErr := plugin.Engine.Call(t.Context(), "buildSubmitRequest", map[string]any{
			"baseUrl": "https://api.modelverse.cn",
			"apiKey": "fixture-key",
			"upstreamModel": "ai-mediakit-video-enhancement-std",
			"requestBody": map[string]any{
				"model": "ai-mediakit-video-enhancement-std",
				"input": map[string]any{"video_url": "https://cdn.example/input.mp4"},
				"parameters": map[string]any{"resolution": "1080p", "fps": 30},
			},
		})
		require.NoError(t, callErr)
		encoded, marshalErr := common.Marshal(value)
		require.NoError(t, marshalErr)
		var request map[string]any
		require.NoError(t, common.Unmarshal(encoded, &request))
		assert.Equal(t, "https://api.modelverse.cn/v1/tasks/submit", request["url"])
		body := request["body"].(map[string]any)
		assert.Equal(t, map[string]any{"model": "ai-mediakit-video-enhancement-std", "input": map[string]any{"video_url": "https://cdn.example/input.mp4"}, "parameters": map[string]any{"resolution": "1080p", "fps": float64(30)}}, body)

		query, callErr := plugin.Engine.Call(t.Context(), "buildQueryRequest", map[string]any{
			"baseUrl": "https://api.modelverse.cn",
			"apiKey": "fixture-key",
			"taskId": "amk-tool-enhance-video-1",
		})
		require.NoError(t, callErr)
		encoded, marshalErr = common.Marshal(query)
		require.NoError(t, marshalErr)
		var queryRequest map[string]any
		require.NoError(t, common.Unmarshal(encoded, &queryRequest))
		assert.Equal(t, "https://api.modelverse.cn/v1/tasks/status?task_id=amk-tool-enhance-video-1", queryRequest["url"])
	})

	t.Run("parses successful status and completion usage", func(t *testing.T) {
		value, callErr := plugin.Engine.Call(t.Context(), "parseTaskResult", map[string]any{}, map[string]any{
			"output": map[string]any{
				"task_status": "Success",
				"urls":        []any{"https://cdn.example/output.mp4"},
			},
		})
		require.NoError(t, callErr)
		encoded, marshalErr := common.Marshal(value)
		require.NoError(t, marshalErr)
		var result map[string]any
		require.NoError(t, common.Unmarshal(encoded, &result))
		assert.Equal(t, "SUCCESS", result["status"])
		assert.Equal(t, "https://cdn.example/output.mp4", result["remoteUrl"])

		facts, callErr := plugin.Engine.Call(t.Context(), "extractUsageOnComplete", map[string]any{}, map[string]any{}, map[string]any{
			"output": map[string]any{"task_status": "Success"},
			"usage": map[string]any{"duration": 15, "output_video_fps": 30, "size": "1080p"},
		})
		require.NoError(t, callErr)
		encoded, marshalErr = common.Marshal(facts)
		require.NoError(t, marshalErr)
		var completion map[string]any
		require.NoError(t, common.Unmarshal(encoded, &completion))
		assert.Equal(t, map[string]any{"seconds": float64(15), "resolution": "1080p", "fps": "30"}, completion)
	})

	t.Run("splits Omni token modalities without total-output double billing", func(t *testing.T) {
		value, callErr := plugin.Engine.Call(t.Context(), "extractUsageOnComplete",
			map[string]any{"model": "gemini-omni-1.2-flash", "action": "omni_video"}, map[string]any{},
			map[string]any{"usage": map[string]any{
				"total_input_tokens": 19, "total_output_tokens": 57930, "total_tokens": 58295,
				"output_tokens_by_modality": []any{map[string]any{"modality": "text", "tokens": 10}, map[string]any{"modality": "video", "tokens": 57920}},
			}})
		require.NoError(t, callErr)
		encoded, marshalErr := common.Marshal(value)
		require.NoError(t, marshalErr)
		var facts map[string]any
		require.NoError(t, common.Unmarshal(encoded, &facts))
		assert.Equal(t, float64(19), facts["input_tokens"])
		assert.Equal(t, float64(10), facts["text_output_tokens"])
		assert.Equal(t, float64(57920), facts["video_output_tokens"])
	})
}
