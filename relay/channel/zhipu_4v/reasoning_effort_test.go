package zhipu_4v

import (
	"testing"

	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/stretchr/testify/require"
)

func TestZhipuPreservesReasoningEffort(t *testing.T) {
	for _, effort := range []string{"", "none", "low", "high"} {
		result := requestOpenAI2Zhipu(dto.GeneralOpenAIRequest{Model: "glm-4.5", ReasoningEffort: effort})
		require.Equal(t, effort, result.ReasoningEffort)
	}
}
