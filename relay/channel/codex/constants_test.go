package codex

import (
	"testing"

	"github.com/QuantumNous/new-api/setting/ratio_setting"
	"github.com/stretchr/testify/assert"
)

func TestModelListPreservesLegacyAndDiscoveredVariants(t *testing.T) {
	models := make(map[string]struct{}, len(ModelList))
	for _, modelName := range ModelList {
		if _, exists := models[modelName]; exists {
			t.Fatalf("duplicate Codex model: %s", modelName)
		}
		models[modelName] = struct{}{}
	}

	for _, modelName := range []string{
		"gpt-5-codex",
		"gpt-5.2-codex",
		"gpt-5.4-mini",
		"gpt-5.5",
		"codex-auto-review",
	} {
		assert.Contains(t, models, modelName)
	}
	assert.Contains(t, models, ratio_setting.WithCompactModelSuffix("gpt-5-codex"))
	assert.NotContains(t, models, ratio_setting.WithCompactModelSuffix("codex-auto-review"))
}
