package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/pkg/jsplugin"
	"github.com/QuantumNous/new-api/setting/billing_setting"
	"github.com/QuantumNous/new-api/setting/config"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestTaskPricingWarningsEnabledAliasesAndReadOnly(t *testing.T) {
	const plugin, name, alias = "warning-regression-plugin", "warning-regression-model", "warning-regression-alias"
	registerPricingValidationPlugin(t, plugin, name, "images")
	oldEnabled := jsplugin.DefaultRegistry.Enabled()
	jsplugin.DefaultRegistry.SetEnabled(true)
	t.Cleanup(func() { jsplugin.DefaultRegistry.SetEnabled(oldEnabled) })
	oldDB := DB
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	DB = db
	t.Cleanup(func() { DB = oldDB })
	require.NoError(t, db.AutoMigrate(&Channel{}, &Ability{}))
	settings := `{"task_plugin_key":"warning-regression-plugin"}`
	broken := `{broken`
	mapping := `{"warning-regression-alias":"middle","middle":"warning-regression-model"}`
	channels := []Channel{
		{Id: 1, Type: constant.ChannelTypeTaskPlugin, Status: common.ChannelStatusEnabled, Models: alias, ModelMapping: &mapping, Setting: &settings},
		{Id: 2, Type: constant.ChannelTypeTaskPlugin, Status: common.ChannelStatusEnabled, Models: alias, ModelMapping: &mapping, Setting: &broken},
		{Id: 3, Type: constant.ChannelTypeTaskPlugin, Status: common.ChannelStatusManuallyDisabled, Models: alias, ModelMapping: &mapping, Setting: &settings},
		{Id: 4, Type: constant.ChannelTypeTaskPlugin, Status: common.ChannelStatusEnabled, Models: alias, ModelMapping: &mapping, Setting: &settings},
		{Id: 5, Type: constant.ChannelTypeOpenAI, Status: common.ChannelStatusEnabled, Models: alias, ModelMapping: &mapping},
		{Id: 6, Type: constant.ChannelTypeTaskPlugin, Status: common.ChannelStatusEnabled, Models: alias, ModelMapping: &mapping, Setting: &settings},
	}
	for i := range channels {
		require.NoError(t, db.Create(&channels[i]).Error)
		require.NoError(t, db.Create(&Ability{Group: "default", Model: alias, ChannelId: channels[i].Id, Enabled: i != 3}).Error)
	}
	// Multiple enabled groups must not duplicate a channel in a warning.
	require.NoError(t, db.Create(&Ability{Group: "premium", Model: alias, ChannelId: 1, Enabled: true}).Error)
	result, err := GetTaskPricingWarnings()
	require.NoError(t, err)
	require.Len(t, result.Warnings, 1)
	require.Equal(t, alias, result.Warnings[0].ModelName)
	require.Equal(t, "missing_price", result.Warnings[0].Reason)
	require.Equal(t, []int{1, 6}, result.Warnings[0].ChannelIDs)
	require.Equal(t, []int{2}, result.SkippedChannels)
	var saved Channel
	require.NoError(t, db.First(&saved, 2).Error)
	require.Equal(t, broken, *saved.Setting)

	prices := ratio_setting.ModelPrice2JSONString()
	t.Cleanup(func() { require.NoError(t, ratio_setting.UpdateModelPriceByJSONString(prices)) })
	// A price on the mapping target does not cover a legacy alias price.
	require.NoError(t, ratio_setting.UpdateModelPriceByJSONString(`{"warning-regression-model":0.5}`))
	result, err = GetTaskPricingWarnings()
	require.NoError(t, err)
	require.Len(t, result.Warnings, 1)
	require.NoError(t, ratio_setting.UpdateModelPriceByJSONString(`{"warning-regression-alias":0}`))
	result, err = GetTaskPricingWarnings()
	require.NoError(t, err)
	require.Empty(t, result.Warnings)
	ratios := ratio_setting.ModelRatio2JSONString()
	t.Cleanup(func() { require.NoError(t, ratio_setting.UpdateModelRatioByJSONString(ratios)) })
	require.NoError(t, ratio_setting.UpdateModelPriceByJSONString(`{}`))
	require.NoError(t, ratio_setting.UpdateModelRatioByJSONString(`{"warning-regression-alias":0}`))
	result, err = GetTaskPricingWarnings()
	require.NoError(t, err)
	require.Empty(t, result.Warnings)
	jsplugin.DefaultRegistry.SetEnabled(false)
	result, err = GetTaskPricingWarnings()
	require.NoError(t, err)
	require.False(t, result.Enabled)
	require.Empty(t, result.Warnings)
}

func TestTaskPricingWarningsReportAmbiguousBindingsAndMappingCycles(t *testing.T) {
	const name = "warning-ambiguous-model"
	registerPricingValidationPlugin(t, "warning-ambiguous-alpha", name, "images")
	registerPricingValidationPlugin(t, "warning-ambiguous-beta", name, "seconds")
	oldEnabled, oldDB := jsplugin.DefaultRegistry.Enabled(), DB
	jsplugin.DefaultRegistry.SetEnabled(true)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	DB = db
	t.Cleanup(func() { DB = oldDB; jsplugin.DefaultRegistry.SetEnabled(oldEnabled) })
	require.NoError(t, db.AutoMigrate(&Channel{}, &Ability{}))
	settings := `{"task_plugin_key":"warning-ambiguous-alpha","task_extend_plugin_keys":["warning-ambiguous-beta"]}`
	cycle := `{"warning-ambiguous-model":"middle","middle":"warning-ambiguous-model"}`
	for id := 1; id <= 2; id++ {
		channel := Channel{Id: id, Type: constant.ChannelTypeTaskPlugin, Status: common.ChannelStatusEnabled, Models: name, Setting: &settings}
		if id == 2 {
			channel.ModelMapping = &cycle
		}
		require.NoError(t, db.Create(&channel).Error)
		require.NoError(t, db.Create(&Ability{Group: "default", Model: name, ChannelId: id, Enabled: true}).Error)
	}
	result, err := GetTaskPricingWarnings()
	require.NoError(t, err)
	require.Empty(t, result.Warnings)
	require.Equal(t, []int{1, 2}, result.SkippedChannels)
}

func TestTaskPricingProblemExpressionPrecedenceAndSharedSchemas(t *testing.T) {
	const name, alias = "warning-shared-model", "warning-shared-alias"
	const alpha, beta = "warning-alpha", "warning-beta"
	registerPricingValidationPlugin(t, alpha, name, "images")
	registerPricingValidationPlugin(t, beta, name, "seconds")
	g := jsplugin.DefaultRegistry.Generation()
	p, ok := g.Get(alpha)
	require.True(t, ok)
	settings := config.GlobalConfig.Get("billing_setting").(*billing_setting.BillingSetting)
	previous := *settings
	settings.BillingMode = map[string]string{name: billing_setting.BillingModeTieredExpr}
	settings.BillingExpr = map[string]string{name: `u("images") * 0.3`}
	settings.PluginBillingExpr = map[string]string{}
	t.Cleanup(func() { *settings = previous })
	require.Empty(t, taskPricingProblem(g, p, alias, name)) // Mapped expression fallback.
	b, _ := g.Get(beta)
	require.Equal(t, "incompatible_expression", taskPricingProblem(g, b, alias, name))
	settings.PluginBillingExpr[billing_setting.PluginBillingExprKey(beta, name)] = `u("seconds") * 0.2`
	require.Empty(t, taskPricingProblem(g, b, alias, name))
	settings.PluginBillingExpr[billing_setting.PluginBillingExprKey(beta, alias)] = `0`
	require.Empty(t, taskPricingProblem(g, b, alias, name)) // Explicit free override wins.
	settings.PluginBillingExpr[billing_setting.PluginBillingExprKey(beta, alias)] = `fixed(0.1)`
	require.Equal(t, "invalid_expression", taskPricingProblem(g, b, alias, name))
	settings.PluginBillingExpr[billing_setting.PluginBillingExprKey(beta, alias)] = `u(`
	require.Equal(t, "invalid_expression", taskPricingProblem(g, b, alias, name))
	settings.PluginBillingExpr[billing_setting.PluginBillingExprKey(beta, alias)] = ""
	require.Equal(t, "missing_expression", taskPricingProblem(g, b, alias, name))
	delete(settings.PluginBillingExpr, billing_setting.PluginBillingExprKey(beta, alias))
	delete(settings.PluginBillingExpr, billing_setting.PluginBillingExprKey(beta, name))
	settings.BillingMode[alias] = billing_setting.BillingModeTieredExpr
	require.Equal(t, "missing_expression", taskPricingProblem(g, b, alias, name)) // Explicit alias mode blocks fallback.
}

func TestTaskPricingProblemOrdinaryUsageDoesNotOverridePluginPricing(t *testing.T) {
	const name, key = "doubao-seedream-5.0-pro", "warning-ordinary-profile"
	registerPricingValidationPlugin(t, key, name, "images")
	g := jsplugin.DefaultRegistry.Generation()
	p, ok := g.Get(key)
	require.True(t, ok)
	settings := config.GlobalConfig.Get("billing_setting").(*billing_setting.BillingSetting)
	previous := *settings
	settings.BillingMode = map[string]string{name: billing_setting.BillingModeTieredExpr}
	settings.BillingExpr = map[string]string{name: `u("generated_images") * 0.3`}
	settings.PluginBillingExpr = map[string]string{}
	t.Cleanup(func() { *settings = previous })
	require.Equal(t, "incompatible_expression", taskPricingProblem(g, p, name, name))
	settings.PluginBillingExpr[billing_setting.PluginBillingExprKey(key, name)] = `u("images") * 0.3`
	require.Empty(t, taskPricingProblem(g, p, name, name))
}
