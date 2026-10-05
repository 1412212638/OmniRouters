package billingexpr

import (
	"testing"
	"time"
)

func TestTimeInZoneAtShanghaiPeakWindow(t *testing.T) {
	// 2026-10-05 is Monday. The second DeepSeek peak window starts at 14:00
	// Shanghai time, so a 15:13 request must be evaluated as Monday afternoon.
	now := time.Date(2026, time.October, 5, 7, 13, 0, 0, time.UTC)
	local := timeInZoneAt(now, "Asia/Shanghai")
	if local.Weekday() != time.Monday || local.Hour() != 15 || local.Minute() != 13 {
		t.Fatalf("Shanghai time = %s, want Monday 15:13", local.Format(time.RFC3339))
	}
}

func TestTimeInZoneAtInvalidZoneFallsBackToUTC(t *testing.T) {
	now := time.Date(2026, time.October, 5, 7, 13, 0, 0, time.UTC)
	got := timeInZoneAt(now, "Invalid/Zone")
	if !got.Equal(now.UTC()) {
		t.Fatalf("invalid zone = %s, want UTC %s", got, now.UTC())
	}
}
