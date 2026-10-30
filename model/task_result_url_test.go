package model

import "testing"

func TestTaskResultURLRequiresSuccess(t *testing.T) {
	for _, status := range []TaskStatus{TaskStatusNotStart, TaskStatusQueued, TaskStatusInProgress, TaskStatusFailure} {
		task := &Task{Status: status, FailReason: "upstream failed", PrivateData: TaskPrivateData{ResultURL: "https://provider.example/stale.mp4"}}
		if got := task.GetResultURL(); got != "" {
			t.Fatalf("status %s returned a result URL: %q", status, got)
		}
	}
	modern := &Task{Status: TaskStatusSuccess, PrivateData: TaskPrivateData{ResultURL: "https://provider.example/video.mp4"}}
	legacy := &Task{Status: TaskStatusSuccess, FailReason: "https://provider.example/legacy.mp4"}
	if modern.GetResultURL() != modern.PrivateData.ResultURL || legacy.GetResultURL() != legacy.FailReason {
		t.Fatal("successful modern/legacy video URLs must remain available to download handlers")
	}
}
