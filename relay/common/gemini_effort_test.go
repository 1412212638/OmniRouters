package common

import (
	"testing"

	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/stretchr/testify/require"
)

func TestGeminiRequestEffortCanonicalizesKnownLevels(t *testing.T) {
	for _, level := range []string{"high", "HIGH", "High"} {
		require.Equal(t, "high", geminiRequestReasoningEffort(&dto.GeminiThinkingConfig{ThinkingLevel: level}))
	}
	require.Empty(t, geminiRequestReasoningEffort(nil))
	require.Equal(t, "future-level", geminiRequestReasoningEffort(&dto.GeminiThinkingConfig{ThinkingLevel: "future-level"}))
}
