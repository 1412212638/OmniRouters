package service

import (
	"time"

	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/pkg/billexport"
)

// Keep only the dimensions needed by the reports. The number of source log
// rows does not determine the size of the finished statement.
type billingUsageGroup struct {
	day                int64
	model              string
	memberID, keyID    int
	source, unit, kind string
}

type billingUsageAccumulator struct {
	indexes map[billingUsageGroup]int
	events  []billexport.Event
}

func (usage *billingUsageAccumulator) add(events []billexport.Event) error {
	_, timezoneOffset := time.Unix(0, 0).In(billexport.Timezone).Zone()
	for _, event := range events {
		day := (event.PostedAt + int64(timezoneOffset)) / 86400
		key := billingUsageGroup{day, event.Model, event.UserID, event.KeyID, event.Source, event.Unit, event.Kind}
		savings := event.Savings()
		index, exists := usage.indexes[key]
		if !exists {
			if len(usage.events) >= model.BillingExportRowLimit {
				return model.ErrBillingExportTooManyRows
			}
			usage.indexes[key] = len(usage.events)
			event.TaskID = ""
			event.Records = 1
			event.Original, event.RecordedSavings = nil, savings
			usage.events = append(usage.events, event)
			continue
		}
		group := &usage.events[index]
		if savings != nil {
			if group.RecordedSavings != nil {
				combined := group.RecordedSavings.Add(*savings)
				savings = &combined
			}
			group.RecordedSavings = savings
		}
		group.Amount = group.Amount.Add(event.Amount)
		group.Records++
		group.Requests += event.Requests
		group.InputTokens += event.InputTokens
		group.OutputTokens += event.OutputTokens
		group.CacheReadTokens += event.CacheReadTokens
		group.CacheWriteTokens += event.CacheWriteTokens
		group.LatestUsedAt = max(group.LatestUsedAt, group.UsedAt, event.UsedAt)
		if event.UsedAt > 0 && (group.UsedAt == 0 || group.UsedAt > event.UsedAt) {
			group.UsedAt = event.UsedAt
		}
		group.CompletedAt = max(group.CompletedAt, event.CompletedAt)
		if event.PostedAt >= group.PostedAt {
			group.PostedAt = event.PostedAt
			if event.KeyName != "" {
				group.KeyName = event.KeyName
			}
		}
	}
	return nil
}
