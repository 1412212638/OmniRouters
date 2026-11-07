package helper

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/pkg/billingexpr"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	relayconstant "github.com/QuantumNous/new-api/relay/constant"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/stretchr/testify/require"
)

func TestSeedreamModelNameLimitsCompatibilityModels(t *testing.T) {
	require.Equal(t, "doubao-seedream-5.0-pro", seedreamModelName("doubao-seedream-5-0-pro-260628"))
	require.Equal(t, "doubao-seedream-5.0-flash", seedreamModelName("DOUBAO-SEEDREAM-5.0-FLASH"))
	require.Empty(t, seedreamModelName("doubao-seedream-4-0-250828"))
	require.Empty(t, seedreamModelName("minimax-h3"))
}

func TestSeedreamImageEstimateUsesOnlySafeRequestFields(t *testing.T) {
	request := &dto.ImageRequest{Model: "doubao-seedream-5.0-pro", Prompt: "private prompt", Size: "1.5K"}
	facts, count, safeBody, err := seedreamImageEstimate(request.Model, request, []byte(`{"image":["https://example.test/1.png","https://example.test/2.png"]}`), 1)
	require.NoError(t, err)
	require.Equal(t, 1, count)
	require.Equal(t, float64(2), facts["input_images"])
	require.Equal(t, float64(1), facts["images_up_to_1_5k"])
	require.Equal(t, float64(0), facts["images_above_1_5k"])
	require.NotContains(t, string(safeBody), "example.test")
	require.NotContains(t, string(safeBody), "private prompt")
}

func TestSeedreamImageEstimateReservesHighTierForUnknownSize(t *testing.T) {
	request := &dto.ImageRequest{Model: "doubao-seedream-5.0-pro"}
	facts, count, _, err := seedreamImageEstimate(request.Model, request, nil, 1)
	require.NoError(t, err)
	require.Equal(t, 1, count)
	require.Equal(t, float64(0), facts["images_up_to_1_5k"])
	require.Equal(t, float64(1), facts["images_above_1_5k"])
}

func TestUpdateSeedreamImageUsageHonorsActualZeroAndOutputPixels(t *testing.T) {
	count := 1
	expr := `u("images_up_to_1_5k") * 0.1 + u("images_above_1_5k") * 0.2 + max(u("input_images") - 1, 0) * 0.02`
	info := &relaycommon.RelayInfo{
		SeedreamImageBilling: true,
		BillingRequestInput: &billingexpr.RequestInput{Usage: map[string]any{
			"images_up_to_1_5k":  float64(0),
			"images_above_1_5k":  float64(1),
			"input_images":       float64(1),
			"layer_decomposition": false,
		}},
		TieredBillingSnapshot: seedreamTestSnapshot(expr, &count),
	}
	response := []byte(`{"data":[{"url":"https://example.test/out.png","size":"1664x2496"}],"usage":{"input_images":0,"generated_images":1,"output_tokens":16224}}`)

	UpdateSeedreamOpenAIImageUsage(info, response, 0)

	require.Equal(t, float64(0), info.BillingRequestInput.Usage["input_images"])
	require.Equal(t, float64(0), info.BillingRequestInput.Usage["images_up_to_1_5k"])
	require.Equal(t, float64(1), info.BillingRequestInput.Usage["images_above_1_5k"])
	require.Equal(t, float64(0), info.TieredBillingSnapshot.UsageFacts["input_images"])
	require.NotNil(t, info.BillingImageCount)
	require.Equal(t, 1, *info.BillingImageCount)
}

func TestUpdateSeedreamImageUsageRetainsEstimateOnInvalidOutputCount(t *testing.T) {
	count := 2
	expr := `u("images_up_to_1_5k") * 0.1 + u("images_above_1_5k") * 0.2 + max(u("input_images") - 1, 0) * 0.02`
	info := &relaycommon.RelayInfo{
		SeedreamImageBilling: true,
		BillingRequestInput: &billingexpr.RequestInput{Usage: map[string]any{
			"images_up_to_1_5k":  float64(2),
			"images_above_1_5k":  float64(0),
			"input_images":       float64(0),
			"layer_decomposition": false,
		}},
		TieredBillingSnapshot: seedreamTestSnapshot(expr, &count),
	}

	UpdateSeedreamOpenAIImageUsage(info, []byte(`{"usage":{"input_images":0,"generated_images":9999999999999999999}}`), 0)

	require.Equal(t, float64(2), info.BillingRequestInput.Usage["images_up_to_1_5k"])
	require.Equal(t, float64(0), info.BillingRequestInput.Usage["images_above_1_5k"])
	require.Equal(t, float64(0), info.BillingRequestInput.Usage["input_images"])
	require.Nil(t, info.BillingImageCount)
}

func seedreamTestSnapshot(expression string, imageCount *int) *billingexpr.BillingSnapshot {
	return &billingexpr.BillingSnapshot{
		ExprString:          expression,
		ExprHash:            billingexpr.ExprHashString(expression),
		BillingMode:        "tiered_expr",
		EstimatedImageCount: imageCount,
	}
}

func TestSeedreamImageEstimateRejectsUnboundedQuantities(t *testing.T) {
	request := &dto.ImageRequest{Model: "doubao-seedream-5.0-pro"}
	_, _, _, err := seedreamImageEstimate(request.Model, request, []byte(`{"layer_decomposition":true}`), 1)
	require.NoError(t, err)
	_, _, _, err = seedreamImageEstimate(request.Model, request, nil, seedreamMaxOutputs+1)
	require.ErrorContains(t, err, "at most 17")
}

func TestSeedreamOrdinaryUsageExpressionIsScopedToOpenAIRequests(t *testing.T) {
	request := &dto.ImageRequest{Model: "doubao-seedream-5.0-pro", Size: "2K"}
	info := &relaycommon.RelayInfo{
		OriginModelName: "doubao-seedream-5.0-pro",
		RelayMode:       relayconstant.RelayModeImagesGenerations,
		ChannelMeta:     &relaycommon.ChannelMeta{ChannelType: constant.ChannelTypeOpenAI},
		TieredBillingSnapshot: seedreamTestSnapshot(
			`u("images_up_to_1_5k") * 0.1 + u("images_above_1_5k") * 0.2 + max(u("input_images") - 1, 0) * 0.02`,
			nil,
		),
	}
	info.PriceData.GroupRatioInfo.GroupRatio = 1
	require.NoError(t, PrepareSeedreamOpenAIImageBilling(info, request, nil, 1))
	require.True(t, info.SeedreamImageBilling)
	require.True(t, info.TieredBillingSnapshot.ResponseUsageBilling)
	require.Equal(t, "doubao-seedream-5.0-pro", info.TieredBillingSnapshot.ModelName)
	require.Equal(t, float64(1), info.BillingRequestInput.Usage["images_above_1_5k"])
	require.Equal(t, common.QuotaRound(0.2*common.QuotaPerUnit), info.PriceData.QuotaToPreConsume)

	baseSnapshot := &billingexpr.BillingSnapshot{ModelName: "base-expression"}
	otherChannel := &relaycommon.RelayInfo{
		OriginModelName:        "doubao-seedream-5.0-pro",
		RelayMode:              relayconstant.RelayModeImagesGenerations,
		ChannelMeta:            &relaycommon.ChannelMeta{ChannelType: constant.ChannelTypeVolcEngine},
		TieredBillingSnapshot: baseSnapshot,
	}
	require.NoError(t, PrepareSeedreamOpenAIImageBilling(otherChannel, request, nil, 1))
	require.False(t, otherChannel.SeedreamImageBilling)
	require.Same(t, baseSnapshot, otherChannel.TieredBillingSnapshot)

	otherModel := &relaycommon.RelayInfo{
		OriginModelName:        "gpt-image-2",
		RelayMode:              relayconstant.RelayModeImagesGenerations,
		ChannelMeta:            &relaycommon.ChannelMeta{ChannelType: constant.ChannelTypeOpenAI},
		TieredBillingSnapshot: baseSnapshot,
	}
	require.NoError(t, PrepareSeedreamOpenAIImageBilling(otherModel, request, nil, 1))
	require.False(t, otherModel.SeedreamImageBilling)
	require.Same(t, baseSnapshot, otherModel.TieredBillingSnapshot)
}
