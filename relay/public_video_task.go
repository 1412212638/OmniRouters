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
	out.Usage = publicVideoUsage(task)
	if task.Status == model.TaskStatusSuccess {
		// Expose both the gateway download URL and the provider result metadata.
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

func publicVideoUsage(task *model.Task) *dto.PublicVideoUsage {
	var provider struct {
		Usage *struct {
			ToolUsage        *int64 `json:"tool_usage"`
			TotalToken       *int64 `json:"total_token"`
			TotalTokens      *int64 `json:"total_tokens"`
			CompletionTokens *int64 `json:"completion_tokens"`
		} `json:"usage"`
	}
	if task == nil || len(task.Data) == 0 || common.Unmarshal(task.Data, &provider) != nil || provider.Usage == nil {
		return nil
	}
	usage := &dto.PublicVideoUsage{
		ToolUsage: provider.Usage.ToolUsage, TotalToken: provider.Usage.TotalToken,
		TotalTokens: provider.Usage.TotalTokens, CompletionTokens: provider.Usage.CompletionTokens,
	}
	if usage.ToolUsage == nil && usage.TotalToken == nil && usage.TotalTokens == nil && usage.CompletionTokens == nil {
		return nil
	}
	return usage
}

func addOpenAIVideoResultURLs(task *model.Task, response []byte) ([]byte, error) {
	if task == nil || task.Status != model.TaskStatusSuccess {
		return response, nil
	}
	var rendered map[string]any
	if err := common.Unmarshal(response, &rendered); err != nil || rendered == nil {
		return nil, fmt.Errorf("invalid OpenAI video response")
	}
	metadata, _ := rendered["metadata"].(map[string]any)
	if metadata == nil {
		metadata = make(map[string]any)
		rendered["metadata"] = metadata
	}

	proxyURL := taskcommon.BuildProxyURL(task.TaskID)
	upstreamURL := task.GetResultURL()
	if isVideoProxyURL(upstreamURL, task.TaskID) {
		upstreamURL = ""
	}
	if upstreamURL == "" {
		if responseURL, ok := metadata["url"].(string); ok && !isVideoProxyURL(responseURL, task.TaskID) {
			upstreamURL = responseURL
		}
	}
	if upstreamURL == "" {
		upstreamURL = upstreamVideoURLFromTaskData(task.Data, task.TaskID)
	}
	if upstreamURL != "" {
		metadata["url"] = upstreamURL
	}
	metadata["proxy_url"] = proxyURL
	return common.Marshal(rendered)
}

func isVideoProxyURL(value, taskID string) bool {
	if value == "" || taskID == "" {
		return false
	}
	path := strings.SplitN(value, "?", 2)[0]
	return strings.HasSuffix(strings.TrimRight(path, "/"), "/v1/videos/"+taskID+"/content")
}

func upstreamVideoURLFromTaskData(data []byte, taskID string) string {
	var root map[string]any
	if err := common.Unmarshal(data, &root); err != nil {
		return ""
	}
	for _, path := range [][]string{
		{"metadata", "url"},
		{"content", "video_url"},
		{"output", "video_url"},
		{"video_url"},
		{"url"},
	} {
		var current any = root
		for _, key := range path {
			object, ok := current.(map[string]any)
			if !ok {
				current = nil
				break
			}
			current = object[key]
		}
		if value, ok := current.(string); ok && value != "" && !isVideoProxyURL(value, taskID) {
			return value
		}
	}
	return ""
}

var unsupportedVideoField = regexp.MustCompile(`^field ([A-Za-z][A-Za-z0-9_]{0,63}) is not supported by model [^\r\n]+$`)
var missingVideoUnitPrice = regexp.MustCompile(`(?i)^output video unit price is required for mode ([A-Za-z0-9_-]{1,32}) and resolution ([A-Za-z0-9_-]{1,32})$`)

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
		if match := missingVideoUnitPrice.FindStringSubmatch(strings.TrimSpace(message)); len(match) == 3 {
			return &dto.PublicVideoTaskError{
				Code:    "invalid_media_billing",
				Message: fmt.Sprintf("Video pricing is not configured for mode %s at %s resolution. Please contact the service provider.", match[1], match[2]),
			}
		}
		if match := unsupportedVideoField.FindStringSubmatch(strings.TrimSpace(message)); len(match) == 2 {
			return &dto.PublicVideoTaskError{
				Code:    "invalid_request",
				Message: fmt.Sprintf("The %s parameter is not supported by this model.", match[1]),
			}
		}
	}
	switch strings.ToLower(provider.Error.Code) {
	case "invalid_media_billing":
		return &dto.PublicVideoTaskError{Code: "invalid_media_billing", Message: "Video pricing is not configured for this generation mode and resolution. Please contact the service provider."}
	case "invalid_request", "invalid_request_error", "invalidparameter", "param_invalid":
		return &dto.PublicVideoTaskError{Code: "invalid_request", Message: "Invalid video generation parameters. Please check your request."}
	case "content_policy_violation", "content_filter", "content_filtered":
		return &dto.PublicVideoTaskError{Code: "content_policy_violation", Message: "The request could not be processed due to content restrictions."}
	default:
		return &dto.PublicVideoTaskError{Code: "video_generation_failed", Message: "Video generation failed. Please retry or contact support with the task ID."}
	}
}

// sanitizeOpenAIVideoTaskError replaces plugin/provider error objects with the
// same public-safe projection used by /v1/video/generations/{id}.
func sanitizeOpenAIVideoTaskError(task *model.Task, response []byte) ([]byte, error) {
	if task.Status != model.TaskStatusFailure {
		return response, nil
	}
	var rendered map[string]any
	if err := common.Unmarshal(response, &rendered); err != nil || rendered == nil {
		return nil, fmt.Errorf("invalid OpenAI video response")
	}
	publicError := publicVideoTaskError(task)
	errorJSON, err := common.Marshal(publicError)
	if err != nil {
		return nil, err
	}
	var safeError map[string]any
	if err := common.Unmarshal(errorJSON, &safeError); err != nil {
		return nil, err
	}
	rendered["error"] = safeError
	return common.Marshal(rendered)
}
