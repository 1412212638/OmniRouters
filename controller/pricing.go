package controller

import (
	"sort"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/ratio_setting"

	"github.com/gin-gonic/gin"
)

func filterPricingByUsableGroups(pricing []model.Pricing, usableGroup map[string]string) []model.Pricing {
	if len(pricing) == 0 {
		return pricing
	}
	if len(usableGroup) == 0 {
		return []model.Pricing{}
	}

	filtered := make([]model.Pricing, 0, len(pricing))
	for _, item := range pricing {
		if common.StringsContains(item.EnableGroup, "all") {
			filtered = append(filtered, item)
			continue
		}
		for _, group := range item.EnableGroup {
			if _, ok := usableGroup[group]; ok {
				filtered = append(filtered, item)
				break
			}
		}
	}
	return filtered
}

func visiblePricingGroups(groups []string, usableGroup map[string]string) []string {
	if common.StringsContains(groups, "all") {
		groups = make([]string, 0, len(usableGroup))
		for group := range usableGroup {
			if ratio_setting.ContainsGroupRatio(group) {
				groups = append(groups, group)
			}
		}
		sort.Strings(groups)
		return groups
	}

	visible := make([]string, 0, len(groups))
	seen := make(map[string]struct{}, len(groups))
	for _, group := range groups {
		if _, ok := usableGroup[group]; !ok || !ratio_setting.ContainsGroupRatio(group) {
			continue
		}
		if _, ok := seen[group]; ok {
			continue
		}
		seen[group] = struct{}{}
		visible = append(visible, group)
	}
	return visible
}

func filterVisibleGroupRatios(ratios map[string]float64, usableGroup map[string]string) map[string]float64 {
	filtered := make(map[string]float64)
	for group, ratio := range ratios {
		if _, ok := usableGroup[group]; ok {
			filtered[group] = ratio
		}
	}
	return filtered
}

func filterVisibleModelRatios(ratios map[string]map[string]float64, usableGroup map[string]string) map[string]map[string]float64 {
	filtered := make(map[string]map[string]float64)
	for group, models := range ratios {
		if _, ok := usableGroup[group]; !ok {
			continue
		}
		filtered[group] = models
	}
	return filtered
}

func filterVisibleModelRatioExpiry(expiry map[string]map[string]int64, usableGroup map[string]string) map[string]map[string]int64 {
	filtered := make(map[string]map[string]int64)
	for group, models := range expiry {
		if _, ok := usableGroup[group]; !ok {
			continue
		}
		filtered[group] = models
	}
	return filtered
}

func GetPricing(c *gin.Context) {
	pricing := model.GetPricing()
	userId, exists := c.Get("id")
	usableGroup := map[string]string{}
	groupRatio := map[string]float64{}
	for s, f := range ratio_setting.GetGroupRatioCopy() {
		groupRatio[s] = f
	}
	var group string
	userID := 0
	var extraGroups []string
	if exists {
		userID = userId.(int)
		user, err := model.GetUserCache(userID)
		if err == nil {
			group = user.Group
			extraGroups = user.GetExtraGroups()
			for g := range groupRatio {
				ratio, ok := ratio_setting.GetGroupGroupRatio(group, g)
				if ok {
					groupRatio[g] = ratio
				}
			}
		}
	}

	usableGroup = service.GetUserUsableGroupsWithExtras(group, extraGroups)
	pricing = filterPricingByUsableGroups(pricing, usableGroup)
	groupRatio = filterVisibleGroupRatios(groupRatio, usableGroup)
	modelGroupRatios := filterVisibleModelRatios(ratio_setting.GetGroupModelRatioForUser(userID), usableGroup)
	modelGroupRatioExpiry := filterVisibleModelRatioExpiry(ratio_setting.GetGroupModelRatioExpiry(), usableGroup)
	for i := range pricing {
		pricing[i].EnableGroup = visiblePricingGroups(pricing[i].EnableGroup, usableGroup)
		resolved := make(map[string]float64, len(groupRatio))
		for g, base := range groupRatio {
			resolved[g] = base
			if _, ok := modelGroupRatios[g][pricing[i].ModelName]; ok {
				resolved[g] = modelGroupRatios[g][pricing[i].ModelName]
			}
		}
		pricing[i].ModelGroupRatio = resolved
		pricing[i].RateLimits = setting.GetModelRateLimits(pricing[i].EnableGroup)
	}

	c.JSON(200, gin.H{
		"success":                 true,
		"data":                    pricing,
		"vendors":                 model.GetVendors(),
		"group_ratio":             groupRatio,
		"group_model_ratio":       modelGroupRatios,
		"group_model_ratio_expiry": modelGroupRatioExpiry,
		"usable_group":             usableGroup,
		"supported_endpoint":       model.GetSupportedEndpointMap(),
		"auto_groups":              service.GetUserAutoGroupWithExtras(group, extraGroups),
		"pricing_version":           "group-model-ratio-v1",
	})
}

func ResetModelRatio(c *gin.Context) {
	defaultStr := ratio_setting.DefaultModelRatio2JSONString()
	err := model.UpdateOption("ModelRatio", defaultStr)
	if err != nil {
		c.JSON(200, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	err = ratio_setting.UpdateModelRatioByJSONString(defaultStr)
	if err != nil {
		c.JSON(200, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"success": true,
		"message": "重置模型倍率成功",
	})
}
