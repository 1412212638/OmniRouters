package reasoning

import (
	"testing"

	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/stretchr/testify/require"
)

func TestGeminiThinkingLevelCaseInsensitive(t *testing.T) {
	for _, level := range []string{"high", "HIGH", "High"} {
		effort, err := ValidateGeminiThinkingConfig("gemini-3-pro-preview", &dto.GeminiThinkingConfig{ThinkingLevel: level})
		require.NoError(t, err)
		require.Equal(t, EffortHigh, effort)
	}
	_, err := ValidateGeminiThinkingConfig("gemini-3-pro-preview", &dto.GeminiThinkingConfig{ThinkingLevel: "MEDIUM"})
	require.Error(t, err, "case normalization must not enable unsupported levels")
}
