package common

import (
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestResponseOutcomeRetryResetsContextAndRelayState(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	info := &RelayInfo{StreamStatus: NewStreamStatus()}
	info.StreamStatus.MarkFailed("upstream", "", 502)
	common.SetContextKey(c, constant.ContextKeyResponseStreamStatus, info.StreamStatus)
	info.ResetResponseAttempt(c)
	require.Nil(t, info.StreamStatus)
	status, _ := common.GetContextKeyType[*StreamStatus](c, constant.ContextKeyResponseStreamStatus)
	require.Nil(t, status)
	require.True(t, status.RequestSucceeded(), "successful non-stream retry must not inherit stream failure")
}

func TestResponseOutcomeSuccessAccounting(t *testing.T) {
	for _, tc := range []struct {
		name string
		configure func(*StreamStatus)
		want bool
	}{
		{"legacy", func(s *StreamStatus) {}, true},
		{"completed", func(s *StreamStatus) { s.MarkCompleted() }, true},
		{"failed_after_completed", func(s *StreamStatus) { s.MarkCompleted(); s.MarkFailed("upstream", "server_error", 503) }, false},
		{"failed_before_completed", func(s *StreamStatus) { s.MarkFailed("upstream", "server_error", 503); s.MarkCompleted() }, false},
		{"cancelled", func(s *StreamStatus) { s.MarkCancelled(); s.MarkCompleted() }, false},
		{"incomplete", func(s *StreamStatus) { s.MarkIncomplete("unknown") }, false},
		{"output_limit", func(s *StreamStatus) { s.MarkIncomplete("max_output_tokens") }, true},
		{"missing_terminal", func(s *StreamStatus) { s.RequireTerminal(); s.SetEndReason(StreamEndReasonEOF, nil) }, false},
		{"done_sentinel", func(s *StreamStatus) { s.RequireTerminal(); s.SetEndReason(StreamEndReasonDone, nil) }, true},
		{"soft_error", func(s *StreamStatus) { s.MarkCompleted(); s.RecordError("decode") }, false},
		{"timeout", func(s *StreamStatus) { s.SetEndReason(StreamEndReasonTimeout, nil) }, false},
		{"client_gone", func(s *StreamStatus) { s.MarkCompleted(); s.SetEndReason(StreamEndReasonClientGone, nil) }, false},
		{"scanner_error", func(s *StreamStatus) { s.SetEndReason(StreamEndReasonScannerErr, nil) }, false},
		{"panic", func(s *StreamStatus) { s.SetEndReason(StreamEndReasonPanic, nil) }, false},
		{"ping_failure", func(s *StreamStatus) { s.SetEndReason(StreamEndReasonPingFail, nil) }, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			status := NewStreamStatus()
			tc.configure(status)
			require.Equal(t, tc.want, status.RequestSucceeded())
		})
	}
	var absent *StreamStatus
	require.True(t, absent.RequestSucceeded())
}

func TestResponseOutcomeSnapshotsAreConcurrentAndPreserveFailureDetails(t *testing.T) {
	status := NewStreamStatus()
	var wg sync.WaitGroup
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			status.RequireTerminal()
			status.SetEndReason(StreamEndReasonEOF, nil)
			status.MarkCompleted()
			status.MarkFailed("rate_limit", "upstream", 429)
			status.MarkFailed("", "", 0)
			_ = status.OutcomeSnapshot()
		}()
	}
	wg.Wait()
	outcome := status.OutcomeSnapshot()
	require.Equal(t, ResponseOutcomeFailed, outcome.Response)
	require.Equal(t, "rate_limit", outcome.ErrorCode)
	require.Equal(t, "upstream", outcome.ErrorType)
	require.Equal(t, 429, outcome.ErrorStatus)
}
