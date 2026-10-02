package jsplugin

import (
	"net/http"
	"testing"
)

func TestOpenAIImageHostProtocol(t *testing.T) {
	definition, ok := HostProtocol(ProtocolOpenAIImage)
	if !ok {
		t.Fatal("openai_image host protocol must be registered")
	}
	if len(definition.Operations) != 2 {
		t.Fatalf("expected two OpenAI Images operations, got %d", len(definition.Operations))
	}

	generation, _, ok := LookupHostProtocolOperation(http.MethodPost, "/v1/images/generations")
	if !ok || generation != ProtocolOpenAIImage {
		t.Fatalf("generation endpoint was not mapped to %q: protocol=%q found=%t", ProtocolOpenAIImage, generation, ok)
	}
	edit, _, ok := LookupHostProtocolOperation(http.MethodPost, "/v1/images/edits")
	if !ok || edit != ProtocolOpenAIImage {
		t.Fatalf("edit endpoint was not mapped to %q: protocol=%q found=%t", ProtocolOpenAIImage, edit, ok)
	}
}
