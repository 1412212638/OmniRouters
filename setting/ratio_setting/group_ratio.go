package ratio_setting

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/config"
	"github.com/QuantumNous/new-api/types"
)

var defaultGroupRatio = map[string]float64{
	"default": 1,
	"vip":     1,
	"svip":    1,
}

var groupRatioMap = types.NewRWMap[string, float64]()

var defaultGroupGroupRatio = map[string]map[string]float64{
	"vip": {
		"edit_this": 0.9,
	},
}

var groupGroupRatioMap = types.NewRWMap[string, map[string]float64]()

var groupModelRatioMap = types.NewRWMap[string, map[string]float64]()
var groupModelUserRatioMap = types.NewRWMap[string, map[string]map[int]float64]()
var groupModelRatioExpiryMap = types.NewRWMap[string, map[string]int64]()

var defaultGroupSpecialUsableGroup = map[string]map[string]string{}

type GroupRatioSetting struct {
	GroupRatio              *types.RWMap[string, float64]                    `json:"group_ratio"`
	GroupGroupRatio         *types.RWMap[string, map[string]float64]         `json:"group_group_ratio"`
	GroupModelRatio         *types.RWMap[string, map[string]float64]         `json:"group_model_ratio"`
	GroupModelUserRatio     *types.RWMap[string, map[string]map[int]float64] `json:"group_model_user_ratio"`
	GroupModelRatioExpiry   *types.RWMap[string, map[string]int64]           `json:"group_model_ratio_expiry"`
	GroupSpecialUsableGroup *types.RWMap[string, map[string]string]          `json:"group_special_usable_group"`
}

var groupRatioSetting GroupRatioSetting

func init() {
	groupSpecialUsableGroup := types.NewRWMap[string, map[string]string]()
	groupSpecialUsableGroup.AddAll(defaultGroupSpecialUsableGroup)

	groupRatioMap.AddAll(defaultGroupRatio)
	groupGroupRatioMap.AddAll(defaultGroupGroupRatio)

	groupRatioSetting = GroupRatioSetting{
		GroupSpecialUsableGroup: groupSpecialUsableGroup,
		GroupRatio:              groupRatioMap,
		GroupGroupRatio:         groupGroupRatioMap,
		GroupModelRatio:         groupModelRatioMap,
		GroupModelUserRatio:     groupModelUserRatioMap,
		GroupModelRatioExpiry:   groupModelRatioExpiryMap,
	}

	config.GlobalConfig.Register("group_ratio_setting", &groupRatioSetting)
}

func GetGroupRatioSetting() *GroupRatioSetting {
	if groupRatioSetting.GroupSpecialUsableGroup == nil {
		groupRatioSetting.GroupSpecialUsableGroup = types.NewRWMap[string, map[string]string]()
		groupRatioSetting.GroupSpecialUsableGroup.AddAll(defaultGroupSpecialUsableGroup)
	}
	return &groupRatioSetting
}

func GetGroupRatioCopy() map[string]float64 {
	return groupRatioMap.ReadAll()
}

func ContainsGroupRatio(name string) bool {
	_, ok := groupRatioMap.Get(name)
	return ok
}

func GroupRatio2JSONString() string {
	return groupRatioMap.MarshalJSONString()
}

func UpdateGroupRatioByJSONString(jsonStr string) error {
	return types.LoadFromJsonString(groupRatioMap, jsonStr)
}

func GetGroupRatio(name string) float64 {
	ratio, ok := groupRatioMap.Get(name)
	if !ok {
		common.SysLog("group ratio not found: " + name)
		return 1
	}
	return ratio
}

func GetGroupGroupRatio(userGroup, usingGroup string) (float64, bool) {
	gp, ok := groupGroupRatioMap.Get(userGroup)
	if !ok {
		return -1, false
	}
	ratio, ok := gp[usingGroup]
	if !ok {
		return -1, false
	}
	return ratio, true
}

// GetGroupModelRatio resolves the model-specific multiplier. A user-specific
// rule wins over the group's general model rule; missing rules inherit 1.
func ResolveGroupModelRatio(group, model string, userID int) (ratio float64, matched bool, userSpecific bool) {
	if isGroupModelRatioExpired(group, model, time.Now().Unix()) {
		return 1, false, false
	}
	if userID > 0 {
		if models, ok := groupModelUserRatioMap.Get(group); ok {
			if users, ok := models[model]; ok {
				if ratio, ok := users[userID]; ok {
					return ratio, true, true
				}
			}
		}
	}
	if models, ok := groupModelRatioMap.Get(group); ok {
		if ratio, ok := models[model]; ok {
			return ratio, true, false
		}
	}
	return 1, false, false
}

func GetGroupModelRatio(group, model string, userID int) (float64, bool) {
	ratio, matched, _ := ResolveGroupModelRatio(group, model, userID)
	return ratio, matched
}

func GetGroupModelRatioForUser(userID int) map[string]map[string]float64 {
	result := make(map[string]map[string]float64)
	now := time.Now().Unix()
	for group, models := range groupModelRatioMap.ReadAll() {
		result[group] = make(map[string]float64, len(models))
		for model, ratio := range models {
			if isGroupModelRatioExpired(group, model, now) {
				continue
			}
			result[group][model] = ratio
		}
	}
	if userID <= 0 {
		return result
	}
	for group, models := range groupModelUserRatioMap.ReadAll() {
		if result[group] == nil {
			result[group] = make(map[string]float64)
		}
		for model, users := range models {
			if isGroupModelRatioExpired(group, model, now) {
				continue
			}
			if ratio, ok := users[userID]; ok {
				result[group][model] = ratio
			}
		}
	}
	return result
}

func isGroupModelRatioExpired(group, model string, now int64) bool {
	models, ok := groupModelRatioExpiryMap.Get(group)
	if !ok {
		return false
	}
	expiresAt, ok := models[model]
	return ok && expiresAt > 0 && now >= expiresAt
}

func CheckGroupModelRatio(jsonStr string) error {
	var ratios map[string]map[string]float64
	if err := common.UnmarshalJsonStr(jsonStr, &ratios); err != nil {
		return err
	}
	for group, models := range ratios {
		if group == "" {
			return errors.New("group model ratio group must not be empty")
		}
		for model, ratio := range models {
			if model == "" || ratio < 0 || math.IsNaN(ratio) || math.IsInf(ratio, 0) {
				return fmt.Errorf("invalid group model ratio: %s/%s", group, model)
			}
		}
	}
	return nil
}

func CheckGroupModelUserRatio(jsonStr string) error {
	var ratios map[string]map[string]map[int]float64
	if err := common.UnmarshalJsonStr(jsonStr, &ratios); err != nil {
		return err
	}
	for group, models := range ratios {
		if group == "" {
			return errors.New("user group model ratio group must not be empty")
		}
		for model, users := range models {
			if model == "" {
				return fmt.Errorf("invalid user group model ratio model: %s", group)
			}
			for userID, ratio := range users {
				if userID <= 0 || ratio < 0 || math.IsNaN(ratio) || math.IsInf(ratio, 0) {
					return fmt.Errorf("invalid user group model ratio: %s/%s/%d", group, model, userID)
				}
			}
		}
	}
	return nil
}

func CheckGroupModelRatioExpiry(jsonStr string) error {
	var expiries map[string]map[string]int64
	if err := common.UnmarshalJsonStr(jsonStr, &expiries); err != nil {
		return err
	}
	for group, models := range expiries {
		if group == "" {
			return errors.New("group model ratio expiry group must not be empty")
		}
		for model, expiresAt := range models {
			if model == "" || expiresAt < 0 {
				return fmt.Errorf("invalid group model ratio expiry: %s/%s", group, model)
			}
		}
	}
	return nil
}

func GroupGroupRatio2JSONString() string {
	return groupGroupRatioMap.MarshalJSONString()
}

func UpdateGroupGroupRatioByJSONString(jsonStr string) error {
	return types.LoadFromJsonString(groupGroupRatioMap, jsonStr)
}

func CheckGroupRatio(jsonStr string) error {
	checkGroupRatio := make(map[string]float64)
	err := common.UnmarshalJsonStr(jsonStr, &checkGroupRatio)
	if err != nil {
		return err
	}
	for name, ratio := range checkGroupRatio {
		if ratio < 0 {
			return errors.New("group ratio must be not less than 0: " + name)
		}
	}
	return nil
}
