package dto

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/require"
)

func TestClaudeRequestPreservesSafeguardsAndMessageOutputConfig(t *testing.T) {
	raw := `{"model":"claude-opus-5-5","max_tokens":64,"safeguards":{"enabled":true},"messages":[{"role":"system","content":[],"output_config":{"effort":"high"}}]}`

	var request ClaudeRequest
	require.NoError(t, common.UnmarshalJsonStr(raw, &request))
	require.Len(t, request.Messages, 1)

	encoded, err := common.Marshal(request)
	require.NoError(t, err)

	var roundTripped ClaudeRequest
	require.NoError(t, common.Unmarshal(encoded, &roundTripped))
	require.JSONEq(t, `{"enabled":true}`, string(roundTripped.Safeguards))
	require.JSONEq(t, `{"effort":"high"}`, string(roundTripped.Messages[0].OutputConfig))
}
