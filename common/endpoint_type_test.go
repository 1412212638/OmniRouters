package common

import (
	"testing"

	"github.com/QuantumNous/new-api/constant"
)

func TestIsOpenAIChatEndpointModel(t *testing.T) {
	tests := []struct {
		name        string
		channelType int
		model       string
		want        bool
	}{
		{
			name:        "openai chat model",
			channelType: constant.ChannelTypeOpenAI,
			model:       "gpt-4o",
			want:        true,
		},
		{
			name:        "anthropic through openai compatible endpoint",
			channelType: constant.ChannelTypeAnthropic,
			model:       "claude-3-5-sonnet",
			want:        true,
		},
		{
			name:        "openai image model",
			channelType: constant.ChannelTypeOpenAI,
			model:       "gpt-image-1",
			want:        false,
		},
		{
			name:        "openai embedding model",
			channelType: constant.ChannelTypeOpenAI,
			model:       "text-embedding-3-small",
			want:        false,
		},
		{
			name:        "jina rerank channel",
			channelType: constant.ChannelTypeJina,
			model:       "jina-reranker-v2-base-multilingual",
			want:        false,
		},
		{
			name:        "kling video channel",
			channelType: constant.ChannelTypeKling,
			model:       "Kling-3.0",
			want:        false,
		},
		{
			name:        "hailuo video model on minimax",
			channelType: constant.ChannelTypeMiniMax,
			model:       "MiniMax-Hailuo-2.3",
			want:        false,
		},
		{
			name:        "minimax chat model",
			channelType: constant.ChannelTypeMiniMax,
			model:       "abab6.5s-chat",
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsOpenAIChatEndpointModel(tt.channelType, tt.model)
			if got != tt.want {
				t.Fatalf("IsOpenAIChatEndpointModel(%d, %q) = %v, want %v", tt.channelType, tt.model, got, tt.want)
			}
		})
	}
}
