package controller

import (
	"net/http"
	"reflect"
	"sort"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/service"

	"github.com/gin-gonic/gin"
)

// ImageParameterCapability describes a parameter that is safe to expose for a
// model. RequestKey may be dotted when the upstream adapter expects a nested
// object (for example parameters.seed).
type ImageParameterCapability struct {
	Key              string   `json:"key"`
	RequestKey       string   `json:"request_key"`
	Type             string   `json:"type"`
	Default          any      `json:"default,omitempty"`
	Min              *float64 `json:"min,omitempty"`
	Max              *float64 `json:"max,omitempty"`
	Step             *float64 `json:"step,omitempty"`
	Options          []string `json:"options,omitempty"`
	EnabledByDefault bool     `json:"enabled_by_default"`
	Operations       []string `json:"operations"`
}

type imageCapabilityProfile struct {
	ChannelTypes []int
	Match        func(string) bool
	Parameters   []ImageParameterCapability
}

func imageNumber(value float64) *float64 { return &value }

func imageParameter(key, requestKey, parameterType string, defaultValue any, operations []string) ImageParameterCapability {
	return ImageParameterCapability{
		Key:              key,
		RequestKey:       requestKey,
		Type:             parameterType,
		Default:          defaultValue,
		EnabledByDefault: true,
		Operations:       operations,
	}
}

func imageEnum(key, requestKey string, defaultValue string, options []string, operations []string) ImageParameterCapability {
	parameter := imageParameter(key, requestKey, "enum", defaultValue, operations)
	parameter.Options = options
	return parameter
}

func imageInteger(key, requestKey string, defaultValue, min, max, step float64, operations []string) ImageParameterCapability {
	parameter := imageParameter(key, requestKey, "integer", defaultValue, operations)
	parameter.Min = imageNumber(min)
	parameter.Max = imageNumber(max)
	parameter.Step = imageNumber(step)
	return parameter
}

func imageNumberParameter(key, requestKey string, defaultValue, min, max, step float64, operations []string) ImageParameterCapability {
	parameter := imageParameter(key, requestKey, "number", defaultValue, operations)
	parameter.Min = imageNumber(min)
	parameter.Max = imageNumber(max)
	parameter.Step = imageNumber(step)
	return parameter
}

func imageBoolean(key, requestKey string, defaultValue bool, operations []string) ImageParameterCapability {
	return imageParameter(key, requestKey, "boolean", defaultValue, operations)
}

func imageOperations() []string { return []string{"generation", "edit"} }

func imageCapabilityProfiles() []imageCapabilityProfile {
	operations := imageOperations()
	standard := []ImageParameterCapability{
		imageInteger("n", "n", 1, 1, dto.MaxImageN, 1, operations),
		imageEnum("size", "size", "1024x1024", []string{"1024x1024", "1536x1024", "1024x1536", "1792x1024", "1024x1792"}, operations),
		imageEnum("quality", "quality", "auto", []string{"auto", "low", "medium", "high"}, operations),
	}
	openAIImage := append([]ImageParameterCapability{}, standard...)
	openAIImage = append(openAIImage,
		imageEnum("background", "background", "auto", []string{"auto", "opaque", "transparent"}, operations),
		imageEnum("output_format", "output_format", "png", []string{"png", "jpeg", "webp"}, operations),
		imageInteger("output_compression", "output_compression", 100, 0, 100, 1, operations),
	)
	dallE := append([]ImageParameterCapability{}, standard...)
	dallE = append(dallE,
		imageEnum("response_format", "response_format", "url", []string{"url", "b64_json"}, operations),
		imageEnum("style", "style", "vivid", []string{"vivid", "natural"}, operations),
	)
	countOnly := []ImageParameterCapability{
		imageInteger("n", "n", 1, 1, dto.MaxImageN, 1, operations),
	}
	return []imageCapabilityProfile{
		{ChannelTypes: []int{constant.ChannelTypeGemini, constant.ChannelTypeVertexAi}, Match: func(name string) bool {
			lower := strings.ToLower(name)
			return strings.HasPrefix(lower, "gemini-") || strings.HasPrefix(lower, "imagen") || strings.HasPrefix(lower, "nano-banana")
		}, Parameters: []ImageParameterCapability{
			imageInteger("n", "n", 1, 1, 1, 1, operations),
			// Gemini's adapter accepts the OpenAI size field and maps it to aspectRatio.
			imageEnum("aspect_ratio", "size", "1:1", []string{"1:1", "3:2", "2:3", "16:9", "9:16", "4:3", "3:4", "21:9"}, operations),
			// Gemini's adapter maps quality to imageSize (1K/2K/4K).
			imageEnum("image_size", "quality", "auto", []string{"auto", "2K", "4K"}, operations),
		},},
		{ChannelTypes: []int{constant.ChannelTypeXai}, Parameters: []ImageParameterCapability{
			imageInteger("n", "n", 1, 1, dto.MaxImageN, 1, operations),
			imageEnum("response_format", "response_format", "url", []string{"url", "b64_json"}, operations),
		}},
		{ChannelTypes: []int{constant.ChannelTypeMiniMax}, Parameters: []ImageParameterCapability{
			imageInteger("n", "n", 1, 1, dto.MaxImageN, 1, operations),
			imageEnum("aspect_ratio", "aspect_ratio", "1:1", []string{"1:1", "16:9", "9:16", "4:3", "3:4", "3:2", "2:3", "21:9"}, operations),
			imageBoolean("prompt_optimizer", "prompt_optimizer", false, operations),
			imageBoolean("watermark", "watermark", false, operations),
		}},
		{ChannelTypes: []int{constant.ChannelTypeAli}, Parameters: []ImageParameterCapability{
			imageInteger("n", "parameters.n", 1, 1, dto.MaxImageN, 1, operations),
			imageEnum("size", "parameters.size", "1024*1024", []string{"1024*1024", "1280*720", "720*1280", "1536*1024", "1024*1536"}, operations),
			imageBoolean("watermark", "parameters.watermark", false, operations),
			imageBoolean("prompt_extend", "parameters.prompt_extend", true, operations),
			imageBoolean("thinking_mode", "parameters.thinking_mode", false, operations),
			imageInteger("seed", "parameters.seed", 0, 0, 2147483647, 1, operations),
		}},
		{ChannelTypes: []int{constant.ChannelTypeSiliconFlow}, Parameters: []ImageParameterCapability{
			imageInteger("n", "n", 1, 1, dto.MaxImageN, 1, operations),
			imageEnum("image_size", "image_size", "1024x1024", []string{"512x512", "768x768", "1024x1024", "1280x720", "720x1280"}, operations),
			imageInteger("seed", "seed", 0, 0, 2147483647, 1, operations),
			imageInteger("num_inference_steps", "num_inference_steps", 20, 1, 100, 1, operations),
			imageNumberParameter("guidance_scale", "guidance_scale", 7.5, 0, 30, 0.5, operations),
			imageParameter("negative_prompt", "negative_prompt", "string", "", operations),
		}},
		{ChannelTypes: []int{constant.ChannelTypeReplicate}, Parameters: []ImageParameterCapability{
			imageInteger("n", "n", 1, 1, dto.MaxImageN, 1, operations),
			imageEnum("size", "size", "1024x1024", []string{"1024x1024", "1536x1024", "1024x1536", "1792x1024", "1024x1792"}, operations),
			imageEnum("quality", "quality", "auto", []string{"auto", "low", "medium", "high"}, operations),
			imageEnum("output_format", "output_format", "png", []string{"png", "jpeg", "webp"}, operations),
		}},
		{ChannelTypes: []int{constant.ChannelTypeOpenAI, constant.ChannelTypeOpenAIMax, constant.ChannelTypeAzure, constant.ChannelTypeOpenRouter}, Match: func(name string) bool {
			return strings.HasPrefix(strings.ToLower(name), "gpt-image-")
		}, Parameters: openAIImage},
		{ChannelTypes: []int{constant.ChannelTypeOpenAI, constant.ChannelTypeOpenAIMax, constant.ChannelTypeAzure, constant.ChannelTypeOpenRouter}, Match: func(name string) bool {
			lower := strings.ToLower(name)
			return strings.HasPrefix(lower, "dall-e-") || lower == "dall-e"
		}, Parameters: dallE},
		{ChannelTypes: []int{constant.ChannelTypeVolcEngine}, Parameters: []ImageParameterCapability{
			imageInteger("n", "n", 1, 1, dto.MaxImageN, 1, operations),
			imageEnum("size", "size", "1024x1024", []string{"1024x1024", "1536x1024", "1024x1536", "1792x1024", "1024x1792"}, operations),
		}},
		{ChannelTypes: []int{constant.ChannelTypeJimeng}, Parameters: []ImageParameterCapability{
			imageInteger("seed", "extra_fields.seed", -1, -1, 2147483647, 1, operations),
			imageInteger("width", "extra_fields.width", 512, 256, 768, 64, operations),
			imageInteger("height", "extra_fields.height", 512, 256, 768, 64, operations),
			imageBoolean("use_pre_llm", "extra_fields.use_pre_llm", true, operations),
			imageBoolean("use_sr", "extra_fields.use_sr", true, operations),
		}},
		{Parameters: countOnly},
	}
}

func imageProfileMatches(profile imageCapabilityProfile, channelType int, modelName string) bool {
	if profile.Match != nil && !profile.Match(modelName) {
		return false
	}
	if len(profile.ChannelTypes) == 0 {
		return profile.Match == nil || profile.Match(modelName)
	}
	for _, supported := range profile.ChannelTypes {
		if supported == channelType {
			return profile.Match == nil || profile.Match(modelName)
		}
	}
	return false
}

func intersectImageCapabilities(profiles []imageCapabilityProfile, channelTypes []int, modelName, operation string) []ImageParameterCapability {
	if len(profiles) == 0 {
		return nil
	}
	sets := make([]map[string]ImageParameterCapability, 0, len(profiles))
	for _, channelType := range channelTypes {
		var matched []ImageParameterCapability
		for _, profile := range profiles {
			if imageProfileMatches(profile, channelType, modelName) {
				matched = profile.Parameters
				break
			}
		}
		if len(matched) == 0 {
			continue
		}
		set := make(map[string]ImageParameterCapability)
		for _, parameter := range matched {
			if containsString(parameter.Operations, operation) {
				set[parameter.Key] = parameter
			}
		}
		sets = append(sets, set)
	}
	if len(sets) == 0 {
		return nil
	}
	result := make([]ImageParameterCapability, 0, len(sets[0]))
	for key, parameter := range sets[0] {
		present := true
		for _, set := range sets[1:] {
			candidate, ok := set[key]
			if !ok || !sameImageCapability(parameter, candidate) {
				present = false
				break
			}
		}
		if present {
			result = append(result, parameter)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Key < result[j].Key })
	return result
}

func sameImageCapability(left, right ImageParameterCapability) bool {
	return left.RequestKey == right.RequestKey &&
		left.Type == right.Type &&
		reflect.DeepEqual(left.Default, right.Default) &&
		reflect.DeepEqual(left.Min, right.Min) &&
		reflect.DeepEqual(left.Max, right.Max) &&
		reflect.DeepEqual(left.Step, right.Step) &&
		reflect.DeepEqual(left.Options, right.Options)
}

func containsString(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

// GetImageModelCapabilities returns the conservative parameter intersection
// for the user's selected group/model. It never exposes another user's groups
// or a channel that is not enabled for image generation.
func GetImageModelCapabilities(c *gin.Context) {
	user, err := model.GetUserCache(c.GetInt("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	modelName := strings.TrimSpace(c.Query("model"))
	if modelName == "" {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}
	operation := strings.TrimSpace(c.Query("operation"))
	if operation != "edit" {
		operation = "generation"
	}
	groups := service.GetUserUsableGroupsWithExtras(user.Group, user.GetExtraGroups())
	group := strings.TrimSpace(c.Query("group"))
	if group == "auto" {
		groupNames := service.GetUserAutoGroupWithExtras(user.Group, user.GetExtraGroups())
		groupsForLookup := make(map[string]string, len(groupNames))
		for _, name := range groupNames {
			if ratio, ok := groups[name]; ok {
				groupsForLookup[name] = ratio
			}
		}
		groups = groupsForLookup
	} else if group != "" {
		if _, ok := groups[group]; !ok {
			common.ApiErrorI18n(c, i18n.MsgInvalidParams)
			return
		}
		groups = map[string]string{group: groups[group]}
	}

	model.GetPricing()
	abilities, err := model.GetAllEnableAbilityWithChannels()
	if err != nil {
		common.ApiError(c, err)
		return
	}
	channelTypes := make([]int, 0)
	seenChannels := make(map[int]struct{})
	for _, ability := range abilities {
		if ability.Model != modelName {
			continue
		}
		if _, ok := groups[ability.Group]; !ok {
			continue
		}
		if !abilitySupportsEndpointType(ability.ChannelType, ability.Model, constant.EndpointTypeImageGeneration) {
			continue
		}
		if _, ok := seenChannels[ability.ChannelType]; ok {
			continue
		}
		seenChannels[ability.ChannelType] = struct{}{}
		channelTypes = append(channelTypes, ability.ChannelType)
	}
	if len(channelTypes) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "image model is not available"})
		return
	}

	parameters := intersectImageCapabilities(imageCapabilityProfiles(), channelTypes, modelName, operation)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"model":      modelName,
			"operation":  operation,
			"parameters": parameters,
		},
	})
}
