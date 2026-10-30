package relay

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/relay/channel/task/taskcommon"
)

func publicVideoTask(task *model.Task) *dto.PublicVideoTask {
	createdAt := task.CreatedAt
	if createdAt == 0 {
		createdAt = task.SubmitTime
	}
	out := &dto.PublicVideoTask{
		TaskID: task.TaskID, Model: task.Properties.OriginModelName,
		Action: constant.NormalizeTaskAction(task.Action), Status: string(task.Status),
		Progress: task.Progress, CreatedAt: createdAt, UpdatedAt: task.UpdatedAt,
		SubmitTime: task.SubmitTime, StartTime: task.StartTime, FinishTime: task.FinishTime,
	}
	if task.Status == model.TaskStatusSuccess {
		// Expose an authenticated gateway resource, never the upstream URL.
		out.ResultURL = taskcommon.BuildProxyURL(task.TaskID)
		if upstreamURL := task.GetResultURL(); upstreamURL != "" {
			out.Metadata = map[string]string{"url": upstreamURL}
		}
	}
	if task.Status == model.TaskStatusFailure {
		out.Error = publicVideoTaskError(task)
		// Preserve the legacy failure field without forwarding raw provider text.
		out.FailReason = out.Error.Message
	}
	return out
}

var unsupportedVideoField = regexp.MustCompile(`^field ([A-Za-z][A-Za-z0-9_]{0,63}) is not supported by model [^\r\n]+$`)

func publicVideoTaskError(task *model.Task) *dto.PublicVideoTaskError {
	var provider struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	// Only use recognized error categories. Messages can include routing names,
	// endpoints, upstream IDs, or credentials and must not be copied verbatim.
	_ = common.Unmarshal(task.Data, &provider)
	for _, message := range []string{task.FailReason, provider.Error.Message} {
		if match := unsupportedVideoField.FindStringSubmatch(strings.TrimSpace(message)); len(match) == 2 {
			return &dto.PublicVideoTaskError{
				Code:    "invalid_request",
				Message: fmt.Sprintf("The %s parameter is not supported by this model.", match[1]),
			}
		}
	}
	switch strings.ToLower(provider.Error.Code) {
	case "invalid_request", "invalid_request_error", "invalidparameter", "param_invalid":
		return &dto.PublicVideoTaskError{Code: "invalid_request", Message: "Invalid video generation parameters. Please check your request."}
	case "content_policy_violation", "content_filter", "content_filtered":
		return &dto.PublicVideoTaskError{Code: "content_policy_violation", Message: "The request could not be processed due to content restrictions."}
	default:
		return &dto.PublicVideoTaskError{Code: "video_generation_failed", Message: "Video generation failed. Please retry or contact support with the task ID."}
	}
}
