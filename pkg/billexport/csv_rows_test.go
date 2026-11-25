package billexport

import (
	"bytes"
	"encoding/csv"
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestCSVReconcilesResourceUsageByPostingMonthAndUnit(t *testing.T) {
	prechargedAt := time.Date(2026, 11, 30, 23, 50, 0, 0, Timezone).Unix()
	settledAt := prechargedAt + int64((15 * time.Minute).Seconds())
	amountCharged := decimal.RequireFromString("2.5")
	walletUnit := "USD"
	const model = "=unsafe-model"
	statement := Statement{
		Account:    "Personal account #1",
		StartMonth: "2026-11",
		EndMonth:   "2026-12",
		Events: []Event{
			{Kind: Precharge, Source: "wallet", Unit: walletUnit, Model: model, KeyID: 11, KeyName: "=unsafe-key",
				PostedAt: prechargedAt, UsedAt: prechargedAt, Amount: amountCharged, Requests: 1},
			{Kind: Refund, Source: "wallet", Unit: walletUnit, Model: model, KeyID: 11, KeyName: "=unsafe-key",
				PostedAt: settledAt, UsedAt: prechargedAt, CompletedAt: settledAt, Amount: amountCharged.Neg()},
			{Kind: UnconfirmedUsage, Source: "wallet", Unit: walletUnit, Model: model, KeyID: 11, PostedAt: settledAt,
				Amount: decimal.NewFromInt(9), Requests: 4, Records: 4},
			{Kind: Charge, Source: "subscription", Unit: "Tokens", Model: model, KeyID: 11, PostedAt: settledAt,
				Amount: decimal.NewFromInt(42), Requests: 1},
			{Kind: Funding, Source: "wallet_paid", Unit: walletUnit, PostedAt: settledAt,
				Amount: decimal.NewFromInt(10)},
			{Kind: Payment, Source: "wallet_topup", Unit: UnverifiedCurrency, PostedAt: settledAt,
				Amount: decimal.NewFromInt(75)},
		},
	}

	resourceRows := readExportCSV(t, resources(statement))
	require.Len(t, resourceRows, 5)
	require.Equal(t, "2026-11", resourceRows[0]["Posting month"])
	require.Equal(t, "'=unsafe-model", resourceRows[0]["Model"])
	require.Equal(t, "'=unsafe-key", resourceRows[0]["API key name"])
	require.Equal(t, "2.5", resourceRows[0]["Net usage"])
	require.Equal(t, "42", resourceRows[1]["Net usage"])
	require.Equal(t, "Tokens", resourceRows[1]["Unit"])
	require.Equal(t, "2026-12", resourceRows[2]["Posting month"])
	require.Equal(t, "-2.5", resourceRows[2]["Net usage"])
	require.Equal(t, "4", resourceRows[2]["Unconfirmed usage (count)"])
	require.Equal(t, "2026-12-01 00:05:00", resourceRows[2]["Last usage end (UTC+08:00)"])
	require.Equal(t, "", resourceRows[3]["Net usage"])
	require.Equal(t, "10", resourceRows[3]["Funding"])
	require.Equal(t, "1", resourceRows[4]["Unverified payments (count)"])
	require.Empty(t, resourceRows[4]["Recorded payments"])

	original := decimal.NewFromInt(6)
	statement.Events = append(statement.Events,
		Event{Kind: Charge, Source: "wallet", Unit: walletUnit, Model: model, KeyID: 99, KeyName: "key-private",
			PostedAt: prechargedAt, Amount: decimal.NewFromInt(4), Original: &original, InputTokens: 12, Requests: 1},
		Event{Kind: Charge, Source: "wallet_tool_fee", Unit: walletUnit, Model: model, KeyID: 99,
			PostedAt: prechargedAt, Amount: decimal.RequireFromString("0.25")},
		Event{Kind: AdditionalCharge, Source: "wallet", Unit: walletUnit, Model: model, KeyID: 11,
			PostedAt: settledAt, Amount: decimal.RequireFromString("0.75")},
		Event{Kind: Charge, Source: "wallet", Unit: walletUnit, PostedAt: settledAt,
			Amount: decimal.NewFromInt(3), Requests: 1},
	)
	detailRows := readExportCSV(t, details(statement))
	require.Len(t, detailRows, 3)
	require.Equal(t, "2026-11 - 2026-12", detailRows[0]["Billing period"])
	require.Equal(t, "Unrecorded model", detailRows[0]["Model"])
	require.Equal(t, "3", detailRows[0]["Net usage"])
	require.Equal(t, "'=unsafe-model", detailRows[1]["Model"])
	require.Equal(t, "Tokens", detailRows[1]["Unit"])
	require.Equal(t, "42", detailRows[1]["Net usage"])
	wallet := detailRows[2]
	require.Equal(t, "'=unsafe-model", wallet["Model"])
	require.Equal(t, walletUnit, wallet["Unit"])
	require.Equal(t, "5", wallet["Net usage"])
	require.Equal(t, "7.5", wallet["Charges"])
	require.Equal(t, "2.5", wallet["Refunds"])
	require.Equal(t, "2.5", wallet["Task precharge"])
	require.Equal(t, "0.75", wallet["Additional charge"])
	require.Equal(t, "2", wallet["Recorded savings"])
	require.Equal(t, "4", wallet["Unconfirmed usage (count)"])
	require.Equal(t, "12", wallet["Input tokens"])
	require.Equal(t, "6", wallet["Requests"])
	require.NotContains(t, wallet, "Request ID")
	require.NotContains(t, wallet, "Task ID")
	require.NotContains(t, wallet, "API key ID")
	require.NotContains(t, wallet, "Funding")
	require.NotContains(t, wallet, "Recorded payments")
	statement.Language = "zh"
	chineseDetails := readExportCSV(t, details(statement))
	require.Equal(t, "未记录模型", chineseDetails[0]["模型"])
	require.Equal(t, "5", chineseDetails[2]["净消耗"])

	byModel := workbookBreakdown(statement, "By model", "Model", func(event Event) string { return event.Model })
	require.Len(t, byModel.tables[0].rows, 5)
	require.Equal(t, "未记录模型", byModel.tables[0].rows[1][0].value)
	require.Equal(t, "3", byModel.tables[0].rows[1][3].value)
}

func readExportCSV(t *testing.T, rows [][]cell) []map[string]string {
	t.Helper()
	var output bytes.Buffer
	require.NoError(t, writeCSV(&output, rows))
	require.True(t, strings.HasPrefix(output.String(), "\ufeff"))
	reader := csv.NewReader(strings.NewReader(strings.TrimPrefix(output.String(), "\ufeff")))
	records, err := reader.ReadAll()
	require.NoError(t, err)
	decoded := make([]map[string]string, 0, len(records)-1)
	for _, record := range records[1:] {
		row := make(map[string]string, len(records[0]))
		for column, header := range records[0] {
			row[header] = record[column]
		}
		decoded = append(decoded, row)
	}
	return decoded
}
