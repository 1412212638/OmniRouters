package service

import (
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
	"github.com/gin-gonic/gin"
)

func GetUserUsableGroups(userGroup string) map[string]string {
	return GetUserUsableGroupsWithExtras(userGroup, nil)
}

func GetUserUsableGroupsWithExtras(userGroup string, extraGroups []string) map[string]string {
	groupsCopy := setting.GetUserUsableGroupsCopy()
	if userGroup != "" {
		specialSettings, b := ratio_setting.GetGroupRatioSetting().GroupSpecialUsableGroup.Get(userGroup)
		if b {
			// 处理特殊可用分组
			for specialGroup, desc := range specialSettings {
				if strings.HasPrefix(specialGroup, "-:") {
					// 移除分组
					groupToRemove := strings.TrimPrefix(specialGroup, "-:")
					delete(groupsCopy, groupToRemove)
				} else if strings.HasPrefix(specialGroup, "+:") {
					// 添加分组
					groupToAdd := strings.TrimPrefix(specialGroup, "+:")
					groupsCopy[groupToAdd] = desc
				} else {
					// 直接添加分组
					groupsCopy[specialGroup] = desc
				}
			}
		}
		// 如果userGroup不在UserUsableGroups中，返回UserUsableGroups + userGroup
		if _, ok := groupsCopy[userGroup]; !ok {
			groupsCopy[userGroup] = "用户分组"
		}
	}
	for _, group := range extraGroups {
		group = strings.TrimSpace(group)
		if group == "" || group == "auto" || group == userGroup {
			continue
		}
		if !ratio_setting.ContainsGroupRatio(group) {
			continue
		}
		groupsCopy[group] = setting.GetUsableGroupDescription(group)
	}
	return groupsCopy
}

func GroupInUserUsableGroups(userGroup, groupName string) bool {
	_, ok := GetUserUsableGroups(userGroup)[groupName]
	return ok
}

func GroupInUserUsableGroupsForContext(c *gin.Context, groupName string) bool {
	_, ok := GetUserUsableGroupsForContext(c)[groupName]
	return ok
}

func IsUserSelectableGroup(userGroup, groupName string) bool {
	return isUserSelectableGroupWithExtras(userGroup, groupName, nil)
}

func isUserSelectableGroupWithExtras(userGroup, groupName string, extraGroups []string) bool {
	if groupName == "" || groupName == "auto" {
		return false
	}
	_, ok := GetUserUsableGroupsWithExtras(userGroup, extraGroups)[groupName]
	return ok && ratio_setting.ContainsGroupRatio(groupName)
}

func GetUserUsableGroupsForContext(c *gin.Context) map[string]string {
	userGroup := common.GetContextKeyString(c, constant.ContextKeyUserGroup)
	return GetUserUsableGroupsWithExtras(userGroup, getUserExtraGroupsFromContext(c))
}

// GetUserAutoGroup 根据用户分组获取自动分组设置
func GetUserAutoGroup(userGroup string) []string {
	return GetUserAutoGroupWithExtras(userGroup, nil)
}

func GetUserAutoGroupWithExtras(userGroup string, extraGroups []string) []string {
	groups := GetUserUsableGroupsWithExtras(userGroup, extraGroups)
	autoGroups := make([]string, 0)
	seen := make(map[string]struct{})
	for _, group := range setting.GetAutoGroups() {
		if _, ok := groups[group]; !ok || !isUserSelectableGroupWithExtras(userGroup, group, extraGroups) {
			continue
		}
		if _, ok := seen[group]; ok {
			continue
		}
		seen[group] = struct{}{}
		autoGroups = append(autoGroups, group)
	}
	return autoGroups
}

func GetUserAutoGroupForContext(c *gin.Context) []string {
	userGroup := common.GetContextKeyString(c, constant.ContextKeyUserGroup)
	return GetUserAutoGroupWithExtras(userGroup, getUserExtraGroupsFromContext(c))
}

func FilterUserTokenAutoGroups(userGroup string, groups []string) []string {
	return filterUserTokenAutoGroups(userGroup, groups, nil)
}

func filterUserTokenAutoGroups(userGroup string, groups, extraGroups []string) []string {
	maxCount := setting.GetMaxTokenAutoGroups()
	filtered := make([]string, 0, min(len(groups), maxCount))
	seen := make(map[string]struct{})
	for _, group := range groups {
		if !isUserSelectableGroupWithExtras(userGroup, group, extraGroups) {
			continue
		}
		if _, ok := seen[group]; ok {
			continue
		}
		seen[group] = struct{}{}
		filtered = append(filtered, group)
		if len(filtered) == maxCount {
			break
		}
	}
	return filtered
}

// GetRequestAutoGroups resolves the token-specific Auto order while applying
// the user's current base and extra-group permissions.
func GetRequestAutoGroups(c *gin.Context, userGroup string) []string {
	value, ok := common.GetContextKey(c, constant.ContextKeyTokenAutoGroups)
	if !ok {
		return GetUserAutoGroupWithExtras(userGroup, getUserExtraGroupsFromContext(c))
	}
	groups, ok := value.([]string)
	if !ok {
		return []string{}
	}
	return filterUserTokenAutoGroups(userGroup, groups, getUserExtraGroupsFromContext(c))
}

func GetGroupsEnabledModels(groups []string) []string {
	seen := make(map[string]struct{})
	models := make([]string, 0)
	for _, group := range groups {
		for _, modelName := range model.GetGroupEnabledModels(group) {
			if _, ok := seen[modelName]; ok {
				continue
			}
			seen[modelName] = struct{}{}
			models = append(models, modelName)
		}
	}
	return models
}

func getUserExtraGroupsFromContext(c *gin.Context) []string {
	value, ok := common.GetContextKey(c, constant.ContextKeyUserExtraGroups)
	if !ok {
		return nil
	}
	groups, ok := value.([]string)
	if !ok {
		return nil
	}
	return groups
}

// GetUserGroupRatio 获取用户使用某个分组的倍率
// userGroup 用户分组
// group 需要获取倍率的分组
func GetUserGroupRatio(userGroup, group string) float64 {
	ratio, ok := ratio_setting.GetGroupGroupRatio(userGroup, group)
	if ok {
		return ratio
	}
	return ratio_setting.GetGroupRatio(group)
}
