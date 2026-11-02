package jsplugin

import (
	"testing"

	pluginruntime "github.com/QuantumNous/new-api/pkg/jsplugin"
	"github.com/stretchr/testify/require"
)

func TestResolvedUsageRequestAcceptsNumericEnumSelectors(t *testing.T) {
	adaptor := &TaskAdaptor{plugin: &pluginruntime.LoadedPlugin{Meta: pluginruntime.Meta{
		UsageSchema: map[string]pluginruntime.UsageFieldSchema{
			"fps": {Enum: []string{"30", "60", "120"}},
		},
	}}}

	require.NoError(t, adaptor.validateResolvedUsageRequest(map[string]any{
		"parameters": map[string]any{"fps": float64(30)},
	}, ""))
	require.ErrorContains(t, adaptor.validateResolvedUsageRequest(map[string]any{
		"parameters": map[string]any{"fps": float64(24)},
	}, ""), "plugin usage enum is not an allowed value")
}

func TestCompletionUsageEnumRemainsStringOnly(t *testing.T) {
	schema := pluginruntime.UsageFieldSchema{Enum: []string{"30", "60", "120"}}
	_, err := validateUsageValue(float64(30), schema, false)
	require.ErrorContains(t, err, "plugin usage enum must be a string")
}

func TestDoubaoNativeAutomaticDurationSkipsBillingValidationOnly(t *testing.T) {
	adaptor := &TaskAdaptor{plugin: &pluginruntime.LoadedPlugin{Meta: pluginruntime.Meta{Key: "doubao"}}}
	request := map[string]any{
		"metadata": map[string]any{"duration": float64(-1)},
		"__doubao_native_body": map[string]any{
			"model":    "doubao-seedance-2.0",
			"duration": float64(-1),
		},
	}

	require.NoError(t, adaptor.validateResolvedUsageRequest(request, "doubao-seedance-2.0"))
	require.Equal(t, float64(-1), request["metadata"].(map[string]any)["duration"], "validation must not mutate persisted request data")
	require.Equal(t, float64(-1), request["__doubao_native_body"].(map[string]any)["duration"], "the provider request must retain duration=-1")
}

func TestDoubaoNativeAutomaticDurationDoesNotBypassOtherUsageValidation(t *testing.T) {
	for _, test := range []struct {
		name        string
		plugin      string
		duration    float64
		seconds     float64
		wantInvalid bool
	}{
		{name: "duration below sentinel", plugin: "doubao", duration: -2, wantInvalid: true},
		{name: "zero duration", plugin: "doubao", duration: 0, wantInvalid: true},
		{name: "negative seconds", plugin: "doubao", duration: -1, seconds: -1, wantInvalid: true},
		{name: "other plugin", plugin: "other", duration: -1, wantInvalid: true},
		{name: "positive duration", plugin: "doubao", duration: 5},
	} {
		t.Run(test.name, func(t *testing.T) {
			adaptor := &TaskAdaptor{plugin: &pluginruntime.LoadedPlugin{Meta: pluginruntime.Meta{Key: test.plugin}}}
			nativeBody := map[string]any{"duration": test.duration}
			if test.seconds != 0 {
				nativeBody["seconds"] = test.seconds
			}
			request := map[string]any{
				"metadata":             map[string]any{"duration": test.duration},
				"__doubao_native_body": nativeBody,
			}
			err := adaptor.validateResolvedUsageRequest(request, "doubao-seedance-2.0")
			if test.wantInvalid {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
