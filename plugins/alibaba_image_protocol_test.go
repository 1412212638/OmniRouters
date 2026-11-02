package plugins_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAlibabaOpenAIImageProtocol(t *testing.T) {
	plugin := newAlibabaPlugin(t)
	decode := func(body map[string]any) map[string]any {
		t.Helper()
		value, err := plugin.Engine.CallPath(t.Context(), "protocols", []string{"openai_image", "decodeRequest"}, map[string]any{
			"model": "wan2.7-image", "operation": "generate",
			"body": map[string]any{"kind": "json", "value": body},
		})
		require.NoError(t, err)
		return alibabaObject(t, value)["requestBody"].(map[string]any)
	}

	t.Run("preserves explicit parameters and sends the requested image count", func(t *testing.T) {
		request := decode(map[string]any{"prompt": "a cat", "n": 2, "seed": 0, "watermark": false})
		body, facts, _ := submitAlibabaRequest(t, plugin, "wan2.7-image", request)
		parameters := body["parameters"].(map[string]any)
		assert.Equal(t, float64(2), parameters["n"])
		assert.Equal(t, float64(0), parameters["seed"])
		assert.Equal(t, false, parameters["watermark"])
		assert.Equal(t, float64(2), facts["image_count"])
	})

	t.Run("passthrough counts are bounded before submission", func(t *testing.T) {
		request := decode(map[string]any{"prompt": "a cat", "parameters": map[string]any{"n": 129}})
		_, err := plugin.Engine.Call(t.Context(), "extractUsage", map[string]any{
			"model": "wan2.7-image", "upstreamModel": "wan2.7-image", "usagePurpose": "facts", "requestBody": request,
		})
		require.ErrorContains(t, err, "n must be an integer between")
	})

	t.Run("Wan3 supports the upstream ultrawide ratio", func(t *testing.T) {
		body, _, _ := submitAlibabaRequest(t, plugin, "wan3.0-video", map[string]any{
			"model": "wan3.0-video", "prompt": "a cat", "ratio": "21:9", "duration": 5,
		})
		assert.Equal(t, "21:9", body["parameters"].(map[string]any)["ratio"])
	})
}
