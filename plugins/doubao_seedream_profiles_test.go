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
