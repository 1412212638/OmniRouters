package model

import (
	"sort"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/pkg/billingexpr"
	"github.com/QuantumNous/new-api/pkg/jsplugin"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/setting/billing_setting"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
)

// TaskPricingWarnings is an administrator-only configuration diagnostic.
// It does not evaluate request-dependent prices or change runtime billing.
type TaskPricingWarnings struct {
	Enabled         bool                 `json:"enabled"`
	Warnings        []TaskPricingWarning `json:"warnings"`
	SkippedChannels []int                `json:"skipped_channels"`
}

type TaskPricingWarning struct {
	ModelName  string `json:"model_name"`
	PluginKey  string `json:"plugin_key"`
	PluginName string `json:"plugin_name"`
	ChannelIDs []int  `json:"channel_ids"`
	Reason     string `json:"reason"`
}

// Keep the selection and compatibility rules aligned with relay/relay_task.go.
// In particular, an explicit zero price is configured, and legacy prices are
// resolved using the public model name, not the channel mapping target.
func taskPricingProblem(g *jsplugin.RoutingGeneration, p *jsplugin.LoadedPlugin, name, mapped string) string {
	expr, found := billing_setting.ResolveTaskBillingExpr(p.Meta.Key, name, mapped)
	if found || billing_setting.GetBillingMode(name) == billing_setting.BillingModeTieredExpr {
		if !found || strings.TrimSpace(expr) == "" {
			return "missing_expression"
		}
		if _, err := billingexpr.CompileFromCache(expr); err != nil || billingexpr.UsesFixedPricing(expr) {
			return "invalid_expression"
		}
		_, override := billing_setting.GetPluginBillingExpr(p.Meta.Key, name)
		if !override && mapped != name {
			_, override = billing_setting.GetPluginBillingExpr(p.Meta.Key, mapped)
		}
		ordinary := !override && len(billingexpr.UsedUsageKeys(expr)) > 0 && (billing_setting.CanonicalOrdinaryResponseUsageModel(name) != "" || billing_setting.CanonicalOrdinaryResponseUsageModel(mapped) != "")
		if g.SharedModel(name) || g.SharedModel(mapped) || ordinary {
			schema, _ := p.Meta.UsageForModel(mapped)
			compatible := billing_setting.TaskExprCompatible(expr, schema)
			if ordinary {
				usageName := name
				if billing_setting.CanonicalOrdinaryResponseUsageModel(usageName) == "" {
					usageName = mapped
				}
				compatible = billing_setting.UsageExprCompatibleWithTaskSchema(usageName, expr, schema)
			}
			if !compatible {
				return "incompatible_expression"
			}
		}
		return ""
	}
	if _, ok := ratio_setting.GetModelPrice(name, false); ok {
		return ""
	}
	if _, ok := ratio_setting.GetDefaultModelPriceMap()[name]; ok {
		return ""
	}
	if _, ok, _ := ratio_setting.GetModelRatio(name); ok {
		return ""
	}
	return "missing_price"
}

func GetTaskPricingWarnings() (*TaskPricingWarnings, error) {
	result := &TaskPricingWarnings{Enabled: jsplugin.DefaultRegistry.Enabled(), Warnings: []TaskPricingWarning{}, SkippedChannels: []int{}}
	if !result.Enabled {
		return result, nil
	}
	var channels []Channel
	if err := DB.Select("id", "type", "models", "model_mapping", "setting").Where("status = ?", common.ChannelStatusEnabled).Order("id").Find(&channels).Error; err != nil {
		return nil, err
	}
	var abilities []Ability
	if err := DB.Select("channel_id", "model").Where("enabled = ?", true).Find(&abilities).Error; err != nil {
		return nil, err
	}
	enabled := make(map[int]map[string]bool)
	for _, a := range abilities {
		if enabled[a.ChannelId] == nil {
			enabled[a.ChannelId] = map[string]bool{}
		}
		enabled[a.ChannelId][a.Model] = true
	}
	g := jsplugin.DefaultRegistry.Generation()
	indices := make(map[string]int)
	for _, channel := range channels {
		// GetSetting repairs malformed data by writing it. Diagnostics must be
		// read-only, so parse directly instead.
		settings := dto.ChannelSettings{}
		mapping := map[string]string{}
		if (channel.Setting != nil && *channel.Setting != "" && common.UnmarshalJsonStr(*channel.Setting, &settings) != nil) ||
			(channel.GetModelMapping() != "" && common.UnmarshalJsonStr(channel.GetModelMapping(), &mapping) != nil) {
			result.SkippedChannels = append(result.SkippedChannels, channel.Id)
			continue
		}
		skipped := false
		seen := map[string]bool{}
		for _, name := range channel.GetModels() {
			if seen[name] || !enabled[channel.Id][name] {
				continue
			}
			seen[name] = true
			mapped, cycle := followChannelModelMapping(mapping, name)
			if cycle {
				skipped = true
				continue
			}
			declared, ok := g.CanonicalModel(mapped)
			if !ok {
				continue // An ordinary channel model is outside this diagnostic.
			}
			candidates := []*jsplugin.LoadedPlugin{}
			for _, p := range g.PluginsByModel(declared) {
				if channel.Type == constant.ChannelTypeTaskPlugin || channel.Type == constant.ChannelTypeNewAPI {
					if settings.BindsTaskPlugin(p.Meta.Key) {
						candidates = append(candidates, p)
					}
				} else if owner, ok := g.GetByChannelType(channel.Type); ok && owner == p {
					candidates = append(candidates, p)
				}
			}
			if len(candidates) > 1 {
				// Endpoint-specific ordering chooses the executor. Do not claim a
				// provider is missing a price without knowing which will execute.
				skipped = true
				continue
			}
			for _, p := range candidates {
				reason := taskPricingProblem(g, p, name, mapped)
				if reason == "" {
					continue
				}
				key := name + "\x00" + p.Meta.Key + "\x00" + reason
				if i, ok := indices[key]; ok {
					result.Warnings[i].ChannelIDs = append(result.Warnings[i].ChannelIDs, channel.Id)
				} else {
					indices[key] = len(result.Warnings)
					result.Warnings = append(result.Warnings, TaskPricingWarning{ModelName: name, PluginKey: p.Meta.Key, PluginName: p.Meta.Name, ChannelIDs: []int{channel.Id}, Reason: reason})
				}
			}
		}
		if skipped {
			result.SkippedChannels = append(result.SkippedChannels, channel.Id)
		}
	}
	sort.Slice(result.Warnings, func(i, j int) bool {
		a, b := result.Warnings[i], result.Warnings[j]
		if a.ModelName != b.ModelName {
			return a.ModelName < b.ModelName
		}
		if a.PluginKey != b.PluginKey {
			return a.PluginKey < b.PluginKey
		}
		return a.Reason < b.Reason
	})
	return result, nil
}
