package openai

import (
	"testing"

	"github.com/QuantumNous/new-api/constant"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/stretchr/testify/require"
	"github.com/samber/lo"
)

func TestConvertOpenAIRequestUsesModelSpecificGPT6Capabilities(t *testing.T) {
	for _, tt := range []struct {
		name, model, effort string
		useMaxTokens       bool
		keepSampling       bool
		wantModel          string
	}{
		{name: "sol keeps sampling", model: "gpt-6-sol", useMaxTokens: true, keepSampling: true, wantModel: "gpt-6-sol"},
		{name: "luna effort removes sampling", model: "gpt-6-luna-high", useMaxTokens: true, wantModel: "gpt-6-luna"},
		{name: "future generation stays compatible", model: "gpt-7-high", keepSampling: true, wantModel: "gpt-7-high"},
		{name: "custom suffix stays compatible", model: "custom-qwen-high", keepSampling: true, wantModel: "custom-qwen-high"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			request := &dto.GeneralOpenAIRequest{
				Model: tt.model, MaxTokens: lo.ToPtr(uint(128)),
				Temperature: lo.ToPtr(0.7), TopP: lo.ToPtr(0.8), LogProbs: lo.ToPtr(true),
			}
			info := &relaycommon.RelayInfo{
				ChannelType:      constant.ChannelTypeOpenAI,
				OriginModelName:  tt.model,
				ChannelMeta:      &relaycommon.ChannelMeta{UpstreamModelName: tt.model},
			}
			_, err := (&Adaptor{}).ConvertOpenAIRequest(nil, info, request)
			require.NoError(t, err)
			require.Equal(t, tt.wantModel, request.Model)
			if tt.useMaxTokens {
				require.NotNil(t, request.MaxCompletionTokens)
				require.Nil(t, request.MaxTokens)
			} else {
				require.Nil(t, request.MaxCompletionTokens)
				require.NotNil(t, request.MaxTokens)
			}
			if tt.keepSampling {
				require.NotNil(t, request.Temperature)
				require.NotNil(t, request.TopP)
				require.NotNil(t, request.LogProbs)
			} else {
				require.Nil(t, request.Temperature)
				require.Nil(t, request.TopP)
				require.Nil(t, request.LogProbs)
			}
		})
	}
}
