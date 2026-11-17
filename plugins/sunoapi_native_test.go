package plugins_test

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/pkg/jsplugin"
	builtinplugins "github.com/QuantumNous/new-api/plugins"
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
