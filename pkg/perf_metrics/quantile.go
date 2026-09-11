package perfmetrics

import "sort"

// quantileReservoir keeps a bounded, uniformly sampled set of observations.
// It prevents percentile memory from growing with request volume.
const quantileReservoirSize = 256
const quantileMinimumSamples = 10

type quantileReservoir struct {
	values []int64
	seen   uint64
}

func (r *quantileReservoir) add(value int64, random uint64) {
	if value <= 0 {
		return
	}
	r.seen++
	if len(r.values) < quantileReservoirSize {
		r.values = append(r.values, value)
		return
	}
	if random%r.seen < quantileReservoirSize {
		r.values[random%quantileReservoirSize] = value
	}
}

func (r *quantileReservoir) percentile(percent float64) int64 {
	if len(r.values) < quantileMinimumSamples {
		return 0
	}
	values := append([]int64(nil), r.values...)
	sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })
	if percent < 0 {
		percent = 0
	} else if percent > 1 {
		percent = 1
	}
	index := int(float64(len(values)-1) * percent)
	return values[index]
}
