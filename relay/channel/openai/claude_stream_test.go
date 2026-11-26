package openai

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/constant"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	relayconstant "github.com/QuantumNous/new-api/relay/constant"
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
			name:          "finish reason already emitted",
			data:          `{"id":"chatcmpl_1","model":"glm-5.3","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}`,
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
				RelayFormat:       types.RelayFormatClaude,
				ChannelMeta:       &relaycommon.ChannelMeta{UpstreamModelName: "glm-5.3"},
				StreamStatus:      relaycommon.NewStreamStatus(),
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

func TestOaiStreamHandlerFinalizesClaudeStreamAfterDoneWithoutFinishReason(t *testing.T) {
	oldMode := gin.Mode()
	gin.SetMode(gin.TestMode)
	t.Cleanup(func() { gin.SetMode(oldMode) })
	oldTimeout := constant.StreamingTimeout
	constant.StreamingTimeout = 30
	t.Cleanup(func() { constant.StreamingTimeout = oldTimeout })

	body := "data: {\"id\":\"chatcmpl_1\",\"object\":\"chat.completion.chunk\",\"created\":1710000000,\"model\":\"glm-5.3\",\"choices\":[{\"index\":0,\"delta\":{\"content\":\"hello\"},\"finish_reason\":null}]}\n\ndata: [DONE]\n\n"
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)
	info := &relaycommon.RelayInfo{
		RelayFormat:       types.RelayFormatClaude,
		RelayMode:         relayconstant.RelayModeChatCompletions,
		ChannelMeta:       &relaycommon.ChannelMeta{UpstreamModelName: "glm-5.3"},
		ClaudeConvertInfo: &relaycommon.ClaudeConvertInfo{
			LastMessagesType: relaycommon.LastMessageTypeNone,
		},
		DisablePing:       true,
	}
	response := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
	}

	_, apiErr := OaiStreamHandler(c, info, response)

	require.Nil(t, apiErr)
	got := recorder.Body.String()
	require.Equal(t, 1, strings.Count(got, "event: message_start\n"))
	require.Equal(t, 1, strings.Count(got, "event: content_block_delta\n"))
	require.Equal(t, 1, strings.Count(got, "event: content_block_stop\n"))
	require.Equal(t, 1, strings.Count(got, "event: message_delta\n"))
	require.Equal(t, 1, strings.Count(got, "event: message_stop\n"))
	require.Contains(t, got, `"text":"hello"`)
	require.Contains(t, got, `"stop_reason":"end_turn"`)
	require.True(t, strings.Index(got, "event: message_delta\n") < strings.Index(got, "event: message_stop\n"))
	require.True(t, info.StreamStatus.RequestSucceeded())
}
