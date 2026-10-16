package model

import (
	"math"
	"regexp"
	"strings"

	"github.com/QuantumNous/new-api/pkg/billingexpr"
)

// PublicBillingSummary is the non-executable pricing information needed by
// the public pricing catalogue. It deliberately contains no expression text,
// request paths, headers, plugin schema, or rule expiry data.
type PublicBillingSummary struct {
	Mode              string   `json:"mode"`
	InputPrice        *float64 `json:"input_price,omitempty"`
	OutputPrice       *float64 `json:"output_price,omitempty"`
	CacheReadPrice    *float64 `json:"cache_read_price,omitempty"`
	CacheCreatePrice  *float64 `json:"cache_create_price,omitempty"`
	ImageInputPrice   *float64 `json:"image_input_price,omitempty"`
	ImageOutputPrice  *float64 `json:"image_output_price,omitempty"`
	AudioInputPrice   *float64 `json:"audio_input_price,omitempty"`
	AudioOutputPrice  *float64 `json:"audio_output_price,omitempty"`
	RequestPrice      *float64 `json:"request_price,omitempty"`
	HasRequestRules   bool     `json:"has_request_rules,omitempty"`
	HasUsagePricing   bool     `json:"has_usage_pricing,omitempty"`
	TierCount         int      `json:"tier_count,omitempty"`
	Rules             []PublicBillingRule `json:"rules,omitempty"`
}

// PublicBillingRule contains only display-safe request-rule metadata. The
// original predicate is intentionally omitted so parameter paths and other
// administrator-specific billing configuration are not exposed publicly.
type PublicBillingRule struct {
	Multiplier         float64                  `json:"multiplier"`
	Conditions         []PublicBillingCondition `json:"conditions,omitempty"`
	HasOtherConditions bool                     `json:"has_other_conditions,omitempty"`
}

type PublicBillingCondition struct {
	TimeFunc   string `json:"time_func"`
	Timezone   string `json:"timezone"`
	Mode       string `json:"mode"`
	Value      string `json:"value,omitempty"`
	RangeStart string `json:"range_start,omitempty"`
	RangeEnd   string `json:"range_end,omitempty"`
}

var publicTimeConditionPattern = regexp.MustCompile(`(?i)(hour|minute|weekday|month|day)\("([^"]*)"\)\s*(==|>=|<=|>|<)\s*(-?[0-9]+)`)

type publicPriceProbe struct {
	name   string
	params billingexpr.TokenParams
}

// BuildPublicBillingSummary evaluates an expression with isolated probe
// inputs and returns only display rates. Request conditions are reduced to
// safe time-window metadata; their source predicates and parameter paths stay
// server-side.
func BuildPublicBillingSummary(expression string) *PublicBillingSummary {
	expression = strings.TrimSpace(expression)
	if expression == "" {
		return nil
	}

	summary := &PublicBillingSummary{
		HasRequestRules: strings.Contains(expression, "param(") ||
			strings.Contains(expression, "header(") ||
			strings.Contains(expression, "hour(") ||
			strings.Contains(expression, "minute(") ||
			strings.Contains(expression, "weekday(") ||
			strings.Contains(expression, "month(") ||
			strings.Contains(expression, "day("),
		HasUsagePricing: strings.Contains(expression, "u("),
		TierCount:       strings.Count(expression, "tier("),
	}
	if summary.HasUsagePricing {
		summary.Mode = "usage"
	} else {
		summary.Mode = "token"
	}
	summary.Rules = buildPublicBillingRules(expression)

	probes := []publicPriceProbe{
		{name: "input", params: billingexpr.TokenParams{P: 1_000_000, Len: 1_000_000}},
		{name: "output", params: billingexpr.TokenParams{C: 1_000_000, Len: 1_000_000}},
		{name: "cache_read", params: billingexpr.TokenParams{CR: 1_000_000, Len: 1_000_000}},
		{name: "cache_create", params: billingexpr.TokenParams{CC: 1_000_000, Len: 1_000_000}},
		{name: "image_input", params: billingexpr.TokenParams{Img: 1_000_000, Len: 1_000_000}},
		{name: "image_output", params: billingexpr.TokenParams{ImgO: 1_000_000, Len: 1_000_000}},
		{name: "audio_input", params: billingexpr.TokenParams{AI: 1_000_000, Len: 1_000_000}},
		{name: "audio_output", params: billingexpr.TokenParams{AO: 1_000_000, Len: 1_000_000}},
	}

	for _, probe := range probes {
		value, trace, ok := runPublicBillingProbe(expression, probe.params)
		if !ok || trace.BillingUnit != billingexpr.BillingUnitToken || value <= 0 {
			continue
		}
		switch probe.name {
		case "input":
			summary.InputPrice = floatPtr(value)
		case "output":
			summary.OutputPrice = floatPtr(value)
		case "cache_read":
			summary.CacheReadPrice = floatPtr(value)
		case "cache_create":
			summary.CacheCreatePrice = floatPtr(value)
		case "image_input":
			summary.ImageInputPrice = floatPtr(value)
		case "image_output":
			summary.ImageOutputPrice = floatPtr(value)
		case "audio_input":
			summary.AudioInputPrice = floatPtr(value)
		case "audio_output":
			summary.AudioOutputPrice = floatPtr(value)
		}
	}

	if value, trace, ok := runPublicBillingProbe(expression, billingexpr.TokenParams{}); ok && trace.BillingUnit == billingexpr.BillingUnitRequest && value > 0 {
		summary.Mode = "request"
		summary.RequestPrice = floatPtr(value)
	}

	if summary.InputPrice == nil && summary.OutputPrice == nil && summary.RequestPrice == nil && !summary.HasUsagePricing {
		return &PublicBillingSummary{Mode: "dynamic", HasRequestRules: summary.HasRequestRules, TierCount: summary.TierCount, Rules: summary.Rules}
	}
	return summary
}

func buildPublicBillingRules(expression string) []PublicBillingRule {
	_, trace, err := billingexpr.RunExpr(expression, billingexpr.TokenParams{})
	if err != nil || len(trace.RequestRules) == 0 {
		return nil
	}

	rules := make([]PublicBillingRule, 0, len(trace.RequestRules))
	for _, traceRule := range trace.RequestRules {
		conditions, hasOther := publicBillingConditions(traceRule.Cond)
		if len(conditions) == 0 && !hasOther {
			continue
		}
		rules = append(rules, PublicBillingRule{
			Multiplier:         traceRule.Multiplier,
			Conditions:         conditions,
			HasOtherConditions: hasOther,
		})
	}
	return rules
}

func publicBillingConditions(condition string) ([]PublicBillingCondition, bool) {
	matches := publicTimeConditionPattern.FindAllStringSubmatch(condition, -1)
	conditions := make([]PublicBillingCondition, 0, len(matches))
	for _, match := range matches {
		mode := map[string]string{"==": "eq", ">": "gt", ">=": "gte", "<": "lt", "<=": "lte"}[match[3]]
		if mode == "" {
			continue
		}
		conditions = append(conditions, PublicBillingCondition{
			TimeFunc: match[1],
			Timezone: match[2],
			Mode:     mode,
			Value:    match[4],
		})
	}

	// Pair adjacent bounds into the same range used by the visual pricing
	// editor, e.g. hour >= 9 && hour < 12.
	merged := make([]PublicBillingCondition, 0, len(conditions))
	for i := 0; i < len(conditions); i++ {
		current := conditions[i]
		if i+1 < len(conditions) {
			next := conditions[i+1]
			if current.TimeFunc == next.TimeFunc &&
				current.Timezone == next.Timezone &&
				current.Mode == "gte" && next.Mode == "lt" {
				current.Mode = "range"
				current.RangeStart = current.Value
				current.RangeEnd = next.Value
				current.Value = ""
				i++
			}
		}
		merged = append(merged, current)
	}

	// Only time predicates are included in the public response. A rule with
	// body/header/plugin predicates is represented by a generic marker.
	hasOther := len(matches) == 0 || strings.Contains(condition, "param(") || strings.Contains(condition, "header(") || strings.Contains(condition, "u(") || strings.Contains(condition, "has(")
	return merged, hasOther
}

func runPublicBillingProbe(expression string, params billingexpr.TokenParams) (float64, billingexpr.TraceResult, bool) {
	value, trace, err := billingexpr.RunExpr(expression, params)
	if err != nil || math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
		return 0, trace, false
	}
	// Time-based request rules are evaluated during the probe. Remove the
	// matched multiplier so the public summary remains a base-price summary;
	// the frontend applies the safe rule metadata to derive the displayed range.
	matchedMultiplier := 1.0
	for _, rule := range trace.RequestRules {
		if !rule.Matched {
			continue
		}
		if rule.Multiplier <= 0 ||
			math.IsNaN(rule.Multiplier) ||
			math.IsInf(rule.Multiplier, 0) {
			return 0, trace, false
		}
		matchedMultiplier *= rule.Multiplier
	}
	if matchedMultiplier > 0 && matchedMultiplier != 1 {
		value /= matchedMultiplier
	}
	// v1 token expressions use USD per million tokens; fixed request prices
	// are also normalized to the same raw million-unit scale by the runtime.
	return value / 1_000_000, trace, true
}

func floatPtr(value float64) *float64 {
	return &value
}
