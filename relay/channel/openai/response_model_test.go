package openai

import (
	"strings"
	"testing"

	relaycommon "github.com/QuantumNous/new-api/relay/common"
	relayconstant "github.com/QuantumNous/new-api/relay/constant"
	"github.com/stretchr/testify/require"
)

func TestChatStreamObservesModelWithoutChangingTokenText(t *testing.T) {
	info := &relaycommon.RelayInfo{
		OriginModelName: "public-model",
		RelayMode: relayconstant.RelayModeChatCompletions,
		ChannelMeta: &relaycommon.ChannelMeta{UpstreamModelName: "mapped-model"},
	}
	var text strings.Builder
	toolCount := 0
	require.NoError(t, processTokenData(info, `{"model":"returned-model","choices":[{"delta":{"content":"hello"}}]}`, &text, &toolCount))
	require.Equal(t, "hello", text.String())
	require.Equal(t, 0, toolCount)
	require.True(t, info.ResponseModel.Mismatch())
	require.Equal(t, "mapped-model", info.UpstreamModelName)
	// A later usage-only frame must not erase an earlier mismatch.
	require.NoError(t, processTokenData(info, `{"model":"mapped-model","choices":[]}`, &text, &toolCount))
	require.Equal(t, "returned-model", info.ResponseModel.ReturnedModel)
}
