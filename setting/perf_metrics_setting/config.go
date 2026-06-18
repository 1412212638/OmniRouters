package perf_metrics_setting

import (
	"github.com/QuantumNous/new-api/setting/config"
	"github.com/QuantumNous/new-api/setting/operation_setting"
)

type PerfMetricsSetting struct {
	Enabled             bool   `json:"enabled"`
	FlushInterval       int    `json:"flush_interval"`
	BucketTime          string `json:"bucket_time"`
	RetentionDays       int    `json:"retention_days"`
	ExcludedStatusCodes string `json:"excluded_status_codes"`
}

var perfMetricsSetting = PerfMetricsSetting{
	Enabled:             true,
	FlushInterval:       5,
	BucketTime:          "hour",
	RetentionDays:       0,
	ExcludedStatusCodes: "",
}

func init() {
	config.GlobalConfig.Register("perf_metrics_setting", &perfMetricsSetting)
}

func GetSetting() PerfMetricsSetting {
	return perfMetricsSetting
}

func GetBucketSeconds() int64 {
	switch perfMetricsSetting.BucketTime {
	case "minute":
		return 60
	case "5min":
		return 300
	case "hour":
		return 3600
	default:
		return 3600
	}
}

func GetFlushIntervalMinutes() int {
	if perfMetricsSetting.FlushInterval < 1 {
		return 1
	}
	return perfMetricsSetting.FlushInterval
}

func ShouldExcludeStatusCode(code int) bool {
	ranges, err := operation_setting.ParseHTTPStatusCodeRanges(perfMetricsSetting.ExcludedStatusCodes)
	if err != nil {
		return false
	}
	return operation_setting.MatchHTTPStatusCodeRanges(ranges, code)
}
