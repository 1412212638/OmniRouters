export type PerformanceSeriesPoint = {
  ts: number
  avg_ttft_ms: number
  avg_latency_ms: number
  success_rate: number
  avg_tps: number
  avg_tpot_ms: number
  cache_rate?: number
  ttft_p10_ms?: number
  ttft_p50_ms?: number
  ttft_p95_ms?: number
  ttft_p99_ms?: number
  tpot_p10_ms?: number
  tpot_p50_ms?: number
  tpot_p95_ms?: number
  tpot_p99_ms?: number
}

export type PerformanceGroup = {
  group: string
  avg_ttft_ms: number
  avg_latency_ms: number
  success_rate: number
  avg_tps: number
  avg_tpot_ms: number
  cache_rate?: number
  ttft_p10_ms?: number
  ttft_p50_ms?: number
  ttft_p95_ms?: number
  ttft_p99_ms?: number
  tpot_p10_ms?: number
  tpot_p50_ms?: number
  tpot_p95_ms?: number
  tpot_p99_ms?: number
  series: PerformanceSeriesPoint[]
}

export type PerformanceMetricsData = {
  success: boolean
  message?: string
  data: {
    model_name: string
    series_schema?: string
    groups: PerformanceGroup[]
  }
}

export type SuccessRatePoint = { ts: number; success_rate: number }

export type PerfModelSummary = {
  cache_rate?: number
  model_name: string
  avg_latency_ms: number
  success_rate: number
  avg_tps: number
  recent_success_series?: SuccessRatePoint[]
  request_count?: number
}

export type PerfSummaryAllData = {
  success: boolean
  message?: string
  data: {
    models: PerfModelSummary[]
  }
}
