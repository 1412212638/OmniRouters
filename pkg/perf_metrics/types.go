package perfmetrics

import (
	"sync"
	"sync/atomic"
)

type Store interface {
	Record(sample Sample)
	Query(params QueryParams) (QueryResult, error)
}

type Sample struct {
	Model        string
	Group        string
	LatencyMs    int64
	TtftMs       int64
	HasTtft      bool
	Success      bool
	OutputTokens int64
	GenerationMs int64
	InputTokens  int64
	CachedTokens int64
}

type QueryParams struct {
	Model string
	Group string
	Hours int
}

type BucketPoint struct {
	Ts           int64   `json:"ts"`
	AvgTtftMs    int64   `json:"avg_ttft_ms"`
	AvgLatencyMs int64   `json:"avg_latency_ms"`
	SuccessRate  float64 `json:"success_rate"`
	AvgTps       float64 `json:"avg_tps"`
	AvgTpotMs    int64   `json:"avg_tpot_ms"`
	CacheRate    *float64 `json:"cache_rate,omitempty"`
}

type GroupResult struct {
	Group        string        `json:"group"`
	AvgTtftMs    int64         `json:"avg_ttft_ms"`
	AvgLatencyMs int64         `json:"avg_latency_ms"`
	SuccessRate  float64       `json:"success_rate"`
	AvgTps       float64       `json:"avg_tps"`
	AvgTpotMs    int64         `json:"avg_tpot_ms"`
	CacheRate    *float64      `json:"cache_rate,omitempty"`
	TtftP95Ms    int64         `json:"ttft_p95_ms,omitempty"`
	TtftP99Ms    int64         `json:"ttft_p99_ms,omitempty"`
	TpotP95Ms    int64         `json:"tpot_p95_ms,omitempty"`
	TpotP99Ms    int64         `json:"tpot_p99_ms,omitempty"`
	Series       []BucketPoint `json:"series"`
}

type QueryResult struct {
	ModelName    string        `json:"model_name"`
	SeriesSchema string        `json:"series_schema"`
	Groups       []GroupResult `json:"groups"`
}

type SuccessRatePoint struct {
	Ts          int64   `json:"ts"`
	SuccessRate float64 `json:"success_rate"`
}

type ModelSummary struct {
	ModelName           string             `json:"model_name"`
	AvgLatencyMs        int64              `json:"avg_latency_ms"`
	SuccessRate         float64            `json:"success_rate"`
	AvgTps              float64            `json:"avg_tps"`
	RecentSuccessSeries []SuccessRatePoint `json:"recent_success_series,omitempty"`
	RequestCount        int64              `json:"-"`
}

type SummaryAllResult struct {
	Models []ModelSummary `json:"models"`
}

type bucketKey struct {
	model    string
	group    string
	bucketTs int64
}

type counters struct {
	requestCount   int64
	successCount   int64
	totalLatencyMs int64
	ttftSumMs      int64
	ttftCount      int64
	outputTokens   int64
	generationMs   int64
	inputTokens    int64
	cachedTokens   int64
	ttftQuantiles  quantileReservoir
	tpotQuantiles  quantileReservoir
}

type atomicBucket struct {
	mu sync.Mutex
	ttftQuantiles quantileReservoir
	tpotQuantiles quantileReservoir
	requestCount   atomic.Int64
	successCount   atomic.Int64
	totalLatencyMs atomic.Int64
	ttftSumMs      atomic.Int64
	ttftCount      atomic.Int64
	outputTokens   atomic.Int64
	generationMs   atomic.Int64
	inputTokens    atomic.Int64
	cachedTokens   atomic.Int64
}

func (b *atomicBucket) add(sample Sample) {
	b.requestCount.Add(1)
	if sample.Success {
		b.successCount.Add(1)
	}
	if sample.LatencyMs > 0 {
		b.totalLatencyMs.Add(sample.LatencyMs)
	}
	if sample.HasTtft && sample.TtftMs >= 0 {
		b.ttftSumMs.Add(sample.TtftMs)
		b.ttftCount.Add(1)
		b.mu.Lock()
		b.ttftQuantiles.add(sample.TtftMs, uint64(sample.TtftMs)^b.ttftQuantiles.seen)
		b.mu.Unlock()
	}
	if sample.OutputTokens > 0 && sample.GenerationMs > 0 {
		b.outputTokens.Add(sample.OutputTokens)
		b.generationMs.Add(sample.GenerationMs)
		b.mu.Lock()
		b.tpotQuantiles.add(sample.GenerationMs*1000/sample.OutputTokens, uint64(sample.OutputTokens)^b.tpotQuantiles.seen)
		b.mu.Unlock()
	}
	if sample.CachedTokens > 0 {
		b.inputTokens.Add(sample.InputTokens)
		b.cachedTokens.Add(sample.CachedTokens)
	}
}

func (b *atomicBucket) quantiles() (int64, int64, int64, int64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.ttftQuantiles.percentile(.95), b.ttftQuantiles.percentile(.99), b.tpotQuantiles.percentile(.95), b.tpotQuantiles.percentile(.99)
}

func (b *atomicBucket) snapshot() counters {
	b.mu.Lock()
	ttftQuantiles := quantileReservoir{values: append([]int64(nil), b.ttftQuantiles.values...), seen: b.ttftQuantiles.seen}
	tpotQuantiles := quantileReservoir{values: append([]int64(nil), b.tpotQuantiles.values...), seen: b.tpotQuantiles.seen}
	b.mu.Unlock()
	return counters{
		requestCount:   b.requestCount.Load(),
		successCount:   b.successCount.Load(),
		totalLatencyMs: b.totalLatencyMs.Load(),
		ttftSumMs:      b.ttftSumMs.Load(),
		ttftCount:      b.ttftCount.Load(),
		outputTokens:   b.outputTokens.Load(),
		generationMs:   b.generationMs.Load(),
		inputTokens: b.inputTokens.Load(), cachedTokens: b.cachedTokens.Load(),
		ttftQuantiles: ttftQuantiles, tpotQuantiles: tpotQuantiles,
	}
}

func (b *atomicBucket) drain() counters {
	b.mu.Lock()
	ttftQuantiles := quantileReservoir{values: append([]int64(nil), b.ttftQuantiles.values...), seen: b.ttftQuantiles.seen}
	tpotQuantiles := quantileReservoir{values: append([]int64(nil), b.tpotQuantiles.values...), seen: b.tpotQuantiles.seen}
	b.ttftQuantiles, b.tpotQuantiles = quantileReservoir{}, quantileReservoir{}
	b.mu.Unlock()
	return counters{
		requestCount:   b.requestCount.Swap(0),
		successCount:   b.successCount.Swap(0),
		totalLatencyMs: b.totalLatencyMs.Swap(0),
		ttftSumMs:      b.ttftSumMs.Swap(0),
		ttftCount:      b.ttftCount.Swap(0),
		outputTokens:   b.outputTokens.Swap(0),
		generationMs:   b.generationMs.Swap(0),
		inputTokens: b.inputTokens.Swap(0), cachedTokens: b.cachedTokens.Swap(0),
		ttftQuantiles: ttftQuantiles, tpotQuantiles: tpotQuantiles,
	}
}

func (b *atomicBucket) addCounters(c counters) {
	if c.requestCount != 0 {
		b.requestCount.Add(c.requestCount)
	}
	if c.successCount != 0 {
		b.successCount.Add(c.successCount)
	}
	if c.totalLatencyMs != 0 {
		b.totalLatencyMs.Add(c.totalLatencyMs)
	}
	if c.ttftSumMs != 0 {
		b.ttftSumMs.Add(c.ttftSumMs)
	}
	if c.ttftCount != 0 {
		b.ttftCount.Add(c.ttftCount)
	}
	if c.outputTokens != 0 {
		b.outputTokens.Add(c.outputTokens)
	}
	if c.generationMs != 0 {
		b.generationMs.Add(c.generationMs)
	}
	if c.inputTokens != 0 { b.inputTokens.Add(c.inputTokens) }
	if c.cachedTokens != 0 { b.cachedTokens.Add(c.cachedTokens) }
	b.mu.Lock()
	mergeReservoir(&b.ttftQuantiles, c.ttftQuantiles)
	mergeReservoir(&b.tpotQuantiles, c.tpotQuantiles)
	b.mu.Unlock()
}
