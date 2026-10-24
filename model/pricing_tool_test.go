package model

import (
	"testing"

	"github.com/QuantumNous/new-api/constant"
	"github.com/stretchr/testify/require"
)

func TestGetPublicPricingToolPricesHidesResponsesBuiltInTools(t *testing.T) {
	tools := getPublicPricingToolPrices("gpt-4.1", []constant.EndpointType{
		constant.EndpointTypeOpenAI,
		constant.EndpointTypeOpenAIResponse,
	})

	require.Len(t, tools, 1)
	require.Equal(t, "web_search", tools[0].Name)
}

func TestGetPublicPricingToolPricesKeepsOtherEndpointRules(t *testing.T) {
	tests := []struct {
		name      string
		modelName string
		endpoints []constant.EndpointType
		want      []string
	}{
		{
			name:      "anthropic web search",
			modelName: "claude-sonnet",
			endpoints: []constant.EndpointType{constant.EndpointTypeAnthropic},
			want:      []string{"web_search"},
		},
		{
			name:      "gemini search",
			modelName: "gemini-2.5-flash",
			endpoints: []constant.EndpointType{constant.EndpointTypeGemini},
			want:      []string{"google_search", "google_search_grounded_prompt"},
		},
		{
			name:      "deepseek tools",
			modelName: "deepseek-v4-flash",
			endpoints: []constant.EndpointType{constant.EndpointTypeOpenAI},
			want:      []string{"web_search", "deepseek_search", "deepseek_fetch"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tools := getPublicPricingToolPrices(test.modelName, test.endpoints)
			got := make([]string, 0, len(tools))
			for _, tool := range tools {
				got = append(got, tool.Name)
			}
			require.Equal(t, test.want, got)
		})
	}
}
