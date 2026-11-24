package model

import (
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/relaykit/types"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func callIOTestContext(t *testing.T, body string, enabled *bool) *gin.Context {
	t.Helper()
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json; charset=utf-8")
	c.Request.Header.Set("Authorization", "Bearer sensitive-header")
	common.SetContextKey(c, constant.ContextKeyUserId, 12)
	common.SetContextKey(c, constant.ContextKeyUserSetting, dto.UserSetting{RecordCallIO: enabled})
	BeginCallIOLog(c)
	storage, err := common.GetBodyStorage(c)
	require.NoError(t, err)
	c.Request.Body = io.NopCloser(storage)
	t.Cleanup(func() { common.CleanupBodyStorage(c) })
	return c
}

func TestCallIOLogKeepsOnlySafeFields(t *testing.T) {
	c := callIOTestContext(t, `{
		"prompt":"sensitive-prompt","messages":[{"content":"sensitive-message"}],
		"tools":[{"arguments":"sensitive-tool"}],"api_key":"sensitive-key",
		"content":[{"type":"image_url","image_url":{"url":"https://private.invalid/sensitive-url"}},
		           {"type":"video_url"},{"type":"text","text":"sensitive-video-prompt"}],
		"image":["sensitive-base64","sensitive-image"],"callback_url":"https://private.invalid/sensitive-callback",
		"temperature":0,"max_tokens":2048,"stream":false,"size":"1024x1024",
		"quality":"sensitive-quality","generationConfig":{"topP":0.75,"custom":"sensitive-value"},
		"parameters":{"duration":5,"generate_audio":false,"resolution":"720p","url":"sensitive-nested"}
	}`, nil)
	c.Status(http.StatusOK)
	c.Writer.WriteHeaderNow()
	other := NewLogOther()
	other.SetPublic("stream_status", map[string]any{
		"end_reason": "client_gone", "usage_source": "upstream_actual",
		"end_error": "sensitive-error", "errors": []string{"sensitive-error-list"},
	})
	other.SetAdmin("channel_id", 89)
	params := &RecordConsumeLogParams{PromptTokens: 12, CompletionTokens: 30, OutputImageCount: 2, TaskStatus: TaskStatusSubmitted}
	result := appendCallIOLog(c, 12, other, "public-model", false, params)
	rawSnapshot, err := common.Marshal(result.Snapshot()[callIOLogKey])
	require.NoError(t, err)
	require.JSONEq(t, `{
		"enabled":true,
		"input":{"model":"public-model","stream":false,"temperature":0,"max_tokens":2048,
		         "size":"1024x1024","media_counts":{"images":1,"videos":1},"input_image_count":2,
		         "generationConfig":{"topP":0.75},"parameters":{"duration":5,"generate_audio":false,"resolution":"720p"}},
		"output":{"status_code":200,"end_reason":"client_gone","usage_source":"upstream_actual",
		          "usage":{"prompt_tokens":12,"completion_tokens":30},"generated_images":2,"task_status":"SUBMITTED"}
	}`, string(rawSnapshot))
	var user, admin map[string]common.JSONRawMessage
	require.NoError(t, common.UnmarshalJsonStr(formatLogOtherJSON(result.JSONString(), logOtherVisibilityUser), &user))
	require.NoError(t, common.UnmarshalJsonStr(formatLogOtherJSON(result.JSONString(), logOtherVisibilityAdmin), &admin))
	require.NotContains(t, user, "admin_info")
	require.JSONEq(t, string(user[callIOLogKey]), string(admin[callIOLogKey]))
	bodyAfterLogging, err := io.ReadAll(c.Request.Body)
	require.NoError(t, err)
	require.Contains(t, string(bodyAfterLogging), "sensitive-prompt", "logging must not consume or mutate the request body")
}

func TestCallIOLogCapturesPreferenceAndSafeErrorStatus(t *testing.T) {
	c := callIOTestContext(t, `{"temperature":0.5,"prompt":"private prompt"}`, common.GetPointer(false))
	common.SetContextKey(c, constant.ContextKeyUserSetting, dto.UserSetting{})
	result := appendCallIOLog(c, 12, nil, "public-model", false, nil)
	require.JSONEq(t, `{"call_io_log":{"enabled":false}}`, result.JSONString())
	require.Nil(t, appendCallIOLog(c, 13, nil, "public-model", false, nil))

	c = callIOTestContext(t, `{"temperature":"private prompt"}`, common.GetPointer(true))
	other := NewLogOther()
	other.SetPublic("status_code", http.StatusBadGateway)
	other.SetPublic("error_code", types.ErrorCode("provider-controlled-error"))
	result = appendCallIOLog(c, 12, other, "public-model", true, nil)
	require.Equal(t, map[string]any{"status_code": http.StatusBadGateway}, result.public[callIOLogKey].(map[string]any)["output"])
	other.SetPublic("error_code", types.ErrorCodeBadResponseStatusCode)
	result = appendCallIOLog(c, 12, other, "public-model", true, nil)
	require.Equal(t, types.ErrorCodeBadResponseStatusCode, result.public[callIOLogKey].(map[string]any)["output"].(map[string]any)["error_code"])
}

func TestCallIOLogVideoEnhancementFrameRate(t *testing.T) {
	const modelName = "ai-mediakit-video-enhancement-std"
	for _, tc := range []struct {
		name       string
		fps        string
		parameters string
	}{
		{"integer frame rate", `30`, `{"resolution":"1080p","fps":30}`},
		{"decimal frame rate encoding", `30.0`, `{"resolution":"1080p","fps":30}`},
		{"free text omitted", `"private text"`, `{"resolution":"1080p"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := callIOTestContext(t, `{
				"model":"`+modelName+`",
				"input":{"video_url":"https://private.invalid/source.mp4"},
				"parameters":{"resolution":"1080p","fps":`+tc.fps+`}
			}`, nil)
			result := appendCallIOLog(c, 12, nil, modelName, false, &RecordConsumeLogParams{})
			require.JSONEq(t, `{"call_io_log":{"enabled":true,"input":{"model":"`+modelName+`","stream":false,"parameters":`+tc.parameters+`}}}`, result.JSONString())
		})
	}
}

func TestCallIOLogLargeBodyUsesBasicMetadata(t *testing.T) {
	c := callIOTestContext(t, `{"max_tokens":2048,"prompt":"`+strings.Repeat("x", maxCallIORequestBytes)+`"}`, nil)
	result := appendCallIOLog(c, 12, nil, "public-model", true, &RecordConsumeLogParams{})
	require.JSONEq(t, `{"call_io_log":{"enabled":true,"input":{"model":"public-model","stream":true}}}`, result.JSONString())
}

func TestCallIOLogMultipartOmitsContentAndFileNames(t *testing.T) {
	c := callIOTestContext(t, "", nil)
	c.Request.Header.Set("Content-Type", "multipart/form-data; boundary=testing")
	c.Request.MultipartForm = &multipart.Form{
		Value: map[string][]string{
			"prompt": {"private prompt"}, "n": {"2"}, "size": {"1024x1024"},
			"quality": {"private value"},
		},
		File: map[string][]*multipart.FileHeader{
			"image[]": {{Filename: "private-source.png"}},
			"mask":    {{Filename: "private-mask.png"}},
		},
	}
	result := appendCallIOLog(c, 12, nil, "public-model", false, &RecordConsumeLogParams{})
	require.JSONEq(t, `{"call_io_log":{"enabled":true,"input":{"model":"public-model","stream":false,"n":2,"size":"1024x1024","input_image_count":1}}}`, result.JSONString())
}
