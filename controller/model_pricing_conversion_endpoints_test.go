package controller

import (
	"net/http"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Model metadata may declare endpoints either as a type array or as a path
// map. Conversion must understand both forms so image billing metadata is not
// silently lost when the array form is used.
func TestModelPricingConversionAcceptsDeclaredEndpointArray(t *testing.T) {
	db := modelManagementDB(t, "sqlite", "")
	previousQuota := common.QuotaPerUnit
	common.QuotaPerUnit = 500000
	t.Cleanup(func() { common.QuotaPerUnit = previousQuota })

	for _, tc := range []struct {
		name        string
		endpoints   string
		draft       model.PricingValues
		wantImageCt bool
	}{
		{"array-endpoints-image", `["image-generation","openai"]`,
			model.PricingValues{"ModelRatio": 1.7875, "CompletionRatio": 6, "ImageRatio": 1.6}, false},
		{"array-endpoints-image-fixed", `["image-generation","openai"]`,
			model.PricingValues{"ModelPrice": 1}, true},
		{"array-endpoints-gemini", `["gemini","openai"]`,
			model.PricingValues{"ModelRatio": 0.10725, "CompletionRatio": 100}, false},
		{"array-endpoints-openai-response", `["openai","openai-response","openai-response-compact","gemini"]`,
			model.PricingValues{"ModelRatio": 0.17875, "CompletionRatio": 120}, false},
		{"map-endpoints-image", `{"image-generation": {"path": "/v1/images/generations", "method": "POST"}}`,
			model.PricingValues{"ModelPrice": 1}, true},
		{"array-endpoints-image-padded", `[" image-generation ","openai"]`,
			model.PricingValues{"ModelPrice": 1}, true},
		{"map-endpoints-image-padded", `{"image-generation ": {"path": "/v1/images/generations", "method": "POST"}}`,
			model.PricingValues{"ModelPrice": 1}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, db.Create(&model.Model{
				ModelName: tc.name, Endpoints: tc.endpoints, Status: 1, NameRule: model.NameRuleExact,
			}).Error)
			var response struct {
				Success bool
				Data    model.ModelPricingConversion
			}
			modelManagementRequest(t, PreviewModelPricingConversion, http.MethodPost,
				"/api/option/model_pricing/convert",
				map[string]any{"model_name": tc.name, "pricing": tc.draft}, &response)
			require.True(t, response.Success)
			require.Empty(t, response.Data.UnsupportedReason)
			require.NotEmpty(t, response.Data.Expression)
			assert.Equal(t, tc.wantImageCt, response.Data.BillingDetails.ImageCount)
		})
	}
}
