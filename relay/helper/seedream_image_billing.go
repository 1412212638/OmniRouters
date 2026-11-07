package helper

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/pkg/billingexpr"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	relayconstant "github.com/QuantumNous/new-api/relay/constant"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/setting/billing_setting"
	"github.com/tidwall/gjson"
)

const (
	seedreamMaxOutputs      = 17
	seedreamMaxReferences   = 14
	seedreamMaxOutputPixels = 2610000
)

// PrepareSeedreamOpenAIImageBilling activates the ordinary model expression's
// response-usage profile for the named Seedream 5.0 models on an OpenAI image
// channel. It never reads or mutates a task plugin expression. Other models,
// channel types, and ordinary token expressions keep the standard image path.
func PrepareSeedreamOpenAIImageBilling(info *relaycommon.RelayInfo, request *dto.ImageRequest, outboundBody []byte, requestedCount int) error {
	if info == nil {
		return nil
	}
	if request == nil || info.ChannelMeta == nil || info.ChannelMeta.ChannelType != constant.ChannelTypeOpenAI ||
		(info.RelayMode != relayconstant.RelayModeImagesGenerations && info.RelayMode != relayconstant.RelayModeImagesEdits) {
		return nil
	}

	ordinarySnapshot := info.TieredBillingSnapshot
	if ordinarySnapshot == nil || ordinarySnapshot.BillingMode != "tiered_expr" || len(billingexpr.UsedUsageKeys(ordinarySnapshot.ExprString)) == 0 {
		return nil
	}
	modelName := seedreamModelName(info.OriginModelName)
	if modelName == "" {
		return nil
	}
	if err := billing_setting.ValidateOrdinaryResponseUsageExpr(modelName, ordinarySnapshot.ExprString); err != nil {
		return err
	}

	facts, imageCount, safeBody, err := seedreamImageEstimate(modelName, request, outboundBody, requestedCount)
	if err != nil {
		return err
	}
	input := billingexpr.RequestInput{Headers: cloneStringMap(info.RequestHeaders), Body: safeBody, Usage: facts}
	if info.BillingRequestInput != nil {
		input.Headers = cloneStringMap(info.BillingRequestInput.Headers)
	}
	input.ImageCount = &imageCount

	// Clone the ordinary snapshot before attaching usage facts. A later channel
	// retry can restore the original ordinary billing state without mutating any
	// plugin snapshot or frozen expression.
	usageSnapshot := cloneBillingSnapshot(ordinarySnapshot)
	hash := usageSnapshot.ExprHash
	if hash == "" {
		hash = billingexpr.ExprHashString(usageSnapshot.ExprString)
		usageSnapshot.ExprHash = hash
	}
	cost, trace, err := billingexpr.RunExprByHashWithRequest(usageSnapshot.ExprString, hash, billingexpr.TokenParams{}, input)
	if err != nil || math.IsNaN(cost) || math.IsInf(cost, 0) || cost < 0 {
		if err == nil {
			err = fmt.Errorf("expression produced an invalid cost")
		}
		return fmt.Errorf("Seedream ordinary image pricing failed: %w", err)
	}

	groupRatio := info.PriceData.GroupRatioInfo.GroupRatio
	quotaBeforeGroup := cost * common.QuotaPerUnit
	quota, err := billingexpr.QuotaRoundStrict(quotaBeforeGroup * groupRatio)
	if err != nil {
		return fmt.Errorf("Seedream image pre-consume quota is invalid: %w", err)
	}
	count := imageCount
	usageSnapshot.EstimatedImageCount = &count
	usageSnapshot.ModelName = modelName
	usageSnapshot.GroupRatio = groupRatio
	usageSnapshot.EstimatedQuotaBeforeGroup = quotaBeforeGroup
	usageSnapshot.EstimatedQuotaAfterGroup = quota
	usageSnapshot.EstimatedTier = trace.MatchedTier
	usageSnapshot.EstimatedRequestRules = trace.RequestRules
	usageSnapshot.EstimatedBillingUnit = trace.BillingUnit
	usageSnapshot.EstimatedFixedPrice = trace.FixedPrice
	usageSnapshot.QuotaPerUnit = common.QuotaPerUnit
	usageSnapshot.ExprVersion = billingexpr.ExprVersion(usageSnapshot.ExprString)
	usageSnapshot.ResponseUsageBilling = true
	usageSnapshot.UsageFacts = cloneAnyMap(facts)
	info.BeginSeedreamImageBilling(usageSnapshot, &input)
	info.PriceData.QuotaToPreConsume = quota
	info.PriceData.FreeModel = false
	return nil
}

func cloneBillingSnapshot(source *billingexpr.BillingSnapshot) *billingexpr.BillingSnapshot {
	if source == nil {
		return nil
	}
	clone := *source
	if source.EstimatedImageCount != nil {
		value := *source.EstimatedImageCount
		clone.EstimatedImageCount = &value
	}
	clone.EstimatedRequestRules = append([]billingexpr.RequestRuleTrace(nil), source.EstimatedRequestRules...)
	clone.UsageFacts = cloneAnyMap(source.UsageFacts)
	return &clone
}

// UpdateSeedreamOpenAIImageUsage replaces only validated Seedream usage facts.
// Missing or malformed output usage leaves its conservative reservation facts
// in place, so an incomplete upstream response cannot cause an accidental zero
// charge. completedCount is supplied by the SSE handler when the upstream
// streams image completion events rather than returning data[].
func UpdateSeedreamOpenAIImageUsage(info *relaycommon.RelayInfo, responseBody []byte, completedCount int64) {
	if info == nil || !info.SeedreamImageBilling || info.BillingRequestInput == nil || info.BillingRequestInput.Usage == nil {
		return
	}

	facts := cloneAnyMap(info.BillingRequestInput.Usage)
	usage := gjson.GetBytes(responseBody, "usage")
	if inputImages, ok := boundedJSONInteger(usage.Get("input_images"), seedreamMaxReferences); ok {
		facts["input_images"] = float64(inputImages)
	}

	lower, higher, count, ok := seedreamResponseOutputFacts(responseBody, completedCount, facts)
	updateImageCount := false
	if ok {
		facts["images_up_to_1_5k"] = float64(lower)
		facts["images_above_1_5k"] = float64(higher)
		facts["generated_images"] = float64(count)
		updateImageCount = true
	}

	snapshot := info.TieredBillingSnapshot
	if snapshot == nil {
		return
	}
	actualInput := *info.BillingRequestInput
	actualInput.Usage = facts
	if updateImageCount {
		imageCount := int(count)
		actualInput.ImageCount = &imageCount
	}
	cost, _, err := billingexpr.RunExprByHashWithRequest(snapshot.ExprString, snapshot.ExprHash, billingexpr.TokenParams{}, actualInput)
	if err != nil || math.IsNaN(cost) || math.IsInf(cost, 0) || cost < 0 {
		return
	}
	if updateImageCount {
		info.UpdateImageCount(count)
	}
	info.BillingRequestInput.Usage = facts
	snapshot.UsageFacts = cloneAnyMap(facts)
}

func seedreamModelName(name string) string {
	return billing_setting.CanonicalOrdinaryResponseUsageModel(name)
}

func seedreamImageEstimate(modelName string, request *dto.ImageRequest, outboundBody []byte, requestedCount int) (map[string]any, int, []byte, error) {
	if requestedCount < 1 || requestedCount > seedreamMaxOutputs {
		return nil, 0, nil, fmt.Errorf("Seedream supports at most %d generated images per request", seedreamMaxOutputs)
	}

	inputImages := seedreamReferenceImageCount(request, outboundBody)
	if inputImages > seedreamMaxReferences {
		return nil, 0, nil, fmt.Errorf("Seedream supports at most %d reference images", seedreamMaxReferences)
	}

	layered := requestValue(outboundBody, request, "layer_decomposition").Type == gjson.True
	sequential := requestValue(outboundBody, request, "sequential_image_generation").String() == "auto"
	outputCount := requestedCount
	if layered {
		outputCount = seedreamMaxOutputs
	} else if sequential {
		maxImages, ok := boundedJSONInteger(gjson.GetBytes(outboundBody, "sequential_image_generation_options.max_images"), 15)
		if !ok || maxImages == 0 {
			maxImages = 15
		}
		outputCount = int(maxImages) - inputImages
		if outputCount < 1 {
			outputCount = 1
		}
	}
	if outputCount > seedreamMaxOutputs {
		return nil, 0, nil, fmt.Errorf("Seedream supports at most %d generated images per request", seedreamMaxOutputs)
	}

	size := requestValue(outboundBody, request, "size").String()
	lower, higher := 0, outputCount
	if seedreamIsLowerTierSize(size) {
		lower, higher = outputCount, 0
	}
	facts := map[string]any{
		"images_up_to_1_5k":   float64(lower),
		"images_above_1_5k":   float64(higher),
		"generated_images":    float64(outputCount),
		"input_images":        float64(inputImages),
		"layer_decomposition": layered,
	}

	// Keep expression request probes useful without ever retaining prompt or
	// image payloads in billing state.
	safeRequest := map[string]any{"model": modelName, "n": requestedCount}
	if size != "" {
		safeRequest["size"] = size
	}
	if quality := requestValue(outboundBody, request, "quality").String(); quality != "" {
		safeRequest["quality"] = quality
	}
	safeRequest["layer_decomposition"] = layered
	if sequential {
		safeRequest["sequential_image_generation"] = "auto"
		if maxImages, ok := boundedJSONInteger(gjson.GetBytes(outboundBody, "sequential_image_generation_options.max_images"), 15); ok {
			safeRequest["sequential_image_generation_options"] = map[string]any{"max_images": maxImages}
		}
	}
	encoded, err := common.Marshal(safeRequest)
	if err != nil {
		return nil, 0, nil, err
	}
	return facts, outputCount, encoded, nil
}

func seedreamReferenceImageCount(request *dto.ImageRequest, outboundBody []byte) int {
	for _, key := range []string{"image", "images", "image_url"} {
		value := requestValue(outboundBody, request, key)
		if !value.Exists() || value.Type == gjson.Null {
			continue
		}
		switch value.Type {
		case gjson.Array:
			return len(value.Array())
		case gjson.String:
			if strings.TrimSpace(value.String()) != "" {
				return 1
			}
		case gjson.JSON:
			return 1
		}
	}
	return 0
}

func requestValue(body []byte, request *dto.ImageRequest, key string) gjson.Result {
	value := gjson.GetBytes(body, key)
	if value.Exists() {
		return value
	}
	if request == nil {
		return gjson.Result{}
	}
	switch key {
	case "n":
		if request.N != nil {
			return gjson.Result{Type: gjson.Number, Raw: strconv.FormatUint(uint64(*request.N), 10)}
		}
	case "size":
		if request.Size != "" {
			return gjson.Result{Type: gjson.String, Str: request.Size}
		}
	case "quality":
		if request.Quality != "" {
			return gjson.Result{Type: gjson.String, Str: request.Quality}
		}
	case "image":
		if len(request.Image) > 0 {
			return gjson.ParseBytes(request.Image)
		}
	case "images":
		if len(request.Images) > 0 {
			return gjson.ParseBytes(request.Images)
		}
	default:
		if raw, ok := request.Extra[key]; ok {
			return gjson.ParseBytes(raw)
		}
		if parent, child, ok := strings.Cut(key, "."); ok {
			if raw, exists := request.Extra[parent]; exists {
				return gjson.GetBytes(raw, child)
			}
		}
	}
	return gjson.Result{}
}

func seedreamResponseOutputFacts(body []byte, completedCount int64, estimated map[string]any) (int, int, int64, bool) {
	usage := gjson.GetBytes(body, "usage")
	data := gjson.GetBytes(body, "data")
	payloads := make([]gjson.Result, 0)
	if data.IsArray() {
		data.ForEach(func(_, item gjson.Result) bool {
			if hasImagePayload(item) {
				payloads = append(payloads, item)
			}
			return true
		})
	} else if data.IsObject() && hasImagePayload(data) {
		payloads = append(payloads, data)
	}
	if len(payloads) > seedreamMaxOutputs {
		return 0, 0, 0, false
	}

	count := int64(0)
	if usage.Get("generated_images").Exists() {
		var valid bool
		count, valid = boundedJSONInteger(usage.Get("generated_images"), seedreamMaxOutputs)
		if !valid || count < int64(len(payloads)) {
			return 0, 0, 0, false
		}
	} else if completedCount > 0 {
		count = completedCount
	} else {
		count = int64(len(payloads))
	}
	if count < 1 || count > seedreamMaxOutputs {
		return 0, 0, 0, false
	}
	if len(payloads) == 0 {
		if estimatedLowerTier(estimated) {
			return int(count), 0, count, true
		}
		return 0, int(count), count, true
	}

	low, high := 0, 0
	for _, item := range payloads {
		if seedreamIsLowerTierSize(item.Get("size").String()) || (!item.Get("size").Exists() && estimatedLowerTier(estimated)) {
			low++
		} else {
			high++
		}
	}
	missing := int(count) - len(payloads)
	if missing > 0 {
		high += missing
	}
	return low, high, count, true
}

func hasImagePayload(item gjson.Result) bool {
	for _, key := range []string{"url", "b64_json"} {
		value := item.Get(key)
		if value.Type == gjson.String && strings.TrimSpace(value.String()) != "" {
			return true
		}
	}
	return false
}

func estimatedLowerTier(facts map[string]any) bool {
	lower, _ := facts["images_up_to_1_5k"].(float64)
	higher, _ := facts["images_above_1_5k"].(float64)
	return lower > higher
}

func seedreamIsLowerTierSize(size string) bool {
	size = strings.ToLower(strings.TrimSpace(size))
	if size == "1k" || size == "1.5k" {
		return true
	}
	parts := strings.Split(size, "x")
	if len(parts) != 2 {
		return false
	}
	width, widthErr := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)
	height, heightErr := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
	return widthErr == nil && heightErr == nil && width > 0 && height > 0 && width <= seedreamMaxOutputPixels/height && width*height <= seedreamMaxOutputPixels
}

func boundedJSONInteger(value gjson.Result, maximum int) (int64, bool) {
	if !value.Exists() || value.Type != gjson.Number {
		return 0, false
	}
	number := value.Float()
	if math.IsNaN(number) || math.IsInf(number, 0) || number < 0 || math.Trunc(number) != number || number > float64(maximum) {
		return 0, false
	}
	return int64(number), true
}

func cloneAnyMap(source map[string]any) map[string]any {
	if len(source) == 0 {
		return nil
	}
	copy := make(map[string]any, len(source))
	for key, value := range source {
		copy[key] = value
	}
	return copy
}
