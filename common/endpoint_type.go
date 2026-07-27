package common

import "github.com/QuantumNous/new-api/constant"

// GetEndpointTypesByChannelType returns endpoint types in preferred order for a channel/model pair.
func GetEndpointTypesByChannelType(channelType int, modelName string) []constant.EndpointType {
	var endpointTypes []constant.EndpointType
	switch channelType {
	case constant.ChannelTypeJina:
		if IsEmbeddingModel(modelName) {
			endpointTypes = []constant.EndpointType{constant.EndpointTypeEmbeddings}
		} else {
			endpointTypes = []constant.EndpointType{constant.EndpointTypeJinaRerank}
		}
	//case constant.ChannelTypeMidjourney, constant.ChannelTypeMidjourneyPlus:
	//	endpointTypes = []constant.EndpointType{constant.EndpointTypeMidjourney}
	//case constant.ChannelTypeSunoAPI:
	//	endpointTypes = []constant.EndpointType{constant.EndpointTypeSuno}
	//case constant.ChannelTypeKling:
	//	endpointTypes = []constant.EndpointType{constant.EndpointTypeKling}
	//case constant.ChannelTypeJimeng:
	//	endpointTypes = []constant.EndpointType{constant.EndpointTypeJimeng}
	case constant.ChannelTypeAws:
		fallthrough
	case constant.ChannelTypeAnthropic:
		endpointTypes = []constant.EndpointType{constant.EndpointTypeAnthropic, constant.EndpointTypeOpenAI}
	case constant.ChannelTypeVertexAi:
		fallthrough
	case constant.ChannelTypeGemini:
		endpointTypes = []constant.EndpointType{constant.EndpointTypeGemini, constant.EndpointTypeOpenAI}
	case constant.ChannelTypeOpenRouter:
		endpointTypes = []constant.EndpointType{constant.EndpointTypeOpenAI}
	case constant.ChannelTypeXai:
		endpointTypes = []constant.EndpointType{constant.EndpointTypeOpenAI, constant.EndpointTypeOpenAIResponse}
	case constant.ChannelTypeSora:
		endpointTypes = []constant.EndpointType{constant.EndpointTypeOpenAIVideo}
	case constant.ChannelTypeSub2API, constant.ChannelTypeNewAPI:
		endpointTypes = []constant.EndpointType{
			constant.EndpointTypeOpenAI,
			constant.EndpointTypeOpenAIResponse,
			constant.EndpointTypeOpenAIResponseCompact,
			constant.EndpointTypeAnthropic,
			constant.EndpointTypeGemini,
			constant.EndpointTypeOpenAIAlphaSearch,
		}
	case constant.ChannelTypeCodex:
		endpointTypes = []constant.EndpointType{
			constant.EndpointTypeOpenAIResponse,
			constant.EndpointTypeOpenAIResponseCompact,
			constant.EndpointTypeOpenAIAlphaSearch,
		}
	default:
		if IsOpenAIResponseOnlyModel(modelName) {
			endpointTypes = []constant.EndpointType{constant.EndpointTypeOpenAIResponse}
		} else {
			endpointTypes = []constant.EndpointType{constant.EndpointTypeOpenAI}
		}
	}
	if IsRerankModel(modelName) {
		endpointTypes = prependEndpointType(endpointTypes, constant.EndpointTypeJinaRerank)
	}
	if IsEmbeddingModel(modelName) {
		endpointTypes = prependEndpointType(endpointTypes, constant.EndpointTypeEmbeddings)
	}
	if IsVideoGenerationModel(modelName) {
		endpointTypes = prependEndpointType(endpointTypes, constant.EndpointTypeOpenAIVideo)
	}
	if IsImageGenerationModel(modelName) {
		endpointTypes = prependEndpointType(endpointTypes, constant.EndpointTypeImageGeneration)
	}
	return endpointTypes
}

func IsOpenAIChatEndpointModel(channelType int, modelName string) bool {
	if isNonChatOnlyChannel(channelType) {
		return false
	}
	endpointTypes := GetEndpointTypesByChannelType(channelType, modelName)
	return IsOpenAIChatEndpointTypes(endpointTypes)
}

func IsOpenAIChatEndpointTypes(endpointTypes []constant.EndpointType) bool {
	hasOpenAI := false
	for _, endpointType := range endpointTypes {
		if endpointType == constant.EndpointTypeOpenAI {
			hasOpenAI = true
			continue
		}
		if isNonChatEndpointType(endpointType) {
			return false
		}
	}
	return hasOpenAI
}

func ContainsEndpointType(endpointTypes []constant.EndpointType, endpointType constant.EndpointType) bool {
	for _, current := range endpointTypes {
		if current == endpointType {
			return true
		}
	}
	return false
}

func prependEndpointType(endpointTypes []constant.EndpointType, endpointType constant.EndpointType) []constant.EndpointType {
	for _, current := range endpointTypes {
		if current == endpointType {
			return endpointTypes
		}
	}
	return append([]constant.EndpointType{endpointType}, endpointTypes...)
}

func isNonChatEndpointType(endpointType constant.EndpointType) bool {
	switch endpointType {
	case constant.EndpointTypeImageGeneration,
		constant.EndpointTypeEmbeddings,
		constant.EndpointTypeJinaRerank,
		constant.EndpointTypeOpenAIVideo:
		return true
	default:
		return false
	}
}

func isNonChatOnlyChannel(channelType int) bool {
	switch channelType {
	case constant.ChannelTypeMidjourney,
		constant.ChannelTypeMidjourneyPlus,
		constant.ChannelTypeSunoAPI,
		constant.ChannelTypeJina,
		constant.ChannelTypeMokaAI,
		constant.ChannelTypeKling,
		constant.ChannelTypeJimeng,
		constant.ChannelTypeVidu,
		constant.ChannelTypeDoubaoVideo,
		constant.ChannelTypeSora,
		constant.ChannelTypeReplicate:
		return true
	default:
		return false
	}
}
