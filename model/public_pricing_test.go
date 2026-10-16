package model

import (
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
)

func TestBuildPublicBillingSummaryTokenPrices(t *testing.T) {
	summary := BuildPublicBillingSummary(`tier("base", p * 2.5 + c * 15 + cr * 0.25)`)
	if summary == nil {
		t.Fatal("expected public billing summary")
	}
	if summary.Mode != "token" {
		t.Fatalf("mode = %q, want token", summary.Mode)
	}
	if summary.InputPrice == nil || *summary.InputPrice != 2.5 {
		t.Fatalf("input price = %v, want 2.5", summary.InputPrice)
	}
	if summary.OutputPrice == nil || *summary.OutputPrice != 15 {
		t.Fatalf("output price = %v, want 15", summary.OutputPrice)
	}
	if summary.CacheReadPrice == nil || *summary.CacheReadPrice != 0.25 {
		t.Fatalf("cache read price = %v, want 0.25", summary.CacheReadPrice)
	}
	if summary.HasRequestRules {
		t.Fatal("plain token pricing should not report request rules")
	}
}

func TestBuildPublicBillingSummaryRequestPrice(t *testing.T) {
	summary := BuildPublicBillingSummary(`tier("request", fixed(0.01))`)
	if summary == nil {
		t.Fatal("expected public billing summary")
	}
	if summary.Mode != "request" {
		t.Fatalf("mode = %q, want request", summary.Mode)
	}
	if summary.RequestPrice == nil || *summary.RequestPrice != 0.01 {
		t.Fatalf("request price = %v, want 0.01", summary.RequestPrice)
	}
	if summary.InputPrice != nil || summary.OutputPrice != nil {
		t.Fatal("request pricing should not expose token prices")
	}
}

func TestBuildPublicBillingSummaryHidesRequestRules(t *testing.T) {
	expression := `tier("base", p * 2.5 + c * 15) * (param("resolution") == "4K" ? 2 : 1)`
	summary := BuildPublicBillingSummary(expression)
	if summary == nil {
		t.Fatal("expected public billing summary")
	}
	if !summary.HasRequestRules {
		t.Fatal("expected request rules marker")
	}
	encoded, err := common.Marshal(summary)
	if err != nil {
		t.Fatalf("marshal summary: %v", err)
	}
	if strings.Contains(string(encoded), "resolution") || strings.Contains(string(encoded), "4K") {
		t.Fatalf("public summary leaked request rule: %s", encoded)
	}
	if len(summary.Rules) != 1 || !summary.Rules[0].HasOtherConditions {
		t.Fatalf("summary rules = %+v, want one opaque request rule", summary.Rules)
	}
}

func TestBuildPublicBillingSummaryIncludesSafeTimeRules(t *testing.T) {
	expression := `tier("base", p * 1 + c * 4) * (weekday("Asia/Shanghai") >= 1 && weekday("Asia/Shanghai") < 6 && hour("Asia/Shanghai") >= 9 && hour("Asia/Shanghai") < 12 ? 2 : 1)`
	summary := BuildPublicBillingSummary(expression)
	if summary == nil || len(summary.Rules) != 1 {
		t.Fatalf("summary rules = %+v, want one safe rule", summary)
	}
	rule := summary.Rules[0]
	if rule.Multiplier != 2 || len(rule.Conditions) != 2 {
		t.Fatalf("rule = %+v, want weekday/hour ranges with 2x", rule)
	}
	if summary.InputPrice == nil || *summary.InputPrice != 1 {
		t.Fatalf("input price = %v, want base price 1", summary.InputPrice)
	}
	if summary.OutputPrice == nil || *summary.OutputPrice != 4 {
		t.Fatalf("output price = %v, want base price 4", summary.OutputPrice)
	}
	if rule.Conditions[0].Mode != "range" || rule.Conditions[0].RangeStart != "1" || rule.Conditions[0].RangeEnd != "6" {
		t.Fatalf("weekday condition = %+v", rule.Conditions[0])
	}
	encoded, err := common.Marshal(summary)
	if err != nil {
		t.Fatalf("marshal summary: %v", err)
	}
	if strings.Contains(string(encoded), "tier(\"") || strings.Contains(string(encoded), "weekday(\"") {
		t.Fatalf("public summary leaked expression source: %s", encoded)
	}
}

func TestBuildPublicBillingSummaryUsagePricingDoesNotLeakExpression(t *testing.T) {
	expression := `tier("video", u("duration") * 0.02)`
	summary := BuildPublicBillingSummary(expression)
	if summary == nil {
		t.Fatal("expected public billing summary")
	}
	if !summary.HasUsagePricing || summary.Mode != "usage" {
		t.Fatalf("summary = %+v, want usage pricing", summary)
	}
	encoded, err := common.Marshal(summary)
	if err != nil {
		t.Fatalf("marshal summary: %v", err)
	}
	if strings.Contains(string(encoded), "duration") || strings.Contains(string(encoded), expression) {
		t.Fatalf("public summary leaked usage expression: %s", encoded)
	}
}
