package controller

import (
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
)

func TestPublicPricingPayloadRemovesInternalBillingConfiguration(t *testing.T) {
	payload, err := publicPricingPayload(model.Pricing{
		ModelName:            "internal-model",
		EnableGroup:          []string{"default", "vip"},
		BillingMode:          "tiered_expr",
		BillingExpr:          `tier("base", p * 2) ||| when(param("resolution") == "4K") * 2`,
		OwnerBy:              "internal-owner",
	}, map[string]string{"default": "Default"})
	if err != nil {
		t.Fatalf("build public payload: %v", err)
	}
	for _, field := range []string{
		"billing_expr",
		"billing_usage_schema",
		"billing_usage_examples",
		"billing_plugin_variants",
		"owner_by",
	} {
		if _, ok := payload[field]; ok {
			t.Fatalf("public payload contains internal field %q", field)
		}
	}
	groups, ok := payload["enable_groups"].([]string)
	if !ok || len(groups) != 1 || groups[0] != "default" {
		t.Fatalf("visible groups = %#v, want [default]", payload["enable_groups"])
	}
	encoded, err := common.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal public payload: %v", err)
	}
	if strings.Contains(string(encoded), "resolution") || strings.Contains(string(encoded), "internal-owner") {
		t.Fatalf("public payload leaked internal configuration: %s", encoded)
	}
	if _, ok := payload["billing_summary"]; !ok {
		t.Fatal("public payload is missing billing summary")
	}
}
