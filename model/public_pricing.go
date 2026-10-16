package model

import (
	"math"
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
}

type publicPriceProbe struct {
	name   string
	params billingexpr.TokenParams
}

// BuildPublicBillingSummary evaluates an expression with isolated probe
// inputs and returns only display rates. Request conditions are intentionally
// summarized as a boolean; their source and exact predicates stay server-side.
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
		return &PublicBillingSummary{Mode: "dynamic", HasRequestRules: summary.HasRequestRules, TierCount: summary.TierCount}
	}
	return summary
}

func runPublicBillingProbe(expression string, params billingexpr.TokenParams) (float64, billingexpr.TraceResult, bool) {
	value, trace, err := billingexpr.RunExpr(expression, params)
	if err != nil || math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
		return 0, trace, false
	}
	// v1 token expressions use USD per million tokens; fixed request prices
	// are also normalized to the same raw million-unit scale by the runtime.
	return value / 1_000_000, trace, true
}

func floatPtr(value float64) *float64 {
	return &value
}
