package relay

import (
	"testing"

	"github.com/QuantumNous/new-api/constant"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	relayconstant "github.com/QuantumNous/new-api/relay/constant"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/stretchr/testify/require"
)

func TestValidateGeminiGenerateContentImageRequestUsesMappedModel(t *testing.T) {
	info := &relaycommon.RelayInfo{
		RelayMode:       relayconstant.RelayModeImagesGenerations,
		OriginModelName: "image-model",
		ChannelMeta: &relaycommon.ChannelMeta{
			ApiType:          constant.APITypeGemini,
			UpstreamModelName: "gemini-3.1-flash-lite-image",
		},
	}
	n := uint(2)
	require.ErrorContains(t, validateGeminiGenerateContentImageRequest(info, &dto.ImageRequest{N: &n}, 2), "one candidate")

	stream := true
	require.ErrorContains(t, validateGeminiGenerateContentImageRequest(info, &dto.ImageRequest{Stream: &stream}, 1), "streaming")
	require.NoError(t, validateGeminiGenerateContentImageRequest(info, &dto.ImageRequest{}, 1))
}

func TestGeminiImageEditAlwaysUsesAdaptorConversion(t *testing.T) {
	info := &relaycommon.RelayInfo{
		RelayMode: relayconstant.RelayModeImagesEdits,
		ChannelMeta: &relaycommon.ChannelMeta{
			ApiType:          constant.APITypeGemini,
			UpstreamModelName: "gemini-3.1-flash-lite-image",
			ChannelSetting:    dto.ChannelSettings{PassThroughBodyEnabled: true},
		},
	}
	require.False(t, shouldPassThroughImageRequest(info, true))
	require.ErrorContains(
		t,
		validateGeminiGenerateContentImageRequest(info, &dto.ImageRequest{}, 1),
		"input image",
	)
}

func TestGeminiImageAlwaysUsesAdaptorConversion(t *testing.T) {
	info := &relaycommon.RelayInfo{
		IsPlayground: true,
		RelayMode:    relayconstant.RelayModeImagesGenerations,
		ChannelMeta: &relaycommon.ChannelMeta{
			ApiType:          constant.APITypeGemini,
			UpstreamModelName: "gemini-3.1-flash-lite-image",
			ChannelSetting:    dto.ChannelSettings{PassThroughBodyEnabled: true},
		},
	}
	require.False(t, shouldPassThroughImageRequest(info, true))

	info.IsPlayground = false
	require.False(t, shouldPassThroughImageRequest(info, true))
}

func TestGeminiImageAlwaysUsesAdaptorConversionForStandardImageRoute(t *testing.T) {
	info := &relaycommon.RelayInfo{
		RelayMode: relayconstant.RelayModeImagesGenerations,
		ChannelMeta: &relaycommon.ChannelMeta{
			ApiType:           constant.APITypeGemini,
			UpstreamModelName: "gemini-3.1-flash-lite-image",
			ChannelSetting:    dto.ChannelSettings{PassThroughBodyEnabled: true},
		},
	}
	require.False(t, shouldPassThroughImageRequest(info, true))

	info.ChannelMeta.ApiType = constant.APITypeVertexAi
	require.False(t, shouldPassThroughImageRequest(info, true))

	info.ChannelMeta.ApiType = constant.APITypeOpenAI
	require.True(t, shouldPassThroughImageRequest(info, true))
}
