package relay

import (
	"encoding/json"
	"testing"

	"github.com/QuantumNous/new-api/constant"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	relayconstant "github.com/QuantumNous/new-api/relay/constant"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/stretchr/testify/require"
)

func TestMergePlaygroundImageExtraFieldsProtectsRoutingFields(t *testing.T) {
	body := []byte(`{"model":"image-model","quality":"auto"}`)
	extra := map[string]json.RawMessage{
		"model":   json.RawMessage(`"other-model"`),
		"group":   json.RawMessage(`"other-group"`),
		"quality": json.RawMessage(`"high"`),
		"seed":    json.RawMessage(`42`),
	}
	merged, err := mergePlaygroundImageExtraFields(body, extra)
	require.NoError(t, err)
	var got map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(merged, &got))
	require.JSONEq(t, `"image-model"`, string(got["model"]))
	require.JSONEq(t, `"high"`, string(got["quality"]))
	require.JSONEq(t, `42`, string(got["seed"]))
	_, hasGroup := got["group"]
	require.False(t, hasGroup)
}

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

func TestPlaygroundImageEditAlwaysUsesAdaptorConversion(t *testing.T) {
	info := &relaycommon.RelayInfo{
		IsPlayground: true,
		RelayMode:    relayconstant.RelayModeImagesEdits,
		ChannelMeta: &relaycommon.ChannelMeta{
			ApiType:       constant.APITypeOpenAI,
			ChannelSetting: dto.ChannelSettings{PassThroughBodyEnabled: true},
		},
	}
	require.False(t, shouldPassThroughImageRequest(info, true))
	info.RelayMode = relayconstant.RelayModeImagesGenerations
	require.False(t, shouldPassThroughImageRequest(info, true))
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
