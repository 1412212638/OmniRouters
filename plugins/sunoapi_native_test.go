package plugins_test

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/pkg/jsplugin"
	builtinplugins "github.com/QuantumNous/new-api/plugins"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSunoNativeSubmitInfersModelAndUsesLowercaseUpstreamPath(t *testing.T) {
	source, err := builtinplugins.Source("sunoapi")
	require.NoError(t, err)
	plugin, err := jsplugin.NewRegistry().RegisterFactory(source, jsplugin.Options{Key: "sunoapi"})
	require.NoError(t, err)

	decodedValue, err := plugin.Engine.CallMember(t.Context(), "native", "decodeSubmit", map[string]any{
		"params": map[string]any{"action": "music"},
		"body": map[string]any{"kind": "json", "value": map[string]any{
			"model": "suno_music", "prompt": "a calm instrumental",
		}},
	})
	require.NoError(t, err)
	decodedBytes, err := common.Marshal(decodedValue)
	require.NoError(t, err)
	var decoded map[string]any
	require.NoError(t, common.Unmarshal(decodedBytes, &decoded))
	require.Equal(t, "suno_music", decoded["model"])
	require.Equal(t, "MUSIC", decoded["action"])
	requestBody := decoded["requestBody"].(map[string]any)
	require.NotContains(t, requestBody, "model")

	descriptorValue, err := plugin.Engine.Call(t.Context(), "buildSubmitRequest", map[string]any{
		"baseUrl":        "https://suno.example",
		"apiKey":         "test-key",
		"params":         map[string]any{"action": "music"},
		"action":         decoded["action"],
		"requestBody":    requestBody,
		"requestHeaders": map[string]any{"Content-Type": "application/json"},
	})
	require.NoError(t, err)
	descriptorBytes, err := common.Marshal(descriptorValue)
	require.NoError(t, err)
	var descriptor map[string]any
	require.NoError(t, common.Unmarshal(descriptorBytes, &descriptor))
	require.Equal(t, "https://suno.example/suno/submit/music", descriptor["url"])
	require.NotContains(t, descriptor["body"].(map[string]any), "model")
}

func TestSunoUsageAndArtifactsHandleProviderTaskEnvelope(t *testing.T) {
	source, err := builtinplugins.Source("sunoapi")
	require.NoError(t, err)
	plugin, err := jsplugin.NewRegistry().RegisterFactory(source, jsplugin.Options{Key: "sunoapi"})
	require.NoError(t, err)

	song := map[string]any{
		"id":        "clip-1",
		"audio_url": "https://cdn.example/clip-1.mp3",
		"image_url": "https://cdn.example/clip-1.jpg",
	}
	song2 := map[string]any{
		"id":        "clip-2",
		"audio_url": "https://cdn.example/clip-2.mp3",
		"image_url": "https://cdn.example/clip-2.jpg",
	}
	song3 := map[string]any{
		"id":        "clip-3",
		"audio_url": "https://cdn.example/clip-3.mp3",
		"image_url": "https://cdn.example/clip-3.jpg",
	}
	providerTask := map[string]any{"created_at": 1, "data": []any{song, song2, song3}, "status": "SUCCESS"}

	artifactsValue, err := plugin.Engine.Call(t.Context(), "listArtifacts", map[string]any{
		"status": "SUCCESS",
		"data":   providerTask,
	})
	require.NoError(t, err)
	artifactsBytes, err := common.Marshal(artifactsValue)
	require.NoError(t, err)
	var artifacts []map[string]any
	require.NoError(t, common.Unmarshal(artifactsBytes, &artifacts))
	require.Len(t, artifacts, 6)
	assert.Equal(t, "audio", artifacts[0]["type"])

	contentValue, err := plugin.Engine.Call(t.Context(), "buildContentRequest", map[string]any{
		"data":        providerTask,
		"artifactKey": artifacts[0]["key"],
		"clientRequest": map[string]any{
			"method": "GET",
		},
	})
	require.NoError(t, err)
	contentBytes, err := common.Marshal(contentValue)
	require.NoError(t, err)
	var content map[string]any
	require.NoError(t, common.Unmarshal(contentBytes, &content))
	require.Equal(t, "https://cdn.example/clip-1.mp3", content["url"])
	require.True(t, content["credentialless"].(bool))

	usageEstimateValue, err := plugin.Engine.Call(t.Context(), "extractUsage", map[string]any{
		"params":        map[string]any{"action": "music"},
		"upstreamModel": "suno_music",
	})
	require.NoError(t, err)
	usageEstimateBytes, err := common.Marshal(usageEstimateValue)
	require.NoError(t, err)
	var usageEstimate map[string]any
	require.NoError(t, common.Unmarshal(usageEstimateBytes, &usageEstimate))
	assert.Equal(t, float64(2), usageEstimate["clips"])
	assert.Equal(t, "music", usageEstimate["action"])

	usageValue, err := plugin.Engine.Call(t.Context(), "extractUsageOnComplete", map[string]any{}, map[string]any{}, providerTask)
	require.NoError(t, err)
	usageBytes, err := common.Marshal(usageValue)
	require.NoError(t, err)
	var usage map[string]any
	require.NoError(t, common.Unmarshal(usageBytes, &usage))
	assert.Equal(t, float64(3), usage["clips"])
	assert.Equal(t, "music", usage["action"])
}
