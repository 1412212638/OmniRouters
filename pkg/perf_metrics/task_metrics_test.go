package perfmetrics

import (
	"testing"
	"time"

	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/setting/config"
	"github.com/QuantumNous/new-api/setting/perf_metrics_setting"
)

func clearTaskMetricBuckets() {
	hotBuckets.Range(func(key, _ any) bool {
		hotBuckets.Delete(key)
		return true
	})
}

func TestRecordTaskResultSamplesTerminalTask(t *testing.T) {
	clearTaskMetricBuckets()
	t.Cleanup(clearTaskMetricBuckets)

	now := time.Now().Unix()
	RecordTaskResult(&model.Task{
		Status:     model.TaskStatusSuccess,
		Group:      "video",
		SubmitTime: now - 120,
		StartTime:  now - 100,
		FinishTime: now,
		Properties: model.Properties{OriginModelName: "video-model"},
	}, &relaycommon.TaskInfo{TotalTokens: 5000})
	RecordTaskResult(&model.Task{
		Status:     model.TaskStatusFailure,
		Group:      "video",
		SubmitTime: now - 60,
		FinishTime: now,
		Properties: model.Properties{OriginModelName: "video-model"},
	}, relaycommon.FailTaskInfo("upstream failed"))

	merged := map[bucketKey]counters{}
	hotBuckets.Range(func(key, value any) bool {
		k := key.(bucketKey)
		mergeCounters(merged, bucketKey{model: k.model, group: k.group, bucketTs: 0}, value.(*atomicBucket).snapshot())
		return true
	})
	got := merged[bucketKey{model: "video-model", group: "video", bucketTs: 0}]
	if got.requestCount != 2 || got.successCount != 1 {
		t.Fatalf("request/success counts = %d/%d, want 2/1", got.requestCount, got.successCount)
	}
	if got.totalLatencyMs != 180000 || got.outputTokens != 5000 || got.generationMs != 100000 {
		t.Fatalf("task timing/token counters = %+v", got)
	}
}

func TestRecordTaskResultUsesBillingModelAndSkipsUnknownModel(t *testing.T) {
	clearTaskMetricBuckets()
	t.Cleanup(clearTaskMetricBuckets)
	now := time.Now().Unix()
	RecordTaskResult(&model.Task{
		Status:     model.TaskStatusSuccess,
		Group:      "default",
		SubmitTime: now - 10,
		FinishTime: now,
		Properties: model.Properties{OriginModelName: "properties-model"},
		PrivateData: model.TaskPrivateData{BillingContext: &model.TaskBillingContext{
			OriginModelName: "billing-model",
		}},
	}, nil)
	RecordTaskResult(&model.Task{Status: model.TaskStatusSuccess, FinishTime: now}, nil)

	var seen int
	hotBuckets.Range(func(key, _ any) bool {
		seen++
		if key.(bucketKey).model != "billing-model" {
			t.Fatalf("recorded unexpected model %q", key.(bucketKey).model)
		}
		return true
	})
	if seen != 1 {
		t.Fatalf("recorded buckets = %d, want 1", seen)
	}
}

func TestRecordTaskResultHonorsExcludedFailureStatuses(t *testing.T) {
	setting := config.GlobalConfig.Get("perf_metrics_setting").(*perf_metrics_setting.PerfMetricsSetting)
	original := *setting
	setting.Enabled = true
	setting.ExcludedStatusCodes = "400-404,502,504"
	t.Cleanup(func() { *setting = original })
	t.Cleanup(clearTaskMetricBuckets)

	for _, tc := range []struct {
		name      string
		status    model.TaskStatus
		reason    string
		result    *relaycommon.TaskInfo
		wantCount int64
	}{
		{"embedded upstream 400", model.TaskStatusFailure, "", relaycommon.FailTaskInfo(`seedance HTTP 400 https://ark.cn-beijing.volces.com/api/v3/contents/generations/tasks: {"error":{"code":"InvalidParameter","message":"reference_audio cannot be the only reference input"}}`), 0},
		{"persisted failure reason", model.TaskStatusFailure, "seedance HTTP 400: invalid content", nil, 0},
		{"excluded range", model.TaskStatusFailure, "upstream HTTP 403: forbidden", nil, 0},
		{"polling failure", model.TaskStatusFailure, "poll failed: auth (HTTP 401)", nil, 0},
		{"excluded server failure", model.TaskStatusFailure, "HTTP/1.1 502 Bad Gateway", nil, 0},
		{"included server failure", model.TaskStatusFailure, "upstream HTTP 500: internal error", nil, 1},
		{"unknown failure", model.TaskStatusFailure, "video generation failed", nil, 1},
		{"business code is not HTTP", model.TaskStatusFailure, "provider code 400: processing failed", nil, 1},
		{"success is never excluded", model.TaskStatusSuccess, "HTTP 400", nil, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			clearTaskMetricBuckets()
			RecordTaskResult(&model.Task{
				Status: tc.status, FailReason: tc.reason,
				Properties: model.Properties{OriginModelName: "video-model"},
			}, tc.result)
			var count int64
			hotBuckets.Range(func(_, value any) bool {
				count += value.(*atomicBucket).snapshot().requestCount
				return true
			})
			if count != tc.wantCount {
				t.Fatalf("recorded request count = %d, want %d", count, tc.wantCount)
			}
		})
	}
}
