package dto

import (
	"encoding/json"
)

type TaskError struct {
	// NoRetry prevents duplicate upstream work after a response has been accepted.
	NoRetry    bool   `json:"-"`
	Code       string `json:"code"`
	Message    string `json:"message"`
	Data       any    `json:"data"`
	StatusCode int    `json:"-"`
	LocalError bool   `json:"-"`
	Error      error  `json:"-"`
}

type TaskData interface {
	SunoDataResponse | []SunoDataResponse | string | any
}

const TaskSuccessCode = "success"

type TaskResponse[T TaskData] struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

func (t *TaskResponse[T]) IsSuccess() bool {
	return t.Code == TaskSuccessCode
}

type TaskDto struct {
	ID         int64           `json:"id"`
	CreatedAt  int64           `json:"created_at"`
	UpdatedAt  int64           `json:"updated_at"`
	TaskID     string          `json:"task_id"`
	Platform   string          `json:"platform"`
	UserId     int             `json:"user_id"`
	Group      string          `json:"group"`
	ChannelId  int             `json:"channel_id"`
	Quota      int             `json:"quota"`
	Action     string          `json:"action"`
	Status     string          `json:"status"`
	FailReason string          `json:"fail_reason"`
	ResultURL  string          `json:"result_url,omitempty"` // 任务结果 URL（视频地址等）
	SubmitTime int64           `json:"submit_time"`
	StartTime  int64           `json:"start_time"`
	FinishTime int64           `json:"finish_time"`
	Progress   string          `json:"progress"`
	Properties any             `json:"properties"`
	Username   string          `json:"username,omitempty"`
	Data       json.RawMessage `json:"data"`
}

// PublicVideoTask is the allowlist for client-facing video task queries.
// TaskDto remains the internal/admin projection and must not be used here.
type PublicVideoTask struct {
	TaskID     string                `json:"task_id"`
	Model      string                `json:"model"`
	Action     string                `json:"action"`
	Status     string                `json:"status"`
	Progress   string                `json:"progress"`
	CreatedAt  int64                 `json:"created_at"`
	UpdatedAt  int64                 `json:"updated_at"`
	SubmitTime int64                 `json:"submit_time"`
	StartTime  int64                 `json:"start_time"`
	FinishTime int64                 `json:"finish_time"`
	ResultURL  string                `json:"result_url,omitempty"`
	Metadata   map[string]string     `json:"metadata,omitempty"`
	Usage      *PublicVideoUsage     `json:"usage,omitempty"`
	FailReason string                `json:"fail_reason,omitempty"`
	Error      *PublicVideoTaskError `json:"error,omitempty"`
}

type PublicVideoUsage struct {
	ToolUsage        *int64 `json:"tool_usage,omitempty"`
	TotalToken       *int64 `json:"total_token,omitempty"`
	TotalTokens      *int64 `json:"total_tokens,omitempty"`
	CompletionTokens *int64 `json:"completion_tokens,omitempty"`
}

type PublicVideoTaskError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type FetchReq struct {
	IDs []string `json:"ids"`
}
