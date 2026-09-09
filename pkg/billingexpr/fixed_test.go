package billingexpr

import (
	"math"
	"testing"
)

func TestFixedRequestPricing(t *testing.T) {
	const expression = `len <= 32000 ? tier("short", fixed(0.01)) : tier("long", p * 2)`
	cost, trace, err := RunExpr(expression, TokenParams{})
	if err != nil {
		t.Fatal(err)
	}
	if cost != 10000 || trace.BillingUnit != BillingUnitRequest || trace.FixedPrice == nil || *trace.FixedPrice != 0.01 {
		t.Fatalf("unexpected fixed result: cost=%v trace=%+v", cost, trace)
	}
	if !UsesFixedPricing(expression) {
		t.Fatal("fixed pricing was not detected")
	}
}

func TestFixedRequestPricingRejectsUnsafeExpressions(t *testing.T) {
	for _, expression := range []string{
		`tier("bad", fixed(-0.01))`,
		`tier("bad", fixed(p))`,
		`tier("bad", fixed(0.01) + p * 2)`,
		`tier("bad", fixed(0.01)) * p`,
		`tier("bad", fixed(1e308))`,
	} {
		if _, err := CompileFromCache(expression); err == nil {
			t.Errorf("expected invalid fixed expression to fail: %s", expression)
		}
	}
	if _, _, err := RunExpr(`tier("huge", fixed(1e20))`, TokenParams{}); err != nil {
		t.Fatal(err)
	}
	result, err := ComputeTieredQuota(&BillingSnapshot{
		ExprString:  `tier("huge", fixed(1e20))`,
		ExprHash:    ExprHashString(`tier("huge", fixed(1e20))`),
		GroupRatio:  1,
		QuotaPerUnit: 500000,
	}, TokenParams{})
	if err != nil || result.ActualQuotaAfterGroup != math.MaxInt32 || result.Clamp == nil {
		t.Fatalf("expected saturation, result=%+v err=%v", result, err)
	}
}
