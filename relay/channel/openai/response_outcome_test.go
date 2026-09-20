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
	relayconstant "github.com/QuantumNous/new-api/relay/constant"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestOpenAIResponsesOutcomePublishedFromActualStream(t *testing.T) {
	previous := constant.StreamingTimeout
	constant.StreamingTimeout = 30
	t.Cleanup(func() { constant.StreamingTimeout = previous })
	for _, tc := range []struct {
		name string
		body string
		want bool
	}{
		{"completed", "data: {\"type\":\"response.completed\"}\n\n", true},
		{"failed", "data: {\"type\":\"response.failed\"}\n\ndata: [DONE]\n\n", false},
		{"cut", "data: {\"type\":\"response.created\"}\n\n", false},
		{"cancelled", "data: {\"type\":\"response.cancelled\"}\n\n", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
			info := &relaycommon.RelayInfo{DisablePing: true, ChannelMeta: &relaycommon.ChannelMeta{UpstreamModelName: "test"}}
			response := &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(tc.body)), Header: http.Header{}}
			_, apiErr := OaiResponsesStreamHandler(c, info, response)
			require.Nil(t, apiErr)
			status, ok := common.GetContextKeyType[*relaycommon.StreamStatus](c, constant.ContextKeyResponseStreamStatus)
			require.True(t, ok)
			require.Same(t, info.StreamStatus, status)
			require.Equal(t, tc.want, status.RequestSucceeded())
		})
	}
}

func TestOpenAIStreamOutcomePreservesTextAndFailureWins(t *testing.T) {
	info := &relaycommon.RelayInfo{
		RelayMode: relayconstant.RelayModeChatCompletions,
		StreamStatus: relaycommon.NewStreamStatus(),
		ChannelMeta: &relaycommon.ChannelMeta{},
	}
	var text strings.Builder
	var tools int
	require.NoError(t, processTokenData(info, `{"choices":[{"delta":{"content":"hello"}}]}`, &text, &tools))
	require.Empty(t, info.StreamStatus.ResponseOutcome())
	require.NoError(t, processTokenData(info, `{"choices":[{"delta":{},"finish_reason":"stop"}]}`, &text, &tools))
	require.True(t, info.StreamStatus.RequestSucceeded())
	require.NoError(t, processTokenData(info, `{"error":{"code":"upstream","type":"server_error","message":"private content"}}`, &text, &tools))
	require.False(t, info.StreamStatus.RequestSucceeded())
	require.Equal(t, "hello", text.String())
	require.Equal(t, 0, tools)
	require.Equal(t, "upstream", info.StreamStatus.OutcomeSnapshot().ErrorCode)
}

func TestOpenAICompletionOutcome(t *testing.T) {
	info := &relaycommon.RelayInfo{RelayMode: relayconstant.RelayModeCompletions, StreamStatus: relaycommon.NewStreamStatus()}
	var text strings.Builder
	var tools int
	require.NoError(t, processTokenData(info, `{"choices":[{"text":"done","finish_reason":"length"}]}`, &text, &tools))
	require.Equal(t, "completed", info.StreamStatus.ResponseOutcome())
	require.Equal(t, "done", text.String())
}
