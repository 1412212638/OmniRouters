package common_test

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/stretchr/testify/assert"
)

func TestWanEndpointsDistinguishImagesFromVideos(t *testing.T) {
	for _, name := range []string{
		"wan2.7-image-pro", "wan2.7-image", "wan2.6-image", "wan2.6-t2i",
		"wan2.5-t2i-preview", "wan2.2-t2i-flash", "wan2.2-t2i-plus",
		"wanx2.1-t2i-turbo", "wanx2.1-t2i-plus", "wanx2.0-t2i-turbo",
	} {
		t.Run(name, func(t *testing.T) {
			assert.Contains(t, common.GetEndpointTypesByChannelType(constant.ChannelTypeAli, name), constant.EndpointTypeImageGeneration)
		})
	}
	for _, name := range []string{
		"gemini-2.0-flash-exp-image-generation",
		"gemini-3-pro-image-preview",
		"gemini-3.1-flash-lite-image",
		"nano-banana-pro-preview",
	} {
		assert.Truef(t, common.IsGeminiGenerateContentImageModel(name), "expected %q to be a Gemini image generation model", name)
	}
	for _, name := range []string{
		"gemini-3-pro-image-understanding",
		"gemini-3-pro-image-preview-edit",
		"gemini-3-pro-image-generation-extra",
	} {
		assert.Falsef(t, common.IsGeminiGenerateContentImageModel(name), "expected %q not to be a Gemini image generation model", name)
		assert.Falsef(t, common.IsImageGenerationModel(name), "expected %q not to expose the image endpoint", name)
	}
	for _, name := range []string{
		"wanx2.1-t2v-plus", "wanx2.1-t2v-turbo", "wanx2.1-i2v-plus", "wanx2.1-i2v-turbo",
	} {
		t.Run(name, func(t *testing.T) {
			assert.NotContains(t, common.GetEndpointTypesByChannelType(constant.ChannelTypeAli, name), constant.EndpointTypeImageGeneration)
		})
	}
}

func TestGeminiGenerateContentImageModelsExposeImageEndpoint(t *testing.T) {
	for _, name := range []string{
		"gemini-2.5-flash-image",
		"gemini-3-pro-image",
		"gemini-3.1-flash-image",
		"gemini-3.1-flash-lite-image",
	} {
		t.Run(name, func(t *testing.T) {
			assert.True(t, common.IsImageGenerationModel(name))
			assert.Contains(t, common.GetEndpointTypesByChannelType(constant.ChannelTypeGemini, name), constant.EndpointTypeImageGeneration)
		})
	}
	assert.True(t, common.IsImageGenerationModel("grok-imagine-image"))
	for _, name := range []string{"gpt-image-2.5-sunburst", "gpt-image-2.5-flare"} {
		assert.True(t, common.IsImageGenerationModel(name))
		assert.Contains(t, common.GetEndpointTypesByChannelType(constant.ChannelTypeOpenAI, name), constant.EndpointTypeImageGeneration)
	}
}
