package plugins_test

import (
	"maps"
	"slices"
	"testing"

	"github.com/QuantumNous/new-api/pkg/jsplugin"
	builtinplugins "github.com/QuantumNous/new-api/plugins"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newDoubaoProfilePlugin(t *testing.T) *jsplugin.LoadedPlugin {
	t.Helper()
	source, err := builtinplugins.Source("doubao")
	require.NoError(t, err)
	plugin, err := jsplugin.NewRegistry().RegisterFactory(source, jsplugin.Options{Key: "doubao"})
	require.NoError(t, err)
	return plugin
}

func TestDoubaoSeedreamProfiles(t *testing.T) {
	plugin := newDoubaoProfilePlugin(t)
	layered := []string{"images_above_1_5k", "images_up_to_1_5k", "input_images", "layer_decomposition"}
	oneTier := []string{"images_above_1_5k", "input_images"}
	for model, want := range map[string][]string{
		"doubao-seedream-5-0-pro-260628":   layered,
		"doubao-seedream-5-0-flash-260915": layered,
		"doubao-seedream-5-0-lite-260128":  oneTier,
		"doubao-seedream-5-0-260128":       oneTier,
		"doubao-seedream-4-5-251128":       oneTier,
		"doubao-seedream-4-0-250828":       []string{"images_above_1_5k", "images_up_to_1_5k", "input_images"},
	} {
		schema, examples := plugin.Meta.UsageForModel(model)
		assert.Equal(t, want, slices.Sorted(maps.Keys(schema)), model)
		assert.NotEmpty(t, examples, model)
	}
}

func TestDoubaoSeedanceAliasesExposeProfilesAndBillingRatios(t *testing.T) {
	plugin := newDoubaoProfilePlugin(t)
	profiles := map[string][]string{
		"doubao-seedance-2.0":      {"480p", "720p", "1080p", "4k"},
		"doubao-seedance-2.0-fast": {"480p", "720p"},
		"doubao-seedance-2.0-mini": {"480p", "720p"},
		"doubao-seedance-2.5":      {"480p", "720p", "1080p"},
	}
	for model, resolutions := range profiles {
		assert.Contains(t, plugin.Meta.Models, model)
		schema, examples := plugin.Meta.UsageForModel(model)
		assert.Equal(t, resolutions, schema["resolution"].Enum, model)
		assert.Contains(t, schema, "tokens", model)
		assert.Contains(t, schema, "video_input", model)
		assert.NotEmpty(t, examples, model)
	}

	aliases := []struct {
		model      string
		resolution string
		wantRatio  float64
	}{
		{model: "doubao-seedance-2.0", resolution: "1080p", wantRatio: 31.0 / 46.0},
		{model: "doubao-seedance-2.0-fast", resolution: "720p", wantRatio: 22.0 / 37.0},
		{model: "doubao-seedance-2.0-mini", resolution: "720p", wantRatio: 14.0 / 23.0},
		{model: "doubao-seedance-2.5", resolution: "1080p", wantRatio: 7.0 / 10.7},
	}
	for _, alias := range aliases {
		t.Run(alias.model, func(t *testing.T) {
			content := []any{map[string]any{
				"type":      "video_url",
				"video_url": map[string]any{"url": "https://example.com/reference.mp4"},
			}}
			value, err := plugin.Engine.Call(t.Context(), "extractUsage", map[string]any{
				"model":         alias.model,
				"upstreamModel": alias.model,
				"usagePurpose":  "billing_ratios",
				"requestBody": map[string]any{
					"model":    alias.model,
					"seconds":  5,
					"metadata": map[string]any{"resolution": alias.resolution, "content": content},
				},
			})
			require.NoError(t, err)
			ratioFacts, ok := value.(map[string]any)
			require.True(t, ok)
			ratio, ok := ratioFacts["video_input_ratio"].(float64)
			require.True(t, ok)
			assert.InDelta(t, alias.wantRatio, ratio, 1e-12)

			facts, err := plugin.Engine.Call(t.Context(), "extractUsage", map[string]any{
				"model":         alias.model,
				"upstreamModel": alias.model,
				"usagePurpose":  "facts",
				"requestBody": map[string]any{
					"model":    alias.model,
					"seconds":  5,
					"metadata": map[string]any{"resolution": alias.resolution, "content": content},
				},
			})
			require.NoError(t, err)
			usage, ok := facts.(map[string]any)
			require.True(t, ok)
			assert.Equal(t, alias.resolution, usage["resolution"])
			assert.Equal(t, "video", usage["video_input"])
			tokens, ok := usage["tokens"].(float64)
			require.True(t, ok)
			assert.Greater(t, tokens, 0.0)
		})
	}

	for _, model := range []string{"doubao-seedance-2.0-fast", "doubao-seedance-2.0-mini"} {
		_, err := plugin.Engine.Call(t.Context(), "buildSubmitRequest", map[string]any{
			"model":         model,
			"upstreamModel": model,
			"requestBody":   map[string]any{
				"model":    model,
				"prompt":   "a short video",
				"metadata": map[string]any{"resolution": "1080p"},
			},
		})
		assert.ErrorContains(t, err, "resolution must be one of 480p, 720p")
	}
}

func TestDoubaoSeedreamFlashBoundsAndBillingFacts(t *testing.T) {
	plugin := newDoubaoProfilePlugin(t)
	request := func(model string, body map[string]any) map[string]any {
		return map[string]any{"model": model, "upstreamModel": model, "requestBody": body}
	}

	_, err := plugin.Engine.Call(t.Context(), "buildSubmitRequest", request("doubao-seedream-5-0-flash-260915", map[string]any{
		"model": "doubao-seedream-5-0-flash-260915", "prompt": "a cat", "size": "4K",
	}))
	require.ErrorContains(t, err, "size must be one of 1K, 1.5K, 2K")

	value, err := plugin.Engine.Call(t.Context(), "extractUsageOnComplete", map[string]any{
		"model": "doubao-seedream-5-0-lite-260128", "upstreamModel": "doubao-seedream-5-0-lite-260128", "action": "text_to_image",
	}, map[string]any{
		"data": []any{map[string]any{"url": "https://cdn.example/1.png"}, map[string]any{"url": "https://cdn.example/2.png"}},
	})
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"images_above_1_5k": float64(2)}, value)

	value, err = plugin.Engine.Call(t.Context(), "extractUsageOnComplete", map[string]any{
		"model": "doubao-seedream-5-0-pro-260628", "upstreamModel": "doubao-seedream-5-0-pro-260628", "action": "text_to_image",
	}, map[string]any{
		"data": []any{
			map[string]any{"url": "https://cdn.example/1.png", "size": "1024x1024"},
			map[string]any{"url": "https://cdn.example/2.png", "size": "2048x2048"},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"images_up_to_1_5k": float64(1), "images_above_1_5k": float64(1)}, value)
}

func TestDoubaoNativeVideoBodyPreservesProviderFields(t *testing.T) {
	plugin := newDoubaoProfilePlugin(t)
	original := map[string]any{
		"model":    "doubao-seedance-2.0",
		"content": []any{
			map[string]any{"type": "text", "text": "keep this exact item"},
			map[string]any{"type": "image_url", "image_url": map[string]any{"url": "https://example.com/input.png"}},
		},
		"duration": 5,
		"ratio":    "16:9",
		"x_vendor": map[string]any{"preserve": true},
	}
	value, err := plugin.Engine.CallPath(t.Context(), "native", []string{"createTask"}, map[string]any{
		"body": map[string]any{"kind": "json", "value": original},
	})
	require.NoError(t, err)
	intent, ok := value.(map[string]any)
	require.True(t, ok)
	require.Equal(t, "doubao-seedance-2.0", intent["model"])
	requestBody, ok := intent["requestBody"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, float64(5), requestBody["seconds"])
	nativeBody, ok := requestBody["__doubao_native_body"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "16:9", nativeBody["ratio"])
	require.Equal(t, map[string]any{"preserve": true}, nativeBody["x_vendor"])

	value, err = plugin.Engine.CallPath(t.Context(), "buildSubmitRequest", map[string]any{
		"model":         "doubao-seedance-2.0",
		"action":        intent["action"],
		"upstreamModel": "ark-video-id",
		"requestBody":   requestBody,
		"apiKey":        "secret",
		"baseUrl":       "https://ark.example",
	})
	require.NoError(t, err)
	descriptor, ok := value.(map[string]any)
	require.True(t, ok)
	body, ok := descriptor["body"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "ark-video-id", body["model"])
	assert.Equal(t, "image_to_video", descriptor["action"])
	assert.Equal(t, "16:9", body["ratio"])
	assert.Equal(t, map[string]any{"preserve": true}, body["x_vendor"])
	content, ok := body["content"].([]any)
	require.True(t, ok)
	require.Len(t, content, 2)
	assert.Equal(t, "text", content[0].(map[string]any)["type"])
	assert.Equal(t, "image_url", content[1].(map[string]any)["type"])
}

func TestDoubaoNativeVideoDurationIsBounded(t *testing.T) {
	plugin := newDoubaoProfilePlugin(t)
	_, err := plugin.Engine.CallPath(t.Context(), "native", []string{"createTask"}, map[string]any{
		"body": map[string]any{"kind": "json", "value": map[string]any{
			"model": "doubao-seedance-2.0", "content": []any{map[string]any{"type": "text", "text": "a cat"}}, "duration": 3601,
		}},
	})
	require.ErrorContains(t, err, "duration must be between 1 and 3600")
}
