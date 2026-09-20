package service

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/stretchr/testify/require"
)

func TestObserveResponsesOutcome(t *testing.T) {
	for _, tc := range []struct {
		input string
		want relaycommon.ResponseOutcome
		success bool
	}{
		{`{"type":"response.completed"}`, relaycommon.ResponseOutcomeCompleted, true},
		{`{"type":"response.done"}`, relaycommon.ResponseOutcomeCompleted, true},
		{`{"type":"response.failed","response":{"error":{"code":"limit","type":"rate_limit","message":"private content"}}}`, relaycommon.ResponseOutcomeFailed, false},
		{`{"type":"error","code":"server_error","message":"private content"}`, relaycommon.ResponseOutcomeFailed, false},
		{`{"type":"response.error"}`, relaycommon.ResponseOutcomeFailed, false},
		{`{"type":"response.incomplete","response":{"incomplete_details":{"reason":"max_output_tokens"}}}`, relaycommon.ResponseOutcomeIncomplete, true},
		{`{"type":"response.incomplete"}`, relaycommon.ResponseOutcomeIncomplete, false},
		{`{"type":"response.cancelled"}`, relaycommon.ResponseOutcomeCancelled, false},
		{`{"type":"response.canceled"}`, relaycommon.ResponseOutcomeCancelled, false},
		{`{"response":{"status":"failed"}}`, relaycommon.ResponseOutcomeFailed, false},
		{`{"response":{"status":"canceled"}}`, relaycommon.ResponseOutcomeCancelled, false},
	} {
		t.Run(tc.input, func(t *testing.T) {
			var event dto.ResponsesStreamResponse
			require.NoError(t, common.UnmarshalJsonStr(tc.input, &event))
			info := &relaycommon.RelayInfo{StreamStatus: relaycommon.NewStreamStatus()}
			ObserveResponsesOutcome(info, &event)
			require.Equal(t, tc.want, info.StreamStatus.OutcomeSnapshot().Response)
			require.Equal(t, tc.success, info.StreamStatus.RequestSucceeded())
			encoded, err := common.Marshal(info.StreamStatus.OutcomeSnapshot())
			require.NoError(t, err)
			require.NotContains(t, string(encoded), "private content")
		})
	}
}
