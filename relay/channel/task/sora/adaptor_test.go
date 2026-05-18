package sora

import "testing"

func TestIsSoraAudioGenerationEnabled(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  bool
	}{
		{name: "bool true", value: true, want: true},
		{name: "enabled string", value: "Enabled", want: true},
		{name: "true string", value: "true", want: true},
		{name: "typo string", value: "ture", want: true},
		{name: "numeric one", value: float64(1), want: true},
		{name: "disabled string", value: "disabled", want: false},
		{name: "false string", value: "false", want: false},
		{name: "nil", value: nil, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSoraAudioGenerationEnabled(tt.value); got != tt.want {
				t.Fatalf("isSoraAudioGenerationEnabled(%v) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}
