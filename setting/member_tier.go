package setting

import (
	"fmt"
	"strings"
	"sync"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
)

type MemberTierRule struct {
	Enabled       bool   `json:"enabled"`
	Group         string `json:"group"`
	DisplayName   string `json:"display_name"`
	Description   string `json:"description,omitempty"`
	MinTopupQuota int64  `json:"min_topup_quota"`
	MinUsedQuota  int64  `json:"min_used_quota"`
}

var (
	memberTierRules      = []MemberTierRule{}
	memberTierRulesMutex sync.RWMutex
)

func GetMemberTierRulesCopy() []MemberTierRule {
	memberTierRulesMutex.RLock()
	defer memberTierRulesMutex.RUnlock()

	rules := make([]MemberTierRule, len(memberTierRules))
	copy(rules, memberTierRules)
	return rules
}

func MemberTierRules2JSONString() string {
	memberTierRulesMutex.RLock()
	defer memberTierRulesMutex.RUnlock()

	jsonBytes, err := common.Marshal(memberTierRules)
	if err != nil {
		common.SysLog("error marshalling member tier rules: " + err.Error())
		return "[]"
	}
	return string(jsonBytes)
}

func UpdateMemberTierRulesByJSONString(jsonStr string) error {
	rules, err := ParseMemberTierRulesJSONString(jsonStr)
	if err != nil {
		return err
	}

	memberTierRulesMutex.Lock()
	defer memberTierRulesMutex.Unlock()
	memberTierRules = rules
	return nil
}

func ParseMemberTierRulesJSONString(jsonStr string) ([]MemberTierRule, error) {
	raw := strings.TrimSpace(jsonStr)
	if raw == "" {
		raw = "[]"
	}

	var rules []MemberTierRule
	if err := common.UnmarshalJsonStr(raw, &rules); err != nil {
		return nil, err
	}

	normalized := make([]MemberTierRule, 0, len(rules))
	seenGroups := make(map[string]struct{}, len(rules))
	for idx, rule := range rules {
		rule.Group = strings.TrimSpace(rule.Group)
		rule.DisplayName = strings.TrimSpace(rule.DisplayName)
		rule.Description = strings.TrimSpace(rule.Description)
		if rule.Group == "" {
			return nil, fmt.Errorf("member tier rule #%d group is required", idx+1)
		}
		if rule.DisplayName == "" {
			rule.DisplayName = rule.Group
		}
		if rule.MinTopupQuota < 0 || rule.MinUsedQuota < 0 {
			return nil, fmt.Errorf("member tier rule #%d thresholds must be non-negative", idx+1)
		}
		if _, exists := seenGroups[rule.Group]; exists {
			return nil, fmt.Errorf("member tier group duplicated: %s", rule.Group)
		}
		if !ratio_setting.ContainsGroupRatio(rule.Group) {
			return nil, fmt.Errorf("member tier group does not exist in group ratios: %s", rule.Group)
		}
		seenGroups[rule.Group] = struct{}{}
		normalized = append(normalized, rule)
	}
	return normalized, nil
}

func CheckMemberTierRulesJSONString(jsonStr string) error {
	_, err := ParseMemberTierRulesJSONString(jsonStr)
	return err
}
