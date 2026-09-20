package openai

import (
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relaykit/dto"
)

// Observe upstream finish reasons, never terminal events synthesized by a converter.
func observeChatOutcome(info *relaycommon.RelayInfo, chunk *dto.ChatCompletionsStreamResponse) {
	if info == nil || chunk == nil {
		return
	}
	for _, choice := range chunk.Choices {
		if choice.FinishReason != nil && *choice.FinishReason != "" {
			info.StreamStatus.MarkCompleted()
		}
	}
}
