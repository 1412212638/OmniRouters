package dto

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetOpenAIChatCapabilities(t *testing.T) {
	tests := []struct {
		name, model, effort string
		max, developer, temperature, topP, logProbs bool
	}{
		{name: "gpt 6 sol defaults", model: "gpt-6-sol", max: true, developer: true, temperature: true, topP: true, logProbs: true},
		{name: "gpt 6 luna none", model: "gpt-6-luna", effort: "none", max: true, developer: true, temperature: true, topP: true, logProbs: true},
		{name: "gpt 6 luna reasoning", model: "gpt-6-luna", effort: "high", max: true, developer: true},
		{name: "gpt 6 astra", model: "gpt-6-astra", max: true, developer: true},
		{name: "gpt 7 stays unknown", model: "gpt-7", temperature: true, topP: true, logProbs: true},
		{name: "gpt 6 snapshot", model: "gpt-6-sol-2026-09-03", max: true, developer: true, temperature: true, topP: true, logProbs: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetOpenAIChatCapabilities(tt.model, tt.effort)
			require.Equal(t, tt.max, got.UseMaxCompletionTokens)
			require.Equal(t, tt.developer, got.UseDeveloperRole)
			require.Equal(t, tt.temperature, got.SupportsTemperature)
			require.Equal(t, tt.topP, got.SupportsTopP)
			require.Equal(t, tt.logProbs, got.SupportsLogProbs)
		})
	}
}
