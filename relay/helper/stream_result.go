package helper

import (
	"errors"

	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relaykit/types"
)

// StreamResult is passed to each dataHandler invocation, providing methods
// to record soft errors, signal fatal stops, or mark normal completion.
// StreamScannerHandler checks IsStopped() after each callback invocation.
type StreamResult struct {
	status  *relaycommon.StreamStatus
	stopped bool
	done    bool
}

func newStreamResult(status *relaycommon.StreamStatus) *StreamResult {
	return &StreamResult{status: status}
}

// Error records a soft error and reports whether it counted it separately.
// The stream continues processing, including after a downstream disconnect.
func (r *StreamResult) Error(err error) bool {
	if err == nil {
		return false
	}
	var apiErr *types.NewAPIError
	if errors.As(err, &apiErr) {
		r.status.MarkFailed(string(apiErr.GetErrorCode()), apiErr.ToOpenAIError().Type, apiErr.StatusCode)
	}
	if errors.Is(err, errStreamRequestContextDone) {
		return r.status.RecordDownstreamContextError(err.Error())
	}
	r.status.RecordError(err.Error())
	return true
}

// Stop records a fatal error and marks the stream to stop after this chunk.
func (r *StreamResult) Stop(err error) {
	r.Error(err)
	r.status.SetEndReason(relaycommon.StreamEndReasonHandlerStop, err)
	r.stopped = true
}

// Done signals that the handler has finished processing normally
// (e.g., Dify "message_end"). The stream stops after this chunk.
func (r *StreamResult) Done() {
	r.status.MarkCompleted()
	r.status.SetEndReason(relaycommon.StreamEndReasonDone, nil)
	r.stopped = true
	r.done = true
}

// IsStopped returns whether Stop() or Done() was called during this chunk.
func (r *StreamResult) IsStopped() bool {
	return r.stopped
}

// reset clears the per-chunk stopped flag so the object can be reused.
func (r *StreamResult) reset() {
	r.stopped = false
}
