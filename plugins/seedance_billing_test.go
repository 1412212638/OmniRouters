package plugins_test

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/pkg/billingexpr"
	"github.com/QuantumNous/new-api/service"
	"github.com/stretchr/testify/require"
)

func seedanceObject(t *testing.T, value any) map[string]any {
	t.Helper()
	data, err := common.Marshal(value)
	require.NoError(t, err)
	var result map[string]any
	require.NoError(t, common.Unmarshal(data, &result))
	return result
}

func TestSeedanceBillingDraftInheritsInputAndSettlesActual4KUsage(t *testing.T) {
	plugin := newDoubaoProfilePlugin(t)
	for _, input := range []string{"none", "video"} {
		t.Run(input, func(t *testing.T) {
			ctx := map[string]any{
				"model": "doubao-seedance-2.5", "upstreamModel": "doubao-seedance-2.5", "baseUrl": "https://ark.example", "apiKey": "secret",
				"requestBody": map[string]any{"metadata": map[string]any{
					"resolution": "4k", "content": []any{map[string]any{"type": "draft_task", "draft_task": map[string]any{"id": "task_sample"}}},
				}},
				"originTasks": []any{map[string]any{"taskId": "task_sample", "upstreamTaskId": "cgt-sample", "billingFacts": map[string]any{"video_input": input}}},
			}
			value, err := plugin.Engine.Call(t.Context(), "extractUsage", ctx)
			require.NoError(t, err)
			facts := seedanceObject(t, value)
			require.Equal(t, input, facts["video_input"])
			require.Equal(t, "4k", facts["resolution"])
			require.Greater(t, facts["tokens"].(float64), float64(0))

			value, err = plugin.Engine.Call(t.Context(), "buildSubmitRequest", ctx)
			require.NoError(t, err)
			body := seedanceObject(t, value)["body"].(map[string]any)
			require.NotContains(t, body, "duration")
			require.Equal(t, "cgt-sample", body["content"].([]any)[0].(map[string]any)["draft_task"].(map[string]any)["id"])

			// Completion overrides only actual tokens/resolution. Its missing
			// reference media must never switch the frozen input price tier.
			value, err = plugin.Engine.Call(t.Context(), "extractUsageOnComplete", ctx, nil, map[string]any{
				"status": "succeeded", "resolution": "4k", "usage": map[string]any{"completion_tokens": 2000, "total_tokens": 2500},
			})
			require.NoError(t, err)
			expr := `u("video_input") == "video" ? tier("video", u("tokens") * 0.000002) : tier("none", u("tokens") * 0.000005)`
			snap := &billingexpr.BillingSnapshot{ExprString: expr, ExprVersion: 1, TaskUsageBilling: true, QuotaPerUnit: 1000000, GroupRatio: 0.6, UsageFacts: facts}
			settled, actual, err := service.EvaluateTaskCompletionUsage(snap, seedanceObject(t, value))
			require.NoError(t, err)
			require.Equal(t, input, actual["video_input"])
			require.Equal(t, float64(2000), actual["tokens"])
			want := 6000
			if input == "video" {
				want = 2400
			}
			require.Equal(t, want, settled.ActualQuotaAfterGroup)
		})
	}
}

func TestSeedanceBillingNativeFieldsAndDraftState(t *testing.T) {
	plugin := newDoubaoProfilePlugin(t)
	original := map[string]any{"model": "doubao-seedance-2.5", "resolution": "480p", "draft": true, "generate_audio": false,
		"duration": -1, "service_tier": "default", "priority": 0, "output_format": "mov", "callback_url": "https://customer.example/hook",
		"content": []any{map[string]any{"type": "video_url", "video_url": map[string]any{"url": "https://example.com/ref.mov"}, "role": "reference_video"}},
	}
	intent, err := plugin.Engine.CallPath(t.Context(), "native", []string{"createTask"}, map[string]any{"body": map[string]any{"kind": "json", "value": original}})
	require.NoError(t, err)
	ctx := map[string]any{"model": "doubao-seedance-2.5", "baseUrl": "https://ark.example", "requestBody": seedanceObject(t, intent)["requestBody"]}
	value, err := plugin.Engine.Call(t.Context(), "buildSubmitRequest", ctx)
	require.NoError(t, err)
	body := seedanceObject(t, value)["body"].(map[string]any)
	for _, key := range []string{"draft", "generate_audio", "duration", "service_tier", "priority", "output_format", "callback_url", "content"} {
		require.Equal(t, seedanceObject(t, original)[key], body[key], key)
	}
	value, err = plugin.Engine.Call(t.Context(), "parseSubmitResponse", ctx, map[string]any{"body": map[string]any{"id": "cgt-sample"}})
	require.NoError(t, err)
	require.Equal(t, "video", seedanceObject(t, value)["state"].(map[string]any)["seedance_video_input"])
	value, err = plugin.Engine.Call(t.Context(), "extractUsage", ctx)
	require.NoError(t, err)
	require.Equal(t, float64(30*854*480*24/1024), seedanceObject(t, value)["tokens"])
	for _, alias := range []string{"doubao-seedance-2.5", "doubao-seedance-2-5-260628"} {
		schema, _ := plugin.Meta.UsageForModel(alias)
		require.Contains(t, schema["resolution"].Enum, "4k")
	}
	for _, alias := range []string{"doubao-seedance-2.5-flash", "doubao-seedance-2-5-flash-261115"} {
		schema, _ := plugin.Meta.UsageForModel(alias)
		require.NotContains(t, schema["resolution"].Enum, "4k")
	}
}

func TestSeedanceBillingMissingDraftFactsFailsBeforeReservation(t *testing.T) {
	plugin := newDoubaoProfilePlugin(t)
	_, err := plugin.Engine.Call(t.Context(), "extractUsage", map[string]any{
		"model": "doubao-seedance-2.5", "requestBody": map[string]any{"metadata": map[string]any{
			"content": []any{map[string]any{"type": "draft_task", "draft_task": map[string]any{"id": "task_without_facts"}}},
		}},
	})
	require.ErrorContains(t, err, "billing information is unavailable")
}
