package model

import (
	"mime"
	"net/http"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/relaykit/types"
	"github.com/gin-gonic/gin"
)

const callIOLogKey = "call_io_log"
const maxCallIORequestBytes = 256 << 10

type callIOCapture struct {
	userID  int
	enabled bool
	input   map[string]any
	read    bool
}

// BeginCallIOLog freezes the caller's preference before relay processing.
func BeginCallIOLog(c *gin.Context) {
	userID := common.GetContextKeyInt(c, constant.ContextKeyUserId)
	settings, ok := common.GetContextKeyType[dto.UserSetting](c, constant.ContextKeyUserSetting)
	if userID <= 0 || !ok {
		return
	}
	c.Set(callIOLogKey, &callIOCapture{
		userID: userID, enabled: settings.RecordCallIO == nil || *settings.RecordCallIO,
	})
}

func appendCallIOLog(c *gin.Context, userID int, other *LogOther, modelName string, stream bool, consume *RecordConsumeLogParams) *LogOther {
	value, _ := c.Get(callIOLogKey)
	capture, _ := value.(*callIOCapture)
	if capture == nil || capture.userID != userID {
		return other
	}
	if other == nil {
		other = NewLogOther()
	}
	snapshot := map[string]any{"enabled": capture.enabled}
	other.SetPublic(callIOLogKey, snapshot)
	if !capture.enabled {
		return other
	}
	if !capture.read {
		capture.input = readCallIOInput(c)
		capture.read = true
	}
	input := copyLogOtherMap(capture.input)
	if input == nil {
		input = make(map[string]any)
	}
	if len(modelName) <= 160 && callIOModelPattern.MatchString(modelName) {
		input["model"] = modelName
	}
	if _, present := input["stream"]; !present {
		input["stream"] = stream
	}
	snapshot["input"] = input

	output := make(map[string]any)
	if consume == nil {
		if status, ok := other.public["status_code"].(int); ok && status >= 100 && status <= 599 {
			output["status_code"] = status
		}
		if code, ok := other.public["error_code"].(types.ErrorCode); ok && isCallIOPlatformError(code) {
			output["error_code"] = code
		}
	} else {
		if c.Writer.Written() {
			output["status_code"] = c.Writer.Status()
		}
		if consume.PromptTokens > 0 || consume.CompletionTokens > 0 {
			output["usage"] = map[string]int{
				"prompt_tokens": consume.PromptTokens, "completion_tokens": consume.CompletionTokens,
			}
		}
		if consume.OutputImageCount > 0 {
			output["generated_images"] = consume.OutputImageCount
		}
		switch consume.TaskStatus {
		case TaskStatusNotStart, TaskStatusSubmitted, TaskStatusQueued, TaskStatusInProgress, TaskStatusSuccess, TaskStatusFailure:
			output["task_status"] = consume.TaskStatus
		}
		if status, ok := other.public["stream_status"].(map[string]any); ok {
			if reason, ok := status["end_reason"].(string); ok && slices.Contains(callIOStreamEndReasons, reason) {
				output["end_reason"] = reason
			}
			if source, ok := status["usage_source"].(string); ok && slices.Contains(callIOUsageSources, source) {
				output["usage_source"] = source
			}
		}
	}
	if len(output) != 0 {
		snapshot["output"] = output
	}
	return other
}

var callIOModelPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:/@+*-]*$`)
var callIOStreamEndReasons = []string{"done", "eof", "client_gone", "timeout", "scanner_error", "handler_stop", "panic", "ping_fail"}
var callIOUsageSources = []string{"upstream_actual", "local_estimated", "missing"}

func isCallIOPlatformError(code types.ErrorCode) bool {
	switch code {
	case types.ErrorCodeInvalidRequest, types.ErrorCodeReadRequestBodyFailed, types.ErrorCodeConvertRequestFailed,
		types.ErrorCodeAccessDenied, types.ErrorCodeModelPriceError, types.ErrorCodeInsufficientUserQuota,
		types.ErrorCodeModelNotFound, types.ErrorCodeGetChannelFailed, types.ErrorCodeDoRequestFailed,
		types.ErrorCodeBadResponseStatusCode, types.ErrorCodeBadResponse, types.ErrorCodeBadResponseBody,
		types.ErrorCodeReadResponseBodyFailed, types.ErrorCodeEmptyResponse, types.ErrorCodePromptBlocked,
		types.ErrorCodeChannelNoAvailableKey, types.ErrorCodeChannelResponseTimeExceeded:
		return true
	default:
		return false
	}
}

// Use the parsed multipart form or an independent cursor over the existing
// relay body cache. Large JSON keeps the model and stream metadata.
func readCallIOInput(c *gin.Context) map[string]any {
	mediaType, _, _ := mime.ParseMediaType(c.Request.Header.Get("Content-Type"))
	if mediaType == "multipart/form-data" {
		return summarizeCallIOMultipart(c.Request)
	}
	if mediaType != "application/json" && !strings.HasSuffix(mediaType, "+json") {
		return nil
	}
	value, _ := c.Get(common.KeyBodyStorage)
	storage, _ := value.(common.BodyStorage)
	if storage == nil || storage.Size() > maxCallIORequestBytes {
		return nil
	}
	reader, err := storage.NewReader()
	if err != nil {
		return nil
	}
	defer reader.Close()
	var body map[string]common.JSONRawMessage
	if common.DecodeJson(reader, &body) != nil {
		return nil
	}
	return summarizeCallIOInput(body)
}

func summarizeCallIOMultipart(request *http.Request) map[string]any {
	form := request.MultipartForm
	if form == nil {
		return nil
	}
	fields := make(map[string]any)
	for _, key := range callIOIntegerFields {
		if values := form.Value[key]; len(values) == 1 && len(values[0]) <= 20 {
			if value, err := strconv.ParseInt(values[0], 10, 64); err == nil {
				fields[key] = value
			}
		}
	}
	for key, allowed := range callIOEnumFields {
		if values := form.Value[key]; len(values) == 1 && slices.Contains(allowed, values[0]) {
			fields[key] = values[0]
		}
	}
	if values := form.Value["size"]; len(values) == 1 && len(values[0]) <= 16 && callIOSizePattern.MatchString(values[0]) {
		fields["size"] = values[0]
	}
	if count := len(form.File["image"]) + len(form.File["image[]"]); count > 0 {
		fields["input_image_count"] = count
	}
	return fields
}

var callIOIntegerFields = []string{
	"max_tokens", "max_completion_tokens", "max_output_tokens", "maxOutputTokens", "top_k", "topK",
	"n", "candidateCount", "seed", "duration", "seconds", "width", "height",
}
var callIONumberFields = []string{"temperature", "top_p", "topP", "frequency_penalty", "presence_penalty"}
var callIOBooleanFields = []string{"stream", "generate_audio", "watermark", "draft", "return_last_frame", "background"}
var callIOEnumFields = map[string][]string{
	"resolution":               {"360p", "480p", "540p", "720p", "768p", "1080p", "1440p", "2160p", "2k", "4k", "2K", "4K"},
	"ratio":                    {"adaptive", "auto", "1:1", "16:9", "9:16", "4:3", "3:4", "3:2", "2:3", "21:9"},
	"aspect_ratio":             {"adaptive", "auto", "1:1", "16:9", "9:16", "4:3", "3:4", "3:2", "2:3", "21:9"},
	"quality":                  {"auto", "standard", "hd", "low", "medium", "high"},
	"output_format":            {"png", "jpeg", "jpg", "webp", "svg", "gif", "mp4", "mov"},
	"response_format":          {"url", "b64_json"},
	"reasoning_effort":         {"none", "minimal", "low", "medium", "high", "xhigh", "max"},
	"omni_reference_task_type": {"auto", "reference", "edit", "extend"},
}
var callIOSizePattern = regexp.MustCompile(`^(auto|[1-9][0-9]{1,4}x[1-9][0-9]{1,4}|[124][kK])$`)

func summarizeCallIOInput(body map[string]common.JSONRawMessage) map[string]any {
	result := callIOParameters(body)
	for _, key := range []string{"generationConfig", "parameters"} {
		var nested map[string]common.JSONRawMessage
		if raw := body[key]; len(raw) > 0 && common.Unmarshal(raw, &nested) == nil {
			if fields := callIOParameters(nested); len(fields) != 0 {
				result[key] = fields
			}
		}
	}
	// Native video content carries media roles alongside prompts and URLs.
	// Decode just the role and type, and emit counts with fixed keys.
	var content []struct {
		Type string `json:"type"`
		Role string `json:"role"`
	}
	if raw := body["content"]; len(raw) > 0 && common.Unmarshal(raw, &content) == nil {
		counts := map[string]int{}
		for _, item := range content {
			switch item.Type {
			case "image_url":
				counts["images"]++
			case "video_url":
				counts["videos"]++
			case "audio_url":
				counts["audio"]++
			case "draft_task":
				counts["draft_tasks"]++
			default:
				switch item.Role {
				case "first_frame", "last_frame", "reference_image":
					counts["images"]++
				case "reference_video":
					counts["videos"]++
				case "reference_audio":
					counts["audio"]++
				}
			}
		}
		if len(counts) != 0 {
			result["media_counts"] = counts
		}
	}
	if raw := body["image"]; len(raw) > 0 {
		var images []common.JSONRawMessage
		if raw[0] == '[' && common.Unmarshal(raw, &images) == nil {
			result["input_image_count"] = len(images)
		} else if raw[0] == '"' && len(raw) > 2 {
			result["input_image_count"] = 1
		}
	}
	return result
}

func callIOParameter[T bool | int64 | float64 | string](body map[string]common.JSONRawMessage, key string) (T, bool) {
	var value T
	raw := body[key]
	// Scalars here are short configuration values. Never unmarshal a long string.
	ok := len(raw) > 0 && len(raw) <= 64 && string(raw) != "null" && common.Unmarshal(raw, &value) == nil
	return value, ok
}

func callIOParameters(body map[string]common.JSONRawMessage) map[string]any {
	fields := make(map[string]any)
	for _, key := range callIOIntegerFields {
		if value, ok := callIOParameter[int64](body, key); ok {
			fields[key] = value
		}
	}
	for _, key := range callIONumberFields {
		if value, ok := callIOParameter[float64](body, key); ok {
			fields[key] = value
		}
	}
	for _, key := range callIOBooleanFields {
		if value, ok := callIOParameter[bool](body, key); ok {
			fields[key] = value
		}
	}
	for key, allowed := range callIOEnumFields {
		if value, ok := callIOParameter[string](body, key); ok && slices.Contains(allowed, value) {
			fields[key] = value
		}
	}
	if size, ok := callIOParameter[string](body, "size"); ok && callIOSizePattern.MatchString(size) {
		fields["size"] = size
	}
	return fields
}
