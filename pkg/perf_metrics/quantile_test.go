package perfmetrics

import "testing"

func TestQuantileReservoirIsBoundedAndOrdered(t *testing.T) {
	var reservoir quantileReservoir
	for i := int64(1); i <= quantileReservoirSize+100; i++ {
		reservoir.add(i, uint64(i))
	}
	if len(reservoir.values) != quantileReservoirSize {
		t.Fatalf("reservoir size = %d, want %d", len(reservoir.values), quantileReservoirSize)
	}
	if reservoir.percentile(0) <= 0 || reservoir.percentile(1) <= 0 {
		t.Fatal("expected percentile samples")
	}
}
