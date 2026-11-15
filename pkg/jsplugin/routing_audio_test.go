package jsplugin

import (
	"net/http"
	"testing"
)

func TestSpeechHostProtocols(t *testing.T) {
	for _, test := range []struct {
		name string
		path string
	}{
		{name: ProtocolOpenAIAudio, path: "/v1/audio/speech"},
		{name: ProtocolMiniMaxT2A, path: "/v1/t2a_v2"},
	} {
		t.Run(test.name, func(t *testing.T) {
			definition, ok := HostProtocol(test.name)
			if !ok {
				t.Fatalf("speech host protocol %q was not registered", test.name)
			}
			if got := definition.DefinedModes(); len(got) != 1 || got[0].Name != "sync" {
				t.Fatalf("speech host protocol %q modes = %#v, want sync", test.name, got)
			}
			protocol, operation, ok := LookupHostProtocolOperation(http.MethodPost, test.path)
			if !ok || protocol != test.name {
				t.Fatalf("speech endpoint %s mapped to protocol=%q found=%t", test.path, protocol, ok)
			}
			if operation.Name != "create" || operation.ModelField != "model" {
				t.Fatalf("speech endpoint %s operation = %#v", test.path, operation)
			}
		})
	}
}
