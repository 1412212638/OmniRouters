package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/gin-gonic/gin"
)

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
