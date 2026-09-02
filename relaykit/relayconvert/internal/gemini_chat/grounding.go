package geminichat

import "github.com/QuantumNous/new-api/relaykit/dto"

func GroundingWebSearchQueries(response *dto.GeminiChatResponse) []string {
	if response == nil || response.Candidates == nil {
		return nil
	}
	seen := map[string]struct{}{}
	queries := make([]string, 0)
	for _, candidate := range response.Candidates {
		if candidate.GroundingMetadata == nil {
			continue
		}
		for _, query := range candidate.GroundingMetadata.WebSearchQueries {
			if query == "" {
				continue
			}
			if _, ok := seen[query]; ok {
				continue
			}
			seen[query] = struct{}{}
			queries = append(queries, query)
		}
	}
	return queries
}
