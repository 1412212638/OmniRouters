package service

import (
	"math"
	"strings"
	"unicode"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/pkg/billexport"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/shopspring/decimal"
)

func billingExportDisplaySnapshot() model.BillingExportDisplay {
	settings := *operation_setting.GetGeneralSetting()
	display := model.BillingExportDisplay{
		Mode: settings.QuotaDisplayType, Rate: 1,
		QuotaPerUnit: validDisplayNumber(common.QuotaPerUnit, 500_000),
	}
	switch display.Mode {
	case operation_setting.QuotaDisplayTypeCNY:
		display.Rate = validDisplayNumber(operation_setting.USDExchangeRate, 1)
	case operation_setting.QuotaDisplayTypeCustom:
		display.Rate = validDisplayNumber(settings.CustomCurrencyExchangeRate, 1)
		symbol := []rune(strings.TrimSpace(strings.Map(func(char rune) rune {
			if unicode.IsControl(char) {
				return -1
			}
			return char
		}, settings.CustomCurrencySymbol)))
		display.Symbol = string(symbol[:min(8, len(symbol))])
		if display.Symbol == "" {
			display.Symbol = "¤"
		}
	case operation_setting.QuotaDisplayTypeTokens, operation_setting.QuotaDisplayTypeUSD:
	default:
		display.Mode = operation_setting.QuotaDisplayTypeUSD
	}
	return display
}

func validDisplayNumber(value, fallback float64) float64 {
	if value > 0 && !math.IsInf(value, 0) {
		return value
	}
	return fallback
}

type billingQuotaDisplay struct {
	unit, symbol  string
	rate, perUnit decimal.Decimal
}

func resolveBillingQuotaDisplay(snapshot model.BillingExportDisplay) billingQuotaDisplay {
	if snapshot.Mode == "" {
		// A job queued before the snapshot columns existed uses the settings
		// available when it first runs. Completed archives are immutable.
		snapshot = billingExportDisplaySnapshot()
	}
	display := billingQuotaDisplay{
		unit: billexport.UnitUSD, symbol: snapshot.Symbol,
		rate:    decimal.NewFromFloat(validDisplayNumber(snapshot.Rate, 1)),
		perUnit: decimal.NewFromFloat(validDisplayNumber(snapshot.QuotaPerUnit, 500_000)),
	}
	switch snapshot.Mode {
	case operation_setting.QuotaDisplayTypeCNY:
		display.unit = billexport.UnitCNY
	case operation_setting.QuotaDisplayTypeCustom:
		display.unit = billexport.UnitCustomCurrency
	case operation_setting.QuotaDisplayTypeTokens:
		display.unit = billexport.UnitQuota
	}
	return display
}

func (display billingQuotaDisplay) amount(raw int64) decimal.Decimal {
	value := decimal.NewFromInt(raw)
	if display.unit == billexport.UnitQuota {
		return value
	}
	return value.Mul(display.rate).Div(display.perUnit)
}
