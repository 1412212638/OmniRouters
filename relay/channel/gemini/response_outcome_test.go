package gemini

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/constant"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestGeminiStreamOutcomeTracksFinishAndLateError(t *testing.T) {
	previous := constant.StreamingTimeout
	constant.StreamingTimeout = 30
	t.Cleanup(func() { constant.StreamingTimeout = previous })
	for _, tc := range []struct {
		name string
		body string
		want bool
	}{
		{"finished", `{"candidates":[{"finishReason":"STOP"}],"usageMetadata":{"promptTokenCount":1,"candidatesTokenCount":1,"totalTokenCount":2}}`, true},
		{"missing_finish", `{"usageMetadata":{"promptTokenCount":1,"totalTokenCount":1}}`, false},
		{"late_error", "{\"candidates\":[{\"finishReason\":\"STOP\"}],\"usageMetadata\":{\"promptTokenCount\":1,\"totalTokenCount\":1}}\n\ndata: {\"error\":{\"status\":\"INTERNAL\"}}", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
			info := &relaycommon.RelayInfo{DisablePing: true, ChannelMeta: &relaycommon.ChannelMeta{UpstreamModelName: "gemini-test"}}
			response := &http.Response{Body: io.NopCloser(strings.NewReader("data: " + tc.body + "\n\n"))}
			usage, apiErr := geminiStreamHandler(c, info, response, func(string, *dto.GeminiChatResponse) bool { return true })
			require.Nil(t, apiErr)
			require.NotNil(t, usage)
			require.Equal(t, tc.want, info.StreamStatus.RequestSucceeded())
		})
	}
}
