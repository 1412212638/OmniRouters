package common

import "strings"

// ZImagePromptExtendMultiplier is the legacy Ali image request surcharge.
const ZImagePromptExtendMultiplier = 2

var (
	// OpenAIResponseOnlyModels is a list of models that are only available for OpenAI responses.
	OpenAIResponseOnlyModels = []string{
		"o3-pro",
		"o3-deep-research",
		"o4-mini-deep-research",
	}
	ImageGenerationModels = []string{
		"dall-e-3",
		"dall-e-2",
		"prefix:dall-e", // Deprecated upstream models; retained for compatible routes.
		"gpt-image-",
		"qwen-image",
		"z-image",
		"wan2.7-image-pro",
		"wan2.7-image",
		"wan2.6-image",
		"wan2.6-t2i",
		"wan2.5-t2i-preview",
		"wan2.2-t2i-flash",
		"wan2.2-t2i-plus",
		"wanx2.1-t2i-turbo",
		"wanx2.1-t2i-plus",
		"wanx2.0-t2i-turbo",
		"gpt-image-1",
		"grok-2-image",
		"grok-image",
		"prefix:imagen-",
		"flux-",
		"flux.1-",
		"stable-diffusion",
		"stable-image",
		"sdxl",
		"kolors",
		"wanx",
		"seedream",
		"seededit",
		"midjourney",
		"ideogram",
		"recraft",
		"nano-banana",
		"t2i",
		"text-to-image",
		"image-generation",
	}
	EmbeddingModels = []string{
		"embedding",
		"embeddings",
		"text-embedding",
		"gemini-embedding",
		"embed-",
		"bge-",
		"bce-embedding",
		"gte-",
		"e5-",
		"text2vec",
		"m3e",
	}
	RerankModels = []string{
		"rerank",
		"reranker",
	}
	VideoGenerationModels = []string{
		"hailuo",
		"kling",
		"jimeng",
		"vidu",
		"sora",
		"veo",
		"seedance",
		"pika",
		"runway",
		"hunyuanvideo",
		"hunyuan-video",
		"hunyuan-1.5",
		"hunyuan-3.0",
		"mingmou",
		"t2v-",
		"i2v-",
		"s2v-",
		"text-to-video",
		"image-to-video",
		"video-generation",
	}
	OpenAITextModels = []string{
		"gpt-",
		"o1",
		"o3",
		"o4",
		"chatgpt",
	}
)

func IsOpenAIResponseOnlyModel(modelName string) bool {
	for _, m := range OpenAIResponseOnlyModels {
		if strings.Contains(modelName, m) {
			return true
		}
	}
	return false
}

func IsImageGenerationModel(modelName string) bool {
	modelName = strings.ToLower(modelName)
	for _, m := range ImageGenerationModels {
		if prefix, ok := strings.CutPrefix(m, "prefix:"); ok {
			if strings.HasPrefix(modelName, prefix) {
				return true
			}
			continue
		}
		if strings.Contains(modelName, m) {
			return true
		}
	}
	return false
}

func IsEmbeddingModel(modelName string) bool {
	modelName = strings.ToLower(modelName)
	for _, m := range EmbeddingModels {
		if strings.Contains(modelName, m) {
			return true
		}
	}
	return false
}

func IsRerankModel(modelName string) bool {
	modelName = strings.ToLower(modelName)
	for _, m := range RerankModels {
		if strings.Contains(modelName, m) {
			return true
		}
	}
	return false
}

func IsVideoGenerationModel(modelName string) bool {
	modelName = strings.ToLower(modelName)
	for _, m := range VideoGenerationModels {
		if strings.Contains(modelName, m) {
			return true
		}
	}
	return false
}

func IsOpenAITextModel(modelName string) bool {
	modelName = strings.ToLower(modelName)
	for _, m := range OpenAITextModels {
		if strings.Contains(modelName, m) {
			return true
		}
	}
	return false
}
