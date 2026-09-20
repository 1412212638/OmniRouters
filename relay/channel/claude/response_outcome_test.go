package claude

import (
	"net/http/httptest"
	"testing"

	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestClaudeResponseOutcomeRequiresMessageStop(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	info := &relaycommon.RelayInfo{StreamStatus: relaycommon.NewStreamStatus()}
	info.StreamStatus.RequireTerminal()
	info.StreamStatus.SetEndReason(relaycommon.StreamEndReasonEOF, nil)
	require.Nil(t, HandleStreamResponseData(c, info, &ClaudeResponseInfo{}, `{"type":"ping"}`))
	require.False(t, info.StreamStatus.RequestSucceeded())
	require.Nil(t, HandleStreamResponseData(c, info, &ClaudeResponseInfo{}, `{"type":"message_stop"}`))
	require.True(t, info.StreamStatus.RequestSucceeded())
}
