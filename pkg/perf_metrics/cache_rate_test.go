package perfmetrics

import (
	"math"
	"testing"
)

func TestQueryResultPreservesCacheAcrossBuckets(t *testing.T) {
	merged := map[bucketKey]counters{}
	// The first bucket represents persisted data, the second live samples.
	mergeCounters(merged, bucketKey{model: "test", group: "default", bucketTs: 3600}, counters{
		requestCount: 1, inputTokens: 3171, cachedTokens: 96768,
	})
	var live atomicBucket
	live.add(Sample{InputTokens: 1073, CachedTokens: 229376})
	mergeCounters(merged, bucketKey{model: "test", group: "default", bucketTs: 7200}, live.snapshot())
	// Requests without a cache hit do not change the hit-only denominator.
	mergeCounters(merged, bucketKey{model: "test", group: "default", bucketTs: 10800}, counters{requestCount: 1})
	result := buildQueryResult("test", merged)
	if len(result.Groups) != 1 {
		t.Fatalf("groups = %d, want 1", len(result.Groups))
	}
	got := result.Groups[0].CacheRate
	want := float64(96768+229376) / float64(99939+230449) * 100
	if got == nil || math.Abs(*got-want) > 0.000001 {
		t.Fatalf("cache rate = %v, want %f", got, want)
	}
	if result.Groups[0].Series[0].CacheRate == nil {
		t.Fatal("bucket cache rate was lost")
	}
}

func TestQueryResultWithoutCacheRemainsUnknown(t *testing.T) {
	result := buildQueryResult("test", map[bucketKey]counters{
		{model: "test", group: "default", bucketTs: 3600}: {requestCount: 1},
	})
	if result.Groups[0].CacheRate != nil {
		t.Fatal("missing cache data must not become a fabricated rate")
	}
}

func TestCacheRateUsesCachedAndUncachedInput(t *testing.T) {
	value := counters{inputTokens: 1073, cachedTokens: 229376}
	rate := cacheRate(value)
	if rate == nil || *rate < 99.52 || *rate > 99.54 {
		t.Fatalf("cache rate = %v, want about 99.53%%", rate)
	}
}
