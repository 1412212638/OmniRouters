package billing_setting

import (
	"fmt"
	"sort"
	"strings"

	"github.com/QuantumNous/new-api/common"
)

type SoraResolutionTier struct {
	Value      string  `json:"value"`
	Multiplier float64 `json:"multiplier"`
}

type SoraPerRequestPricing struct {
	Enabled         bool                 `json:"enabled"`
	ResolutionTiers []SoraResolutionTier `json:"resolution_tiers"`
}

func (s SoraPerRequestPricing) FindResolutionMultiplier(resolution string) (float64, bool) {
	resolution = strings.TrimSpace(resolution)
	if resolution == "" {
		return 0, false
	}
	for _, tier := range s.ResolutionTiers {
		if strings.TrimSpace(tier.Value) == resolution {
			return tier.Multiplier, true
		}
	}
	return 0, false
}

func (s SoraPerRequestPricing) ResolutionValues() []string {
	values := make([]string, 0, len(s.ResolutionTiers))
	for _, tier := range s.ResolutionTiers {
		value := strings.TrimSpace(tier.Value)
		if value != "" {
			values = append(values, value)
		}
	}
	sort.Strings(values)
	return values
}

func GetSoraPerRequestPricing(model string) (SoraPerRequestPricing, bool) {
	rule, ok := billingSetting.SoraPerRequestPricing[model]
	if !ok {
		return SoraPerRequestPricing{}, false
	}
	return rule, true
}

func ValidateSoraPerRequestPricingJSONString(jsonStr string) error {
	if strings.TrimSpace(jsonStr) == "" {
		return nil
	}

	rules := make(map[string]SoraPerRequestPricing)
	if err := common.Unmarshal([]byte(jsonStr), &rules); err != nil {
		return err
	}

	for modelName, rule := range rules {
		if err := validateSoraPerRequestPricingRule(modelName, rule); err != nil {
			return err
		}
	}
	return nil
}

func validateSoraPerRequestPricingRule(modelName string, rule SoraPerRequestPricing) error {
	seen := make(map[string]struct{}, len(rule.ResolutionTiers))
	normalizedCount := 0
	for _, tier := range rule.ResolutionTiers {
		value := strings.TrimSpace(tier.Value)
		if value == "" {
			return fmt.Errorf("model %s has an empty resolution tier", modelName)
		}
		if tier.Multiplier <= 0 {
			return fmt.Errorf("model %s resolution %s must use a multiplier > 0", modelName, value)
		}
		if _, exists := seen[value]; exists {
			return fmt.Errorf("model %s has duplicated resolution tier %s", modelName, value)
		}
		seen[value] = struct{}{}
		normalizedCount++
	}

	if rule.Enabled && normalizedCount == 0 {
		return fmt.Errorf("model %s must define at least one resolution tier when Sora parameter pricing is enabled", modelName)
	}

	return nil
}
