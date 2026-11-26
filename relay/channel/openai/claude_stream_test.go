package openai

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/relaykit/types"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestHandleFinalClaudeResponseFinalizesOnlySuccessfulStreams(t *testing.T) {
	oldMode := gin.Mode()
	gin.SetMode(gin.TestMode)
	t.Cleanup(func() { gin.SetMode(oldMode) })

	for _, tc := range []struct {
		name         string
		data         string
		failed       bool
		wantTerminal bool
	}{
		{
			name:         "finish reason already emitted",
			data:         `{"id":"chatcmpl_1","model":"glm-5.3","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}`,
			wantTerminal: true,
		},
		{
			name:   "failed stream",
			data:   `{"id":"chatcmpl_1","model":"glm-5.3","choices":[{"index":0,"delta":{"content":"partial"},"finish_reason":null}]}`,
			failed: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)

			info := &relaycommon.RelayInfo{
				RelayFormat:  types.RelayFormatClaude,
				ChannelMeta:  &relaycommon.ChannelMeta{UpstreamModelName: "glm-5.3"},
				StreamStatus: relaycommon.NewStreamStatus(),
				ClaudeConvertInfo: &relaycommon.ClaudeConvertInfo{
					LastMessagesType: relaycommon.LastMessageTypeNone,
				},
			}
			info.StreamStatus.RequireTerminal()
			info.StreamStatus.SetEndReason(relaycommon.StreamEndReasonDone, nil)
			if tc.failed {
				info.StreamStatus.MarkFailed("upstream_error", "", 0)
			}
			usage := &dto.Usage{PromptTokens: 8, CompletionTokens: 2, TotalTokens: 10}

			HandleFinalResponse(c, info, tc.data, "chatcmpl_1", 1710000000, "glm-5.3", "", usage, true)

			got := recorder.Body.String()
			require.Equal(t, 1, strings.Count(got, "event: message_start\n"))
			if tc.wantTerminal {
				require.Equal(t, 1, strings.Count(got, "event: message_delta\n"))
				require.Equal(t, 1, strings.Count(got, "event: message_stop\n"))
				require.Contains(t, got, `"input_tokens":8`)
				require.Contains(t, got, `"output_tokens":2`)
				require.True(t, strings.Index(got, "event: message_delta\n") < strings.Index(got, "event: message_stop\n"))
			} else {
				require.NotContains(t, got, "event: message_delta\n")
				require.NotContains(t, got, "event: message_stop\n")
			}
		})
	}
}

func TestOaiStreamHandlerReportsFinalClaudeUsage(t *testing.T) {
	oldMode := gin.Mode()
	gin.SetMode(gin.TestMode)
	t.Cleanup(func() { gin.SetMode(oldMode) })
	oldTimeout := constant.StreamingTimeout
	constant.StreamingTimeout = 30
	t.Cleanup(func() { constant.StreamingTimeout = oldTimeout })

	const roleChunk = `{"id":"chatcmpl_1","model":"glm-5.3","choices":[{"index":0,"delta":{"role":"assistant"}}]}`
	const textChunk = `{"id":"chatcmpl_1","model":"glm-5.3","choices":[{"index":0,"delta":{"content":"hello"}}]}`
	const finishChunk = `{"id":"chatcmpl_1","model":"glm-5.3","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}`
	const toolChunk = `{"id":"chatcmpl_1","model":"glm-5.3","choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"id":"call_1","type":"function","function":{"name":"lookup","arguments":"{\"query\":\"hello\"}"}}]}}]}`
	const toolFinishChunk = `{"id":"chatcmpl_1","model":"glm-5.3","choices":[{"index":0,"delta":{},"finish_reason":"tool_calls"}]}`
	const usageChunk = `{"id":"chatcmpl_1","model":"glm-5.3","choices":[],"usage":{"prompt_tokens":10835,"completion_tokens":2,"total_tokens":10837,"prompt_tokens_details":{"cached_tokens":10816}}}`
	for _, tc := range []struct {
		name           string
		chunks         []string
		estimated      bool
		wantContent    string
		wantStopReason string
	}{
		{
			name:           "single content chunk without usage",
			chunks:         []string{textChunk},
			estimated:      true,
			wantContent:    `"text":"hello"`,
			wantStopReason: "end_turn",
		},
		{
			name:           "multiple chunks without usage",
			chunks:         []string{roleChunk, textChunk, finishChunk},
			estimated:      true,
			wantContent:    `"text":"hello"`,
			wantStopReason: "end_turn",
		},
		{
			name:           "tool call without usage",
			chunks:         []string{roleChunk, toolChunk, toolFinishChunk},
			estimated:      true,
			wantContent:    `"type":"input_json_delta"`,
			wantStopReason: "tool_use",
		},
		{
			name:           "upstream usage at end",
			chunks:         []string{roleChunk, textChunk, finishChunk, usageChunk},
			wantContent:    `"text":"hello"`,
			wantStopReason: "end_turn",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := "data: " + strings.Join(tc.chunks, "\n\ndata: ") + "\n\ndata: [DONE]\n\n"
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)
			stream := true
			request := &dto.ClaudeRequest{Model: "glm-5.3", Stream: &stream}
			common.SetContextKey(c, constant.ContextKeyOriginalModel, request.Model)
			common.SetContextKey(c, constant.ContextKeyChannelType, constant.ChannelTypeVolcEngine)
			info := relaycommon.GenRelayInfoClaude(c, request)
			info.InitChannelMeta(c)
			info.DisablePing = true
			info.SetEstimatePromptTokens(64)
			response := &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			}

			usage, apiErr := OaiStreamHandler(c, info, response)

			require.Nil(t, apiErr)
			require.NotNil(t, usage)
			require.True(t, info.StreamStatus.RequestSucceeded())
			got := recorder.Body.String()
			require.Equal(t, 1, strings.Count(got, "event: message_start\n"))
			require.Equal(t, 1, strings.Count(got, "event: content_block_delta\n"))
			require.Equal(t, 1, strings.Count(got, "event: content_block_stop\n"))
			require.Equal(t, 1, strings.Count(got, "event: message_delta\n"))
			require.Equal(t, 1, strings.Count(got, "event: message_stop\n"))
			require.Contains(t, got, tc.wantContent)
			require.Contains(t, got, `"stop_reason":"`+tc.wantStopReason+`"`)
			require.True(t, strings.Index(got, "event: message_delta\n") < strings.Index(got, "event: message_stop\n"))

			var startUsage, finalUsage *dto.ClaudeUsage
			for _, line := range strings.Split(got, "\n") {
				data, ok := strings.CutPrefix(line, "data: ")
				if !ok {
					continue
				}
				var event dto.ClaudeResponse
				require.NoError(t, common.UnmarshalJsonStr(data, &event))
				switch event.Type {
				case "message_start":
					require.NotNil(t, event.Message)
					startUsage = event.Message.Usage
				case "message_delta":
					finalUsage = event.Usage
				}
			}
			require.Equal(t, &dto.ClaudeUsage{}, startUsage)
			require.NotNil(t, finalUsage)
			require.Equal(t, tc.estimated, common.GetContextKeyBool(c, constant.ContextKeyLocalCountTokens))
			require.Positive(t, finalUsage.OutputTokens)
			require.Equal(t, usage.CompletionTokens, finalUsage.OutputTokens)
			require.Equal(t, usage.PromptTokens+usage.CompletionTokens, usage.TotalTokens)
			require.NotNil(t, finalUsage.BillingUsage)
			require.NotNil(t, finalUsage.BillingUsage.OpenAIUsage)
			require.Equal(t, usage.TotalTokens, finalUsage.BillingUsage.OpenAIUsage.TotalTokens)
			require.Zero(t, finalUsage.CacheCreationInputTokens)
			if tc.estimated {
				require.Equal(t, 64, finalUsage.InputTokens)
				require.Zero(t, finalUsage.CacheReadInputTokens)
			} else {
				require.Equal(t, 19, finalUsage.InputTokens)
				require.Equal(t, 10816, finalUsage.CacheReadInputTokens)
				require.Equal(t, 2, finalUsage.OutputTokens)
			}
		})
	}
}
