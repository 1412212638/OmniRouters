package controller

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/relay"
)

func TestTasksToDtoHidesMappedUpstreamModelFromUser(t *testing.T) {
	const raw = `{"id":"task_upstream","model":"Bloom-minimax-h3-pro","usage":{"completion_tokens":42624},"error":{"message":"Bloom-minimax-h3-pro request failed"},"result":{"model_name":"Bloom-minimax-h3-pro","upstream_model_name":"Bloom-minimax-h3-pro","label":"Bloom-minimax-h3-pro"}}`
	task := &model.Task{
		TaskID: "task_public",
		Properties: model.Properties{
			OriginModelName:   "minimax-h3",
			UpstreamModelName: "Bloom-minimax-h3-pro",
		},
		FailReason: "Bloom-minimax-h3-pro generation failed",
		Data:       json.RawMessage(raw),
	}

	userDTO := tasksToDto([]*model.Task{task}, false)[0]
	if userDTO.Properties.(model.Properties).UpstreamModelName != "" {
		t.Fatalf("user properties exposed upstream model: %+v", userDTO.Properties)
	}
	if strings.Contains(string(userDTO.Data), "Bloom-minimax-h3-pro") || strings.Contains(string(userDTO.Data), "upstream_model_name") {
		t.Fatalf("user task data exposed upstream model: %s", userDTO.Data)
	}
	if strings.Contains(userDTO.FailReason, "Bloom-minimax-h3-pro") || !strings.Contains(userDTO.FailReason, "minimax-h3") {
		t.Fatalf("user task failure reason exposed upstream model: %s", userDTO.FailReason)
	}
	var data map[string]json.RawMessage
	if err := common.Unmarshal(userDTO.Data, &data); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"model", "error", "result"} {
		if _, ok := data[field]; !ok {
			t.Fatalf("sanitization dropped task data field %q: %s", field, userDTO.Data)
		}
	}
	var modelName string
	if err := common.Unmarshal(data["model"], &modelName); err != nil || modelName != "minimax-h3" {
		t.Fatalf("expected public model name, got %q (err=%v)", modelName, err)
	}
	var usage struct {
		CompletionTokens int `json:"completion_tokens"`
	}
	if err := common.Unmarshal(data["usage"], &usage); err != nil || usage.CompletionTokens != 42624 {
		t.Fatalf("usage data was not preserved: %+v (err=%v)", usage, err)
	}

	adminDTO := relay.TaskModel2Dto(task)
	if string(adminDTO.Data) != raw || adminDTO.Properties.(model.Properties).UpstreamModelName != "Bloom-minimax-h3-pro" || adminDTO.FailReason != task.FailReason {
		t.Fatalf("admin task data was altered: %+v", adminDTO)
	}
	if string(task.Data) != raw {
		t.Fatal("user projection mutated the stored task data")
	}
}

func TestSanitizeUserTaskDataPreservesLargeNumbers(t *testing.T) {
	const raw = `{"model":"Bloom-minimax-h3-pro","usage_id":18446744073709551615}`
	got := sanitizeUserTaskData(json.RawMessage(raw), "Bloom-minimax-h3-pro", "minimax-h3")
	if !strings.Contains(string(got), `"usage_id":18446744073709551615`) {
		t.Fatalf("large JSON integer changed during projection: %s", got)
	}
}
