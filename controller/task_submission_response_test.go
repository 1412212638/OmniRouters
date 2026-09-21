package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	pluginruntime "github.com/QuantumNous/new-api/pkg/jsplugin"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/gin-gonic/gin"
)

func TestPresentNativeSubmissionReturnsSynchronousAnswers(t *testing.T) {
	plugin, err := pluginruntime.NewRegistry().Register(`
export const meta = {apiVersion: 1, key: "sync-test", name: "Sync test", version: "1.0.0", author: {name: "Test"}, models: ["jev-1.13.0"], fetchMode: "per_task",
routes: [{method: "POST", path: "/sync-test/create", type: "submit", decode: "decode", render: "render"}]};
export function buildSubmitRequest(ctx) { return {url: ctx.baseUrl}; }
export function parseSubmitResponse() { return {taskId: "task"}; }
export function buildQueryRequest(ctx) { return {url: ctx.baseUrl}; }
export function parseTaskResult() { return {status: "SUCCESS"}; }
export const native = {decode: function(ctx) { return {}; }, render: function(ctx, task) { return task.data; }};
`, pluginruntime.Options{Key: "sync-test"})
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/systemone", nil)
	c.Set(pluginruntime.ContextKeyPinnedRoute, pluginruntime.PinnedRoute{Plugin: plugin, Route: plugin.Meta.Routes[0]})
	c.Set(pluginruntime.ContextKeyRouteRequest, pluginruntime.RouteRequestContext{Path: "/v1/systemone", Method: http.MethodPost})
	payload := []byte(`{"model":"jev-1.13.0","answers":{"refund":{"type":"noul","value":true}},"usage":{"input_tokens":24}}`)
	presentTaskSubmission(c, &taskSubmissionOutcome{
		Task: &model.Task{TaskID: "task_public", Status: model.TaskStatusSuccess, Data: payload}, RelayInfo: &relaycommon.RelayInfo{},
	})
	var got, want map[string]any
	if err := common.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if err := common.Unmarshal(payload, &want); err != nil {
		t.Fatal(err)
	}
	gotJSON, _ := common.Marshal(got)
	wantJSON, _ := common.Marshal(want)
	if w.Code != http.StatusOK || string(gotJSON) != string(wantJSON) {
		t.Fatalf("expected synchronous answers and usage, got %d: %s", w.Code, w.Body.String())
	}
}

func TestPresentTaskSubmissionLegacyEndpointReturnsTaskID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/video/generations", nil)
	info := &relaycommon.RelayInfo{}
	info.OriginModelName = "doubao-seedance-test"
	presentTaskSubmission(c, &taskSubmissionOutcome{
		Task: &model.Task{TaskID: "task_public", SubmitTime: 123}, RelayInfo: info,
	})
	var body map[string]any
	if err := common.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected JSON response: %v; body=%q", err, w.Body.String())
	}
	if w.Code != http.StatusOK || body["task_id"] != "task_public" || body["model"] != info.OriginModelName {
		t.Fatalf("unexpected task response: status=%d body=%v", w.Code, body)
	}
}

func TestPresentTaskSubmissionPreservesWrittenResponse(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Data(http.StatusOK, "application/json", []byte(`{"id":"provider-task"}`))
	presentTaskSubmission(c, nil)
	if w.Body.String() != `{"id":"provider-task"}` {
		t.Fatalf("provider response was changed: %q", w.Body.String())
	}
}
