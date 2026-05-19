package billing_setting

import "testing"

func TestValidateSoraPerRequestPricingAllowsAudioGenerationSurcharge(t *testing.T) {
	jsonStr := `{
		"sora-2": {
			"enabled": true,
			"resolution_tiers": [
				{"value": "720p", "multiplier": 1},
				{"value": "1080p", "multiplier": 1.5}
			],
			"audio_generation_surcharge": 0.05
		}
	}`

	if err := ValidateSoraPerRequestPricingJSONString(jsonStr); err != nil {
		t.Fatalf("expected valid sora pricing, got %v", err)
	}
}

func TestValidateSoraPerRequestPricingRejectsNegativeAudioGenerationSurcharge(t *testing.T) {
	jsonStr := `{
		"sora-2": {
			"enabled": true,
			"resolution_tiers": [
				{"value": "720p", "multiplier": 1}
			],
			"audio_generation_surcharge": -0.01
		}
	}`

	if err := ValidateSoraPerRequestPricingJSONString(jsonStr); err == nil {
		t.Fatal("expected negative audio generation surcharge to be rejected")
	}
}

func TestSoraAudioGenerationSurchargeIgnoresZeroPrice(t *testing.T) {
	surcharge := 0.0
	rule := SoraPerRequestPricing{AudioGenerationSurcharge: &surcharge}

	if _, ok := rule.AudioGenerationSurchargePrice(); ok {
		t.Fatal("expected zero surcharge to be ignored")
	}
}

func TestSoraAudioGenerationSurchargeReturnsConfiguredPrice(t *testing.T) {
	surcharge := 0.05
	rule := SoraPerRequestPricing{AudioGenerationSurcharge: &surcharge}

	got, ok := rule.AudioGenerationSurchargePrice()
	if !ok || got != surcharge {
		t.Fatalf("AudioGenerationSurchargePrice() = %v, %v; want %v, true", got, ok, surcharge)
	}
}
