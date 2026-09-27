package controller

import (
	"testing"

	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/relaykit/dto"
)

func TestImageCapabilitiesAreProviderSpecific(t *testing.T) {
	profiles := imageCapabilityProfiles()
	tests := []struct {
		name        string
		channelType int
		model       string
		wantKeys    map[string]string
		wantMissing []string
	}{
		{
			name:        "Gemini image fields map to its compatible gateway keys",
			channelType: constant.ChannelTypeGemini,
			model:       "gemini-3.1-flash-lite-image",
			wantKeys: map[string]string{
				"aspect_ratio": "size",
				"image_size":   "quality",
			},
			wantMissing: []string{"n", "response_format", "background", "output_format"},
		},
		{
			name:        "xAI does not expose unsupported dimensions or quality",
			channelType: constant.ChannelTypeXai,
			model:       "grok-imagine-image",
			wantKeys: map[string]string{
				"response_format": "response_format",
			},
			wantMissing: []string{"n", "size", "quality", "background"},
		},
		{
			name:        "Ali parameters are nested for its adapter",
			channelType: constant.ChannelTypeAli,
			model:       "qwen-image-plus",
			wantKeys: map[string]string{
				"seed":         "parameters.seed",
				"prompt_extend": "parameters.prompt_extend",
			},
			wantMissing: []string{"n"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parameters := intersectImageCapabilities(profiles, []int{tt.channelType}, tt.model, "generation")
			byKey := make(map[string]ImageParameterCapability, len(parameters))
			for _, parameter := range parameters {
				byKey[parameter.Key] = parameter
			}
			for key, requestKey := range tt.wantKeys {
				parameter, ok := byKey[key]
				if !ok {
					t.Fatalf("expected parameter %q", key)
				}
				if parameter.RequestKey != requestKey {
					t.Errorf("%s request key = %q, want %q", key, parameter.RequestKey, requestKey)
				}
			}
			for _, key := range tt.wantMissing {
				if _, ok := byKey[key]; ok {
					t.Errorf("unsupported parameter %q was exposed", key)
				}
			}
		})
	}
}

func TestImageCapabilityIntersectionRequiresCompatibleWireFields(t *testing.T) {
	profiles := imageCapabilityProfiles()
	parameters := intersectImageCapabilities(
		profiles,
		[]int{constant.ChannelTypeAli, constant.ChannelTypeOpenAI},
		"shared-image-model",
		"generation",
	)
	for _, parameter := range parameters {
		if parameter.Key == "n" {
			t.Fatal("n must be omitted when enabled channels require different request keys")
		}
	}
}

func TestGPTImageCountCapabilityUsesProviderBound(t *testing.T) {
	profiles := imageCapabilityProfiles()
	parameters := intersectImageCapabilities(profiles, []int{constant.ChannelTypeOpenAI}, "gpt-image-2.5-sunburst", "generation")
	for _, parameter := range parameters {
		if parameter.Key != "n" {
			continue
		}
		if parameter.Max == nil || *parameter.Max != dto.MaxGPTImageN {
			t.Fatalf("n max = %v, want %d", parameter.Max, dto.MaxGPTImageN)
		}
		return
	}
	t.Fatal("n capability was not exposed")
}

func TestGPTImageCapabilitiesExposeExtendedParameters(t *testing.T) {
	parameters := intersectImageCapabilities(
		imageCapabilityProfiles(),
		[]int{constant.ChannelTypeOpenAI},
		"gpt-image-2.5-flare",
		"generation",
	)
	byKey := make(map[string]ImageParameterCapability, len(parameters))
	for _, parameter := range parameters {
		byKey[parameter.Key] = parameter
	}

	quality := byKey["quality"]
	if quality.Type != "enum" || !containsString(quality.Options, "xhigh") || !containsString(quality.Options, "max") {
		t.Fatalf("GPT quality options = %#v, want xhigh and max", quality.Options)
	}
	size := byKey["size"]
	if size.Type != "string" || !containsString(size.Options, "3840x2160") {
		t.Fatalf("GPT size capability = %#v, want editable string with common sizes", size)
	}
}

func TestNonGPTImageCapabilitiesDoNotExposeCount(t *testing.T) {
	profiles := imageCapabilityProfiles()
	tests := []struct {
		name        string
		channelType int
		model       string
	}{
		{name: "DALL-E", channelType: constant.ChannelTypeOpenAI, model: "dall-e-3"},
		{name: "MiniMax", channelType: constant.ChannelTypeMiniMax, model: "image-01"},
		{name: "SiliconFlow", channelType: constant.ChannelTypeSiliconFlow, model: "black-forest-labs/FLUX.1-schnell"},
		{name: "Replicate", channelType: constant.ChannelTypeReplicate, model: "black-forest-labs/flux-schnell"},
		{name: "VolcEngine", channelType: constant.ChannelTypeVolcEngine, model: "doubao-seedream"},
		{name: "unknown provider", channelType: 9999, model: "custom-image-model"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parameters := intersectImageCapabilities(profiles, []int{test.channelType}, test.model, "generation")
			for _, parameter := range parameters {
				if parameter.Key == "n" {
					t.Fatal("non-GPT image model must not expose n")
				}
			}
		})
	}
}
