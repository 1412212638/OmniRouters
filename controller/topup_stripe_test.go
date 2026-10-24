package controller

import (
	"math"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/stretchr/testify/require"
)

func TestGetStripePayMoneyWithFeeOffsetsDiscountOnOriginalAmount(t *testing.T) {
	settings := operation_setting.GetPaymentSetting()
	originalStripeUnitPrice := setting.StripeUnitPrice
	originalDiscounts := settings.AmountDiscount
	originalDisplayType := operation_setting.GetGeneralSetting().QuotaDisplayType
	originalTopupGroupRatio := common.TopupGroupRatio2JSONString()
	t.Cleanup(func() {
		setting.StripeUnitPrice = originalStripeUnitPrice
		settings.AmountDiscount = originalDiscounts
		operation_setting.GetGeneralSetting().QuotaDisplayType = originalDisplayType
		require.NoError(t, common.UpdateTopupGroupRatioByJSONString(originalTopupGroupRatio))
	})

	setting.StripeUnitPrice = 8
	operation_setting.GetGeneralSetting().QuotaDisplayType = operation_setting.QuotaDisplayTypeUSD
	settings.AmountDiscount = map[int]float64{1000: 0.95}
	require.NoError(t, common.UpdateTopupGroupRatioByJSONString(`{"default":1}`))

	baseAmount := getStripePayMoney(1000, "default")
	payAmount := getStripePayMoneyWithFee(1000, "default", 0.05)
	require.InDelta(t, 7600, baseAmount, 0.000001)
	require.InDelta(t, 8000, payAmount, 0.000001)

	settings.AmountDiscount = map[int]float64{}
	payAmount = getStripePayMoneyWithFee(1000, "default", 0.05)
	require.InDelta(t, 8400, payAmount, 0.000001)
}

func TestStripeAmountInMinorUnits(t *testing.T) {
	tests := []struct {
		name     string
		amount   float64
		currency string
		want     int64
	}{
		{name: "two decimal currency", amount: 12.345, currency: "usd", want: 1235},
		{name: "zero decimal currency", amount: 123, currency: "JPY", want: 123},
		{name: "three decimal currency", amount: 1.2345, currency: "KWD", want: 1235},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := stripeAmountInMinorUnits(tt.amount, tt.currency)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}

	for _, amount := range []float64{0, -1, math.NaN(), math.Inf(1)} {
		_, err := stripeAmountInMinorUnits(amount, "usd")
		require.Error(t, err)
	}
}

func TestStripeAmountInPriceCurrency(t *testing.T) {
	originalExchangeRate := operation_setting.USDExchangeRate
	t.Cleanup(func() {
		operation_setting.USDExchangeRate = originalExchangeRate
	})
	operation_setting.USDExchangeRate = 7.2

	amount, err := stripeAmountInPriceCurrency(7416, "usd")
	require.NoError(t, err)
	require.InDelta(t, 1030, amount, 0.000001)

	amount, err = stripeAmountInPriceCurrency(7416, "cny")
	require.NoError(t, err)
	require.Equal(t, float64(7416), amount)

	operation_setting.USDExchangeRate = 0
	_, err = stripeAmountInPriceCurrency(7416, "usd")
	require.Error(t, err)
}
