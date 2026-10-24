package model

import (
	"reflect"
	"testing"
)

func TestResolvePricingEndpointTypesUsesExplicitModelConfiguration(t *testing.T) {
	inferred := []string{"openai"}
	got := resolvePricingEndpointTypes(inferred, `{"decision":{"path":"/v1/systemone","method":"POST"}}`)
	if !reflect.DeepEqual(got, []string{"decision"}) {
		t.Fatalf("explicit endpoint config should replace inferred endpoints: got %v", got)
	}
}

func TestResolvePricingEndpointTypesKeepsInferenceWithoutValidConfiguration(t *testing.T) {
	inferred := []string{"openai", "anthropic"}
	for _, configured := range []string{"", "not-json", `{}`, `{"ignored":null}`} {
		got := resolvePricingEndpointTypes(inferred, configured)
		if !reflect.DeepEqual(got, inferred) {
			t.Errorf("configured=%q: expected inferred endpoints %v, got %v", configured, inferred, got)
		}
	}
}

func TestResolvePricingEndpointTypesSortsConfiguredEndpoints(t *testing.T) {
	got := resolvePricingEndpointTypes(nil, `{"decision":"/v1/systemone","openai":"/v1/chat/completions"}`)
	want := []string{"decision", "openai"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected stable configured endpoint ordering %v, got %v", want, got)
	}
}
