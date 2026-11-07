package common

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/pkg/billingexpr"
	"github.com/QuantumNous/new-api/relay/constant"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/relaykit/relayconvert/convmeta"
	"github.com/QuantumNous/new-api/relaykit/types"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenRelayInfoAlphaSearch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/alpha/search", nil)

	info, err := GenRelayInfo(c, types.RelayFormatOpenAIAlphaSearch, &dto.AlphaSearchRequest{Model: "gpt-5.1"}, nil)
	require.NoError(t, err)
	assert.Equal(t, types.RelayFormatOpenAIAlphaSearch, info.RelayFormat)
	assert.Equal(t, constant.RelayModeAlphaSearch, info.RelayMode)
	assert.Equal(t, []types.RelayFormat{types.RelayFormatOpenAIAlphaSearch}, info.RequestConversionChain)
}

func TestRelayInfoGetFinalRequestRelayFormatPrefersExplicitFinal(t *testing.T) {
	info := &RelayInfo{RelayFormat: types.RelayFormatOpenAI, RequestConversionChain: []types.RelayFormat{types.RelayFormatOpenAI, types.RelayFormatClaude}, FinalRequestRelayFormat: types.RelayFormatOpenAIResponses}
	require.Equal(t, types.RelayFormat(types.RelayFormatOpenAIResponses), info.GetFinalRequestRelayFormat())
}

func TestRelayInfoGetFinalRequestRelayFormatFallsBackToConversionChain(t *testing.T) {
	info := &RelayInfo{RelayFormat: types.RelayFormatOpenAI, RequestConversionChain: []types.RelayFormat{types.RelayFormatOpenAI, types.RelayFormatClaude}}
	require.Equal(t, types.RelayFormat(types.RelayFormatClaude), info.GetFinalRequestRelayFormat())
}

func TestRelayInfoGetFinalRequestRelayFormatFallsBackToRelayFormat(t *testing.T) {
	info := &RelayInfo{RelayFormat: types.RelayFormatGemini}
	require.Equal(t, types.RelayFormat(types.RelayFormatGemini), info.GetFinalRequestRelayFormat())
}

func TestRelayInfoGetFinalRequestRelayFormatNilReceiver(t *testing.T) {
	var info *RelayInfo
	require.Equal(t, types.RelayFormat(""), info.GetFinalRequestRelayFormat())
}

func TestRelayInfoMetaTypedNilReceiver(t *testing.T) {
	var info *RelayInfo
	var meta convmeta.Meta = info
	assert.Empty(t, meta.GetOriginModelName())
	assert.Empty(t, meta.GetUpstreamModelName())
	assert.False(t, meta.HasChannelMeta())
	assert.Zero(t, meta.GetChannelID())
	assert.Zero(t, meta.GetChannelType())
	assert.False(t, meta.GetIsStream())
	assert.Empty(t, meta.GetReasoningEffort())
	assert.Zero(t, meta.GetSendResponseCount())
	assert.NotPanics(t, func() {
		meta.SetReasoningEffort("high")
		meta.IncrSendResponseCount()
		meta.AppendRequestConversion(types.RelayFormatClaude)
		meta.SetResponsesToolState(&convmeta.ResponsesToolState{})
	})
	assert.Nil(t, meta.ResponsesToolState())
	firstState := meta.EnsureClaudeConvertInfo()
	secondState := meta.EnsureClaudeConvertInfo()
	require.NotNil(t, firstState)
	require.NotNil(t, secondState)
	assert.Equal(t, convmeta.LastMessageTypeNone, firstState.LastMessagesType)
	assert.NotSame(t, firstState, secondState)
	firstOptions := meta.ConvOptions()
	secondOptions := meta.ConvOptions()
	require.NotNil(t, firstOptions)
	require.NotNil(t, secondOptions)
	assert.NotSame(t, firstOptions, secondOptions)
	assert.NotNil(t, firstOptions.Claude.DefaultMaxTokens)
	assert.NotNil(t, firstOptions.Gemini.SupportsImagine)
	assert.NotNil(t, firstOptions.Gemini.SafetySetting)
	assert.NotNil(t, firstOptions.PreserveThinkingSuffix)
}

func TestRelayInfoResponsesToolStateRoundTrip(t *testing.T) {
	info := &RelayInfo{}
	state := &convmeta.ResponsesToolState{CustomToolNames: map[string]struct{}{"exec": {}}}
	info.SetResponsesToolState(state)
	assert.Same(t, state, info.ResponsesToolState())
	assert.True(t, convmeta.ResponsesToolStateOf(info).IsCustomTool("exec"))
	info.SetResponsesToolState(nil)
	assert.Nil(t, info.ResponsesToolState())
	assert.False(t, convmeta.ResponsesToolStateOf(info).IsCustomTool("exec"))
}

func TestSeedreamImageBillingRestoresBaseStateForChannelRetry(t *testing.T) {
	baseSnapshot := &billingexpr.BillingSnapshot{ModelName: "base-model"}
	baseInput := &billingexpr.RequestInput{Body: []byte(`{"size":"2K"}`)}
	baseImageCount := 3
	info := &RelayInfo{
		TieredBillingSnapshot: baseSnapshot,
		BillingRequestInput:   baseInput,
		BillingImageCount:     &baseImageCount,
	}
	info.PriceData.QuotaToPreConsume = 42
	info.PriceData.FreeModel = true
	compatSnapshot := &billingexpr.BillingSnapshot{ModelName: "doubao-seedream-5.0-pro", TaskUsageBilling: true}
	compatInput := &billingexpr.RequestInput{Usage: map[string]any{"images_above_1_5k": float64(1)}}

	info.BeginSeedreamImageBilling(compatSnapshot, compatInput)
	require.True(t, info.SeedreamImageBilling)
	require.Same(t, compatSnapshot, info.TieredBillingSnapshot)
	info.PriceData.QuotaToPreConsume = 99
	info.PriceData.FreeModel = false

	info.ResetSeedreamImageBilling()

	require.False(t, info.SeedreamImageBilling)
	require.Same(t, baseSnapshot, info.TieredBillingSnapshot)
	require.Same(t, baseInput, info.BillingRequestInput)
	require.NotNil(t, info.BillingImageCount)
	require.Equal(t, baseImageCount, *info.BillingImageCount)
	require.Equal(t, 42, info.PriceData.QuotaToPreConsume)
	require.True(t, info.PriceData.FreeModel)
}
