package billexport

import (
	"cmp"
	"slices"
	"time"

	"github.com/shopspring/decimal"
)

type billingCategory uint8

const (
	usageCategory billingCategory = iota
	fundingCategory
	adjustmentCategory
	paymentCategory
)

func eventCategory(kind string) billingCategory {
	switch kind {
	case Funding:
		return fundingCategory
	case Adjustment:
		return adjustmentCategory
	case Payment:
		return paymentCategory
	default:
		return usageCategory
	}
}

func (category billingCategory) label(statement Statement) string {
	return statement.label([...]string{"Usage", "Funding", "Adjustment", "Payment"}[category])
}

func postedMonth(timestamp int64) string {
	if timestamp <= 0 {
		return ""
	}
	return time.Unix(timestamp, 0).In(Timezone).Format("2006-01")
}

type resourceScope struct {
	month    string
	category billingCategory
}

func resources(statement Statement) [][]cell {
	rows := [][]cell{statement.header(
		"Posting month", "Account", "Record category", "Billing source", "Model", "API key ID", "API key name", "Unit",
		"Charges", "Refunds", "Net usage", "Recorded savings", "Funding", "Adjustments", "Recorded payments", "Unverified payments (count)",
		"Requests", "Unconfirmed usage (count)", "Input tokens", "Output tokens", "Cache read tokens", "Cache write tokens",
		"First usage (UTC+08:00)", "Last usage (UTC+08:00)", "Last usage end (UTC+08:00)", "Latest posting (UTC+08:00)",
	)}
	byScope := make(map[resourceScope][]Event)
	for _, event := range statement.Events {
		scope := resourceScope{postedMonth(event.PostedAt), eventCategory(event.Kind)}
		byScope[scope] = append(byScope[scope], event)
	}
	scopes := make([]resourceScope, 0, len(byScope))
	for scope := range byScope {
		scopes = append(scopes, scope)
	}
	slices.SortFunc(scopes, func(a, b resourceScope) int {
		if a.month != b.month {
			return cmp.Compare(a.month, b.month)
		}
		return cmp.Compare(a.category, b.category)
	})
	for _, scope := range scopes {
		groups, keys := aggregate(byScope[scope], func(event Event) (string, string) {
			return event.Model, optionalInt(event.KeyID).value
		})
		for _, key := range keys {
			total := groups[key]
			amounts := resourceAmounts(scope.category, key.unit, *total)
			rows = append(rows, []cell{
				str(scope.month), str(statement.Account), str(scope.category.label(statement)), str(statement.codeLabel(key.source)),
				str(key.first), str(key.second), str(total.keyName), str(statement.unitLabel(key.unit)),
				amounts.charges, amounts.refunds, amounts.net, amounts.savings, amounts.funding, amounts.adjustments, amounts.payments, amounts.unverifiedPayments,
				amounts.requests, amounts.unconfirmed, amounts.input, amounts.output, amounts.read, amounts.write,
				str(postedTime(total.firstUsage)), str(postedTime(total.lastUsage)), str(postedTime(total.lastUsageEnd)), str(postedTime(total.latestPosting)),
			})
		}
	}
	return rows
}

type resourceCells struct {
	charges, refunds, net, savings, funding, adjustments, payments, unverifiedPayments cell
	requests, unconfirmed, input, output, read, write                                  cell
}

func resourceAmounts(category billingCategory, unit string, total totals) resourceCells {
	var cells resourceCells
	switch category {
	case usageCategory:
		cells.charges, cells.refunds, cells.net = amount(total.charges), amount(total.refunds), amount(total.net())
		if total.hasSavings {
			cells.savings = amount(total.savings)
		}
		cells.requests = num(total.requests)
		cells.unconfirmed = num(total.unconfirmedUsage)
		cells.input = optionalTokenCount(total.inputTokens)
		cells.output = optionalTokenCount(total.outputTokens)
		cells.read = optionalTokenCount(total.cacheReadTokens)
		cells.write = optionalTokenCount(total.cacheWriteTokens)
	case fundingCategory:
		cells.funding = amount(total.funding)
	case adjustmentCategory:
		cells.adjustments = amount(total.adjustments)
	case paymentCategory:
		if unit == UnverifiedCurrency {
			cells.unverifiedPayments = num(total.unverifiedPayments)
		} else {
			cells.payments = amount(total.payments)
		}
	}
	return cells
}

func optionalTokenCount(count int64) cell {
	if count == 0 {
		return str("")
	}
	return amount(decimal.NewFromInt(count))
}

type detailTotals struct {
	totals
	precharges, supplementalCharges decimal.Decimal
}

func details(statement Statement) [][]cell {
	rows := [][]cell{statement.header(
		"Billing period", "Account", "Model", "Unit",
		"Requests", "Charges", "Task precharge", "Additional charge", "Refunds", "Net usage",
		"Recorded savings", "Unconfirmed usage (count)",
		"Input tokens", "Output tokens", "Cache read tokens", "Cache write tokens",
	)}
	period := statement.StartMonth
	if statement.EndMonth != statement.StartMonth {
		period += " - " + statement.EndMonth
	}
	type modelUnit struct {
		model, unit string
	}
	groups := make(map[modelUnit]*detailTotals)
	for _, event := range statement.Events {
		if eventCategory(event.Kind) != usageCategory {
			continue
		}
		key := modelUnit{event.Model, event.Unit}
		total := groups[key]
		if total == nil {
			total = &detailTotals{}
			groups[key] = total
		}
		total.add(event)
		// The charge total includes the precharge and the positive settlement
		// delta. Snapshot values from settlement rows would count them again.
		switch event.Kind {
		case Precharge:
			total.precharges = total.precharges.Add(event.Amount)
		case AdditionalCharge:
			total.supplementalCharges = total.supplementalCharges.Add(event.Amount)
		}
	}
	keys := make([]modelUnit, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	slices.SortFunc(keys, func(a, b modelUnit) int {
		if a.model != b.model {
			return cmp.Compare(a.model, b.model)
		}
		return cmp.Compare(a.unit, b.unit)
	})
	for _, key := range keys {
		total := groups[key]
		amounts := resourceAmounts(usageCategory, key.unit, total.totals)
		rows = append(rows, []cell{
			str(period), str(statement.Account), str(statement.billedModelName(key.model)), str(statement.unitLabel(key.unit)),
			amounts.requests, amounts.charges, amount(total.precharges), amount(total.supplementalCharges), amounts.refunds, amounts.net,
			amounts.savings, amounts.unconfirmed,
			amounts.input, amounts.output, amounts.read, amounts.write,
		})
	}
	return rows
}
