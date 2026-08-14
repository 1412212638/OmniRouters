package perfmetrics

import "testing"

func TestRecentBucketsFillsEmptyHourlyBuckets(t *testing.T) {
	const endTs int64 = 24 * 3600
	buckets := map[int64]counters{
		endTs - 3600: {
			requestCount: 4,
			successCount: 3,
		},
	}

	result := recentBuckets(buckets, endTs, 24)
	if len(result) != 24 {
		t.Fatalf("expected 24 buckets, got %d", len(result))
	}
	if result[22].Ts != endTs-3600 {
		t.Fatalf("expected penultimate bucket timestamp %d, got %d", endTs-3600, result[22].Ts)
	}
	if result[22].RequestCount != 4 || result[22].SuccessRate != 75 {
		t.Fatalf("unexpected populated bucket: %+v", result[22])
	}
	if result[23].RequestCount != 0 || result[23].SuccessRate != 0 {
		t.Fatalf("expected current empty bucket, got %+v", result[23])
	}
}
