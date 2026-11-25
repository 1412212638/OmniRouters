// Package billexport renders the same normalized billing events at four levels.
package billexport

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

var Timezone = time.FixedZone("UTC+08:00", 8*60*60)

const (
	Payment            = "payment"
	Funding            = "funding"
	Adjustment         = "adjustment"
	Charge             = "charge"
	Precharge          = "precharge"
	AdditionalCharge   = "additional_charge"
	Refund             = "refund"
	UnconfirmedUsage   = "unconfirmed_usage"
	UnverifiedCurrency = "Unverified currency"
	UnitUSD            = "USD"
	UnitCNY            = "CNY"
	UnitCustomCurrency = "Custom currency"
	UnitQuota          = "Quota units"
	UnitTokens         = "Tokens"
)

type Event struct {
	PostedAt, UsedAt, CompletedAt                                int64
	LatestUsedAt                                                 int64
	Kind, Source, Unit                                           string
	Model, KeyName                                               string
	UserID, KeyID                                                int
	TaskID                                                       string
	Amount                                                       decimal.Decimal
	Original                                                     *decimal.Decimal
	RecordedSavings                                              *decimal.Decimal
	Requests                                                     int
	Records                                                      int
	InputTokens, OutputTokens, CacheReadTokens, CacheWriteTokens int64
}

func (event Event) RecordCount() int {
	return max(1, event.Records)
}

func (event Event) Savings() *decimal.Decimal {
	if event.RecordedSavings != nil {
		return event.RecordedSavings
	}
	if event.Original != nil && !event.Original.LessThan(event.Amount) {
		savings := event.Original.Sub(event.Amount)
		return &savings
	}
	return nil
}

type Statement struct {
	StartMonth, EndMonth, Account             string
	Language                                  string
	DisplayUnit, CurrencySymbol, ExchangeRate string
	Organization                              bool
	AsOf                                      time.Time
	Events                                    []Event
}

type cell struct {
	value   string
	numeric bool
}

func str(s string) cell             { return cell{value: s} }
func num(n int) cell                { return cell{value: strconv.Itoa(n), numeric: true} }
func amount(n decimal.Decimal) cell { return cell{value: n.String(), numeric: true} }
func optionalInt(n int) cell {
	if n == 0 {
		return str("")
	}
	return num(n)
}

func postedTime(timestamp int64) string {
	if timestamp <= 0 {
		return ""
	}
	return time.Unix(timestamp, 0).In(Timezone).Format("2006-01-02 15:04:05")
}

func postedDate(timestamp int64) string {
	date, _, _ := strings.Cut(postedTime(timestamp), " ")
	return date
}

type totals struct {
	payments, funding, adjustments, charges, refunds             decimal.Decimal
	requests                                                     int
	records                                                      int
	unverifiedPayments                                           int
	unconfirmedUsage                                             int
	inputTokens, outputTokens, cacheReadTokens, cacheWriteTokens int64
	firstUsage, lastUsage, lastUsageEnd, latestPosting           int64
	keyName                                                      string
	savings                                                      decimal.Decimal
	hasSavings                                                   bool
}

func (total *totals) add(event Event) {
	switch event.Kind {
	case Payment:
		if event.Unit == UnverifiedCurrency {
			total.unverifiedPayments += event.RecordCount()
		} else {
			total.payments = total.payments.Add(event.Amount)
		}
	case Funding:
		total.funding = total.funding.Add(event.Amount)
	case Adjustment:
		total.adjustments = total.adjustments.Add(event.Amount)
	case Refund:
		total.refunds = total.refunds.Sub(event.Amount)
	case UnconfirmedUsage:
		total.unconfirmedUsage += event.RecordCount()
	default:
		total.charges = total.charges.Add(event.Amount)
	}
	total.requests += event.Requests
	total.records += event.RecordCount()
	total.inputTokens += event.InputTokens
	total.outputTokens += event.OutputTokens
	total.cacheReadTokens += event.CacheReadTokens
	total.cacheWriteTokens += event.CacheWriteTokens
	if event.UsedAt > 0 && (total.firstUsage == 0 || event.UsedAt < total.firstUsage) {
		total.firstUsage = event.UsedAt
	}
	total.lastUsage = max(total.lastUsage, event.UsedAt, event.LatestUsedAt)
	total.lastUsageEnd = max(total.lastUsageEnd, event.CompletedAt)
	if event.PostedAt >= total.latestPosting {
		total.latestPosting = event.PostedAt
		if event.KeyName != "" {
			total.keyName = event.KeyName
		}
	}
	if savings := event.Savings(); savings != nil {
		total.savings = total.savings.Add(*savings)
		total.hasSavings = true
	}
}

func (total totals) net() decimal.Decimal { return total.charges.Sub(total.refunds) }

type grouping struct {
	first, second, source, unit string
}

func aggregate(events []Event, dimension func(Event) (string, string)) (map[grouping]*totals, []grouping) {
	groups := make(map[grouping]*totals)
	var keys []grouping
	for _, event := range events {
		first, second := "", ""
		if dimension != nil {
			first, second = dimension(event)
		}
		key := grouping{first, second, event.Source, event.Unit}
		if groups[key] == nil {
			groups[key] = &totals{}
			keys = append(keys, key)
		}
		groups[key].add(event)
	}
	sort.Slice(keys, func(i, j int) bool {
		a, b := keys[i], keys[j]
		if a.first != b.first {
			return a.first < b.first
		}
		if a.second != b.second {
			return a.second < b.second
		}
		if a.source != b.source {
			return a.source < b.source
		}
		return a.unit < b.unit
	})
	return groups, keys
}

// CSV text may be interpreted as a formula by a spreadsheet application.
func csvText(value string) string {
	if strings.HasPrefix(value, "\t") || strings.HasPrefix(value, "\r") || strings.HasPrefix(value, "\n") ||
		strings.HasPrefix(value, "\uFEFF") || strings.ContainsAny(firstRune(strings.TrimSpace(value)), "=+-@＝＋－＠") {
		return "'" + value
	}
	return value
}

func firstRune(value string) string {
	for _, char := range value {
		return string(char)
	}
	return ""
}

func writeCSV(out io.Writer, rows [][]cell) error {
	if _, err := io.WriteString(out, "\ufeff"); err != nil {
		return err
	}
	writer := csv.NewWriter(out)
	for _, row := range rows {
		values := make([]string, len(row))
		for i, column := range row {
			values[i] = column.value
			if !column.numeric {
				values[i] = csvText(column.value)
			}
		}
		if err := writer.Write(values); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}

var FileNames = map[string]string{"l0": "L0-statement.pdf", "l1": "L1-summary.xlsx", "l2": "L2-resources.csv", "l3": "L3-details.csv", "zip": "billing-statements.zip"}

func WriteArchive(statement Statement) ([]byte, error) {
	var out bytes.Buffer
	archive := zip.NewWriter(&out)
	files := []struct {
		level string
		write func(io.Writer) error
	}{
		{"l0", func(w io.Writer) error { return writePDF(w, statement) }},
		{"l1", func(w io.Writer) error { return writeWorkbook(w, statement) }},
		{"l2", func(w io.Writer) error { return writeCSV(w, resources(statement)) }},
		{"l3", func(w io.Writer) error { return writeCSV(w, details(statement)) }},
	}
	for _, file := range files {
		writer, err := archive.Create(FileNames[file.level])
		if err != nil {
			return nil, err
		}
		if err := file.write(writer); err != nil {
			return nil, err
		}
	}
	if err := archive.Close(); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
