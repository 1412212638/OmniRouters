package perfmetrics

import "testing"

func TestCacheRateUsesCachedAndUncachedInput(t *testing.T) {
	value := counters{inputTokens: 1073, cachedTokens: 229376}
	rate := cacheRate(value)
	if rate == nil || *rate < 99.52 || *rate > 99.54 {
		t.Fatalf("cache rate = %v, want about 99.53%%", rate)
	}
}
