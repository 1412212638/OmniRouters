package perfmetrics

import (
	"testing"
	"time"

	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
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
