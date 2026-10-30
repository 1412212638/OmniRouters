package relay

import (
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/system_setting"
)

func TestPublicVideoTaskDoesNotExposeInternalData(t *testing.T) {
	originalAddress := system_setting.ServerAddress
	system_setting.ServerAddress = "https://gateway.example"
	t.Cleanup(func() { system_setting.ServerAddress = originalAddress })
	task := &model.Task{
		ID: 13, UserId: 1, ChannelId: 36, Quota: 12345, Group: "internal-group",
		Platform: "private-plugin", TaskID: "task_public", Status: model.TaskStatusSuccess,
		Action: "text_to_video", Progress: "100%", SubmitTime: 100, FinishTime: 150,
		Properties:  model.Properties{OriginModelName: "minimax-h3", UpstreamModelName: "Bloom-minimax-h3-pro"},
		PrivateData: model.TaskPrivateData{Key: "secret-key", UpstreamTaskID: "task_private", ResultURL: "https://private-provider.example/video.mp4"},
	}
	task.SetData(map[string]any{"id": "task_private", "model": "Bloom-minimax-h3-pro", "gateway_request_id": "private-request-id"})
	encoded, err := common.Marshal(dto.TaskResponse[any]{Code: "success", Data: publicVideoTask(task)})
	if err != nil {
		t.Fatal(err)
	}
	var response struct {
		Code string         `json:"code"`
		Data map[string]any `json:"data"`
	}
	if err := common.Unmarshal(encoded, &response); err != nil {
		t.Fatal(err)
	}
	wantFields := map[string]bool{
		"task_id": true, "model": true, "action": true, "status": true, "progress": true,
		"created_at": true, "updated_at": true, "submit_time": true, "start_time": true,
		"finish_time": true, "result_url": true, "metadata": true,
	}
	if response.Code != "success" || len(response.Data) != len(wantFields) {
		t.Fatalf("unexpected public response: %s", encoded)
	}
	for key := range response.Data {
		if !wantFields[key] {
			t.Fatalf("internal field %q exposed", key)
		}
	}
	if response.Data["result_url"] != "https://gateway.example/v1/videos/task_public/content" || response.Data["model"] != "minimax-h3" || response.Data["created_at"] != float64(100) {
		t.Fatalf("incorrect public result/identity/timestamp: %s", encoded)
	}
	metadata, ok := response.Data["metadata"].(map[string]any)
	if !ok || metadata["url"] != "https://private-provider.example/video.mp4" {
		t.Fatalf("upstream metadata URL missing: %s", encoded)
	}
	for _, secret := range []string{"Bloom-", "secret-key", "task_private", "private-provider", "private-request-id", "internal-group", "private-plugin"} {
		if strings.Contains(string(encoded), secret) {
			t.Fatalf("public response leaked %q", secret)
		}
	}
	// Public presentation must not mutate the stored task or admin projection.
	admin := TaskModel2Dto(task)
	if admin.ChannelId != 36 || admin.Quota != 12345 || admin.ResultURL != task.PrivateData.ResultURL || string(admin.Data) != string(task.Data) || admin.Properties.(model.Properties).UpstreamModelName != "Bloom-minimax-h3-pro" {
		t.Fatalf("admin task details were altered: %+v", admin)
	}
}

func TestPublicVideoTaskFailureMessagesAreSafe(t *testing.T) {
	for _, tc := range []struct {
		name, reason, raw, code, message string
	}{
		{
			"unsupported field", "field watermark is not supported by model minimax-h3",
			`{"error":{"code":"invalid_request","message":"field watermark is not supported by model Bloom-minimax-h3-pro","gateway_code":"A40001","gateway_request_id":"private-id"}}`,
			"invalid_request", "The watermark parameter is not supported by this model.",
		},
		{
			"wrapped upstream details", "HTTP 400 https://private-provider.example?key=secret-key",
			`{"error":{"code":"invalid_request","message":"invalid parameter for Bloom-minimax-h3-pro"}}`,
			"invalid_request", "Invalid video generation parameters. Please check your request.",
		},
		{
			"unknown provider code", "upstream private-provider failed with secret-key",
			`{"error":{"code":"A40001","message":"provider-specific details"}}`,
			"video_generation_failed", "Video generation failed. Please retry or contact support with the task ID.",
		},
		{
			"malformed raw data", "private diagnostic", `not json`,
			"video_generation_failed", "Video generation failed. Please retry or contact support with the task ID.",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			task := &model.Task{TaskID: "task_public", Status: model.TaskStatusFailure, FailReason: tc.reason}
			task.Data = []byte(tc.raw)
			public := publicVideoTask(task)
			if public.ResultURL != "" || public.Metadata != nil || public.Error == nil || public.Error.Code != tc.code || public.Error.Message != tc.message || public.FailReason != tc.message {
				t.Fatalf("unexpected public failure: %+v", public)
			}
			encoded, err := common.Marshal(public)
			if err != nil {
				t.Fatal(err)
			}
			for _, secret := range []string{"result_url", "Bloom-", "private-provider", "secret-key", "gateway_", "private-id"} {
				if strings.Contains(string(encoded), secret) {
					t.Fatalf("public error leaked %q: %s", secret, encoded)
				}
			}
			if task.FailReason != tc.reason || string(task.Data) != tc.raw {
				t.Fatal("public failure presentation mutated diagnostic data")
			}
		})
	}
}

func TestPublicVideoTaskPendingHasNoResultOrError(t *testing.T) {
	task := &model.Task{
		TaskID: "task_public", Status: model.TaskStatusInProgress,
		PrivateData: model.TaskPrivateData{ResultURL: "https://private-provider.example/stale.mp4"},
	}
	got := publicVideoTask(task)
	if got.ResultURL != "" || got.Metadata != nil || got.Error != nil || got.FailReason != "" {
		t.Fatalf("pending task exposed a result/error: %+v", got)
	}
}
