package reasoning

import (
	"strings"

	kitreasoning "github.com/QuantumNous/new-api/relaykit/relayconvert/reasoning"
	"github.com/QuantumNous/new-api/setting/model_setting"
)

type ModelModifier struct {
	Key   string
	Value string
}

type ModelModifierSpec struct {
	Base      string
	Modifiers []ModelModifier
}

func (s ModelModifierSpec) HasModifiers() bool {
	return len(s.Modifiers) > 0
}

// ParseModelModifiers removes a contiguous trailing chain of @key:value segments.
func ParseModelModifiers(modelName string) ModelModifierSpec {
	spec := ModelModifierSpec{Base: modelName}
	parts := strings.Split(modelName, "@")
	if len(parts) < 2 {
		return spec
	}
	firstModifier := len(parts)
	for i := len(parts) - 1; i > 0; i-- {
		key, value, ok := parseModelModifierSegment(parts[i])
		if !ok {
			break
		}
		firstModifier = i
		spec.Modifiers = append([]ModelModifier{{Key: key, Value: value}}, spec.Modifiers...)
	}
	if firstModifier == len(parts) {
		return spec
	}
	base := strings.Join(parts[:firstModifier], "@")
	if base != "" {
		spec.Base = base
	}
	return spec
}

func parseModelModifierSegment(segment string) (string, string, bool) {
	colon := strings.IndexByte(segment, ':')
	if colon <= 0 {
		return "", "", false
	}
	key := segment[:colon]
	for i, r := range key {
		letter := r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z'
		if i == 0 && !letter || i > 0 && !letter && (r < '0' || r > '9') && r != '_' && r != '-' {
			return "", "", false
		}
	}
	return strings.ToLower(key), segment[colon+1:], true
}

func ParseLegacyModelSuffix(modelName string, allowClaudeThinkingAlias bool, allowGeminiThinkingAlias bool) (string, kitreasoning.Intent, bool, error) {
	prefix, bare := splitModelNamespace(modelName)
	var base string
	var intent kitreasoning.Intent
	var found bool
	var err error
	switch {
	case strings.HasPrefix(bare, "claude-"):
		base, intent, found, err = kitreasoning.ParseClaudeModelSuffix(bare, allowClaudeThinkingAlias)
	case strings.HasPrefix(bare, "gemini-"):
		base, intent, found, err = kitreasoning.ParseGeminiModelSuffix(bare, allowGeminiThinkingAlias)
	default:
		effort, openAIBase := ParseOpenAIReasoningEffortFromModelSuffix(bare)
		if effort == "" {
			return modelName, kitreasoning.Intent{}, false, nil
		}
		parsedEffort, parseErr := kitreasoning.ParseEffort(effort)
		if parseErr != nil {
			return modelName, kitreasoning.Intent{}, false, parseErr
		}
		mode := kitreasoning.ModeEnabled
		if parsedEffort == kitreasoning.EffortNone {
			mode = kitreasoning.ModeDisabled
		}
		base = openAIBase
		intent = kitreasoning.Intent{Mode: mode, Effort: parsedEffort, Source: kitreasoning.SourceSuffix}
		found = true
	}
	if err != nil || !found {
		return modelName, kitreasoning.Intent{}, false, err
	}
	return prefix + base, intent, true, nil
}

func BaseModelName(modelName string) string {
	if model_setting.ShouldPreserveThinkingSuffix(modelName) {
		return modelName
	}
	base := ParseModelModifiers(modelName).Base
	if model_setting.ShouldPreserveThinkingSuffix(base) {
		return base
	}
	legacyBase, _, found, err := ParseLegacyModelSuffix(base, model_setting.GetClaudeSettings().ThinkingAdapterEnabled, model_setting.GetGeminiSettings().ThinkingAdapterEnabled)
	if err == nil && found {
		return legacyBase
	}
	return base
}

func CanonicalBillingModelNames(modelName string) []string {
	if model_setting.ShouldPreserveThinkingSuffix(modelName) {
		return nil
	}
	spec := ParseModelModifiers(modelName)
	base := spec.Base
	intent, hasThinking := billingIntentFromModifiers(spec)
	if !model_setting.ShouldPreserveThinkingSuffix(base) {
		legacyBase, legacyIntent, found, err := ParseLegacyModelSuffix(base, model_setting.GetClaudeSettings().ThinkingAdapterEnabled, model_setting.GetGeminiSettings().ThinkingAdapterEnabled)
		if err == nil && found {
			base = legacyBase
			if !hasThinking {
				intent, hasThinking = legacyIntent, true
			}
		}
	}
	if !hasThinking {
		return nil
	}
	thinking, effort, ok := normalizeBillingThinking(intent)
	if !ok || base == "" {
		return nil
	}
	var names []string
	if effort != "" {
		names = append(names, base+"@effort:"+effort+"@thinking:"+thinking)
	}
	names = append(names, base+"@thinking:"+thinking)
	return names
}

func splitModelNamespace(modelName string) (string, string) {
	if slash := strings.LastIndex(modelName, "/"); slash >= 0 {
		return modelName[:slash+1], modelName[slash+1:]
	}
	return "", modelName
}

func billingIntentFromModifiers(spec ModelModifierSpec) (kitreasoning.Intent, bool) {
	last := make(map[string]int, len(spec.Modifiers))
	for index, modifier := range spec.Modifiers {
		last[modifier.Key] = index
	}
	var intent kitreasoning.Intent
	hasThinking := false
	for index, modifier := range spec.Modifiers {
		if last[modifier.Key] != index {
			continue
		}
		switch modifier.Key {
		case "thinking":
			parsed, ok := kitreasoning.ParseThinkingModifier(modifier.Value)
			if ok {
				intent, hasThinking = parsed, true
			}
		case "effort":
			effort, err := kitreasoning.ParseEffort(modifier.Value)
			if err != nil || effort == "" {
				continue
			}
			intent = kitreasoning.Intent{Mode: kitreasoning.ModeEnabled, Effort: effort, Source: kitreasoning.SourceSuffix}
			if effort == kitreasoning.EffortNone {
				intent.Mode = kitreasoning.ModeDisabled
			}
			hasThinking = true
		}
	}
	return intent, hasThinking
}

func normalizeBillingThinking(intent kitreasoning.Intent) (string, string, bool) {
	if intent.Effort == kitreasoning.EffortNone || intent.Mode == kitreasoning.ModeDisabled || intent.BudgetTokens != nil && *intent.BudgetTokens == 0 {
		return "off", "", true
	}
	if intent.Effort != "" && intent.Effort != kitreasoning.EffortNone {
		return "on", strings.ToLower(string(intent.Effort)), true
	}
	if intent.Mode == kitreasoning.ModeEnabled || intent.Mode == kitreasoning.ModeAdaptive || intent.BudgetTokens != nil {
		return "on", "", true
	}
	return "", "", false
}
