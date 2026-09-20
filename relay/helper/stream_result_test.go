package helper

import (
	"fmt"
	"net/http"
	"testing"

	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relaykit/types"
	"github.com/stretchr/testify/require"
)

func TestStreamResultAPIErrorOverridesCompletion(t *testing.T) {
	status := relaycommon.NewStreamStatus()
	result := newStreamResult(status)
	result.Done()
	apiErr := types.NewOpenAIError(fmt.Errorf("private upstream content"), types.ErrorCodeBadResponse, http.StatusBadGateway)
	result.Stop(fmt.Errorf("wrapped: %w", apiErr))
	require.True(t, result.IsStopped())
	require.True(t, status.ResponseFailed())
	require.False(t, status.RequestSucceeded())
	require.Equal(t, http.StatusBadGateway, status.OutcomeSnapshot().ErrorStatus)
	require.Equal(t, 1, status.TotalErrorCount())
}
