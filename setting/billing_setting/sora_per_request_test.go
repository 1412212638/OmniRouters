package billing_setting

import "testing"

func TestValidateSoraPerRequestPricingAllowsAudioGenerationMultiplier(t *testing.T) {
	jsonStr := `{
		"sora-2": {
			"enabled": true,
			"resolution_tiers": [
				{"value": "720p", "multiplier": 1},
				{"value": "1080p", "multiplier": 1.5}
			],
			"audio_generation_multiplier": 1.3
		}
	}`

	if err := ValidateSoraPerRequestPricingJSONString(jsonStr); err != nil {
		t.Fatalf("expected valid sora pricing, got %v", err)
	}
}

func TestValidateSoraPerRequestPricingRejectsDiscountAudioGenerationMultiplier(t *testing.T) {
	jsonStr := `{
		"sora-2": {
			"enabled": true,
			"resolution_tiers": [
				{"value": "720p", "multiplier": 1}
			],
			"audio_generation_multiplier": 0.9
		}
	}`

	if err := ValidateSoraPerRequestPricingJSONString(jsonStr); err == nil {
		t.Fatal("expected audio generation multiplier below 1 to be rejected")
	}
}

func TestSoraAudioGenerationRatioIgnoresDiscountMultiplier(t *testing.T) {
	multiplier := 0.9
	rule := SoraPerRequestPricing{AudioGenerationMultiplier: &multiplier}

	if _, ok := rule.AudioGenerationRatio(); ok {
		t.Fatal("expected multiplier below 1 to be ignored")
	}
}

func TestSoraAudioGenerationRatioReturnsConfiguredMultiplier(t *testing.T) {
	multiplier := 1.3
	rule := SoraPerRequestPricing{AudioGenerationMultiplier: &multiplier}

	got, ok := rule.AudioGenerationRatio()
	if !ok || got != multiplier {
		t.Fatalf("AudioGenerationRatio() = %v, %v; want %v, true", got, ok, multiplier)
	}
}
