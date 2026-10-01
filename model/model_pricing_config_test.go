package model

import (
	"testing"

	"github.com/QuantumNous/new-api/pkg/jsplugin"
	"github.com/QuantumNous/new-api/setting/billing_setting"
	"github.com/stretchr/testify/require"
)

func registerPricingValidationPlugin(t *testing.T, key, model string, field string) {
	t.Helper()
	source := `
export const meta = {
  apiVersion: 1,
  key: "` + key + `",
  name: "Pricing validation test",
  version: "1.0.0",
  author: {name: "Test"},
  models: ["` + model + `"],
  fetchMode: "per_task",
  usageProfiles: [{models: ["` + model + `"], schema: {` + field + `: {type: "number", unit: "count"}}}]
};
export function buildSubmitRequest() { return {}; }
export function parseSubmitResponse() { return {}; }
export function parseTaskResult() { return {}; }
export function buildQueryRequest() { return {}; }
`
	_, err := jsplugin.DefaultRegistry.Register(source, jsplugin.Options{})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, jsplugin.DefaultRegistry.Unregister(key)) })
}

func TestValidateModelPricingKeepsUnchangedSingleProviderPricesAfterProfileNarrowing(t *testing.T) {
	const pluginKey = "pricing-validation-single"
	const modelName = "pricing-validation-single-model"
	registerPricingValidationPlugin(t, pluginKey, modelName, "images")

	const expression = `u("image_count")`
	previousBase := PricingValues{"billing_setting.billing_expr": expression}
	unchangedBase := PricingValues{"billing_setting.billing_expr": expression}
	require.NoError(t, validateModelPricing(modelName, unchangedBase, previousBase))
	require.ErrorContains(t, validateModelPricing(modelName, PricingValues{
		"billing_setting.billing_expr": `u("image_count") * 2`,
	}, previousBase), "not declared")

	previousOverride := PricingValues{
		"billing_setting.billing_expr": expression,
		billing_setting.PluginBillingExprOption: map[string]any{pluginKey: expression},
	}
	unchangedOverride := PricingValues{
		"billing_setting.billing_expr": expression,
		billing_setting.PluginBillingExprOption: map[string]any{pluginKey: expression},
	}
	require.NoError(t, validateModelPricing(modelName, unchangedOverride, previousOverride))

	edited := PricingValues{
		"billing_setting.billing_expr": `u("image_count") * 2`,
		billing_setting.PluginBillingExprOption: map[string]any{pluginKey: `u("image_count") * 2`},
	}
	require.ErrorContains(t, validateModelPricing(modelName, edited, previous), "not declared")
}

func TestValidateModelPricingStillChecksUnchangedSharedProviderPrices(t *testing.T) {
	const modelName = "pricing-validation-shared-model"
	registerPricingValidationPlugin(t, "pricing-validation-alpha", modelName, "images")
	registerPricingValidationPlugin(t, "pricing-validation-beta", modelName, "credits")

	const expression = `u("image_count")`
	values := PricingValues{"billing_setting.billing_expr": expression}
	previous := PricingValues{"billing_setting.billing_expr": expression}
	require.ErrorContains(t, validateModelPricing(modelName, values, previous), "not declared")
}
