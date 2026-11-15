package plugins_test

import (
	"net/http"
	"testing"

	"github.com/QuantumNous/new-api/pkg/jsplugin"
	builtinplugins "github.com/QuantumNous/new-api/plugins"
	"github.com/stretchr/testify/require"
)

func TestHailuoPollingStatusBackport(t *testing.T) {
	plugin := h3NativePlugin(t)
	for _, test := range []struct {
		name, expected string
		body           map[string]any
	}{
		{"H3 queued", "QUEUED", map[string]any{"task": map[string]any{"status": "queued"}}},
		{"H3 running", "IN_PROGRESS", map[string]any{"task": map[string]any{"status": "running"}}},
		{"H3 succeeded", "SUCCESS", map[string]any{"task": map[string]any{"status": "succeeded", "content": map[string]any{"url": "https://cdn.example/video.mp4"}}}},
		{"H3 failed", "FAILURE", map[string]any{"task": map[string]any{"status": "failed"}}},
		{"H3 cancelled", "FAILURE", map[string]any{"task": map[string]any{"status": "cancelled"}}},
		{"H3 unknown", "UNKNOWN", map[string]any{"task": map[string]any{"status": "new-provider-state"}}},
		{"H3 prototype key", "UNKNOWN", map[string]any{"task": map[string]any{"status": "toString"}}},
		{"v1 preparing", "IN_PROGRESS", map[string]any{"status": "Preparing"}},
		{"v1 queueing", "IN_PROGRESS", map[string]any{"status": "Queueing"}},
		{"v1 processing", "IN_PROGRESS", map[string]any{"status": "Processing"}},
		{"v1 success", "SUCCESS", map[string]any{"status": "Success", "base_resp": map[string]any{"status_code": 0}}},
		{"v1 success without error code", "SUCCESS", map[string]any{"status": "Success", "base_resp": map[string]any{}}},
		{"v1 processing without error code", "IN_PROGRESS", map[string]any{"status": "Processing", "base_resp": map[string]any{"status_msg": ""}}},
		{"v1 failure", "FAILURE", map[string]any{"status": "Fail", "base_resp": map[string]any{"status_code": 0}}},
		{"v1 error overrides stale running", "FAILURE", map[string]any{"status": "Processing", "base_resp": map[string]any{"status_code": 2013, "status_msg": "invalid task"}}},
		{"v1 error overrides success", "FAILURE", map[string]any{"status": "Success", "base_resp": map[string]any{"status_code": 2013, "status_msg": "invalid task"}}},
		{"v1 unknown", "UNKNOWN", map[string]any{"status": "new-provider-state"}},
		{"v1 missing", "UNKNOWN", map[string]any{}},
		{"v1 prototype key", "UNKNOWN", map[string]any{"status": "constructor"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			result := h3Call(t, plugin, "parseTaskResult", nil, map[string]any{}, test.body)
			require.Equal(t, test.expected, result["status"])
			if test.expected == "UNKNOWN" || test.expected == "FAILURE" {
				require.NotEmpty(t, result["reason"])
			}
			if test.expected == "IN_PROGRESS" {
				require.Empty(t, result["reason"], "pending polls must not be converted into failures by the poller")
			}
		})
	}
}

func TestAlibabaNativeQueryReflectsHostFailure(t *testing.T) {
	source, err := builtinplugins.Source("alibaba")
	require.NoError(t, err)
	plugin, err := jsplugin.NewRegistry().RegisterFactory(source, jsplugin.Options{Key: "alibaba"})
	require.NoError(t, err)
	for _, status := range []string{"PENDING", "RUNNING", ""} {
		t.Run(status, func(t *testing.T) {
			output := map[string]any{"task_id": "private-id"}
			if status != "" {
				output["task_status"] = status
			}
			usage := map[string]any{"duration": 5}
			data := map[string]any{"request_id": "request-id", "output": output, "usage": usage}
			view := map[string]any{"task_id": "public-id", "status": "FAILURE", "fail_reason": "task timed out", "data": data}
			result := h3Call(t, plugin, "native", []string{"taskStatus"}, map[string]any{}, view)
			resultOutput := result["output"].(map[string]any)
			require.Equal(t, "public-id", resultOutput["task_id"])
			require.Equal(t, "FAILED", resultOutput["task_status"])
			require.Equal(t, "task timed out", resultOutput["message"])
			require.Equal(t, "InternalError", resultOutput["code"])
			require.EqualValues(t, 5, result["usage"].(map[string]any)["duration"])
			require.Equal(t, "private-id", output["task_id"])
			require.NotContains(t, output, "message", "rendering must not mutate the stored upstream snapshot")
		})
	}
	for _, status := range []string{"PENDING", "RUNNING", "SUCCEEDED", "FAILED", "CANCELED"} {
		result := h3Call(t, plugin, "native", []string{"taskStatus"}, map[string]any{}, map[string]any{
			"task_id": "public-id", "status": "IN_PROGRESS", "data": map[string]any{"output": map[string]any{"task_status": status}},
		})
		require.Equal(t, status, result["output"].(map[string]any)["task_status"])
	}
}

func TestHailuoFileContentDescriptorUsesRetrieve(t *testing.T) {
	plugin := h3NativePlugin(t)
	for _, method := range []string{http.MethodGet, http.MethodHead} {
		result := h3Call(t, plugin, "buildContentRequest", nil, map[string]any{
			"artifactKey": "video", "baseUrl": "https://vendor.example/", "apiKey": "private-key",
			"data": map[string]any{"file_id": "file/with?reserved"}, "clientRequest": map[string]any{"method": method},
		})
		require.Equal(t, "https://vendor.example/v1/files/retrieve?file_id=file%2Fwith%3Freserved", result["url"])
		require.Equal(t, http.MethodGet, result["method"])
		require.Equal(t, "json", result["responseType"])
	}
	result := h3Call(t, plugin, "buildContentRequest", nil, map[string]any{
		"artifactKey": "video", "data": map[string]any{"task": map[string]any{"content": map[string]any{"url": "https://cdn.example/h3.mp4"}}},
		"clientRequest": map[string]any{"method": http.MethodHead},
	})
	require.Equal(t, map[string]any{"url": "https://cdn.example/h3.mp4", "method": http.MethodHead, "credentialless": true}, result)
}
