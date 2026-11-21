package controller

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/stretchr/testify/require"
)

func TestRechargeThresholdDiscountAcrossPaymentMethods(t *testing.T) {
	payment := operation_setting.GetPaymentSetting()
	discounts, byThreshold := payment.AmountDiscount, payment.AmountDiscountByThreshold
	price, stripePrice, waffoPrice := operation_setting.Price, setting.StripeUnitPrice, setting.WaffoUnitPrice
	display := operation_setting.GetGeneralSetting()
	displayType := display.QuotaDisplayType
	groupRatios := common.TopupGroupRatio2JSONString()
	t.Cleanup(func() {
		payment.AmountDiscount, payment.AmountDiscountByThreshold = discounts, byThreshold
		operation_setting.Price, setting.StripeUnitPrice, setting.WaffoUnitPrice = price, stripePrice, waffoPrice
		display.QuotaDisplayType = displayType
		require.NoError(t, common.UpdateTopupGroupRatioByJSONString(groupRatios))
	})

	payment.AmountDiscount = map[int]float64{1000: 0.95, 100: 0.98}
	payment.AmountDiscountByThreshold = false
	operation_setting.Price, setting.StripeUnitPrice, setting.WaffoUnitPrice = 1, 1, 1
	display.QuotaDisplayType = operation_setting.QuotaDisplayTypeUSD
	require.NoError(t, common.UpdateTopupGroupRatioByJSONString(`{"default":1}`))
	require.Equal(t, 500.0, getPayMoney(500, "default"))
	require.Equal(t, 98.0, getPayMoney(100, "default"))

	payment.AmountDiscountByThreshold = true
	for _, tc := range []struct {
		amount int64
		rate   float64
	}{
		{99, 1},
		{100, 0.98},
		{500, 0.98},
		{800, 0.98},
		{999, 0.98},
		{1000, 0.95},
		{1500, 0.95},
		{2000, 0.95},
	} {
		rate := operation_setting.GetRechargeDiscountRate(tc.amount)
		require.Equal(t, tc.rate, rate, "amount %d", tc.amount)
		base := float64(tc.amount) * tc.rate
		total := base + float64(tc.amount)*0.05
		require.InDelta(t, base, getPayMoney(tc.amount, "default"), 1e-8)
		require.InDelta(t, total, getPayMoneyWithFee(tc.amount, "default", 0.05), 1e-8)
		require.InDelta(t, base, getStripePayMoney(tc.amount, "default"), 1e-8)
		require.InDelta(t, total, getStripePayMoneyWithFee(tc.amount, "default", 0.05), 1e-8)
		require.InDelta(t, base, getWaffoPayMoney(tc.amount, "default", rate), 1e-8)
		pancakeBase, _ := getWaffoPancakePriceQuote(tc.amount, "default")
		require.InDelta(t, base, pancakeBase.InexactFloat64(), 1e-8)
	}
}

func TestGetPayMoneyWithFeeOffsetsDiscountOnOriginalAmount(t *testing.T) {
	settings := operation_setting.GetPaymentSetting()
	originalPrice := operation_setting.Price
	originalDiscounts := settings.AmountDiscount
	originalFeeRate := settings.FeeRate
	originalPayMethods := operation_setting.PayMethods
	originalDisplayType := operation_setting.GetGeneralSetting().QuotaDisplayType
	originalTopupGroupRatio := common.TopupGroupRatio2JSONString()
	t.Cleanup(func() {
		operation_setting.Price = originalPrice
		settings.AmountDiscount = originalDiscounts
		settings.FeeRate = originalFeeRate
		operation_setting.PayMethods = originalPayMethods
		operation_setting.GetGeneralSetting().QuotaDisplayType = originalDisplayType
		require.NoError(t, common.UpdateTopupGroupRatioByJSONString(originalTopupGroupRatio))
	})

	operation_setting.Price = 1
	operation_setting.GetGeneralSetting().QuotaDisplayType = operation_setting.QuotaDisplayTypeUSD
	settings.AmountDiscount = map[int]float64{1000: 0.95}
	require.NoError(t, common.UpdateTopupGroupRatioByJSONString(`{"default":1}`))

	baseAmount := getPayMoney(1000, "default")
	payAmount := getPayMoneyWithFee(1000, "default", 0.05)
	require.InDelta(t, 950, baseAmount, 0.000001)
	require.InDelta(t, 1000, payAmount, 0.000001)

	settings.AmountDiscount = map[int]float64{}
	payAmount = getPayMoneyWithFee(1000, "default", 0.05)
	require.InDelta(t, 1050, payAmount, 0.000001)
}

func TestGetPaymentMethodFeeRatePrefersMethodOverride(t *testing.T) {
	settings := operation_setting.GetPaymentSetting()
	originalFeeRate := settings.FeeRate
	originalPayMethods := operation_setting.PayMethods
	t.Cleanup(func() {
		settings.FeeRate = originalFeeRate
		operation_setting.PayMethods = originalPayMethods
	})

	settings.FeeRate = 0.05
	operation_setting.PayMethods = []map[string]string{
		{"type": "alipay", "fee_rate": "0.03"},
		{"type": "wxpay", "fee_rate": "0"},
	}
	require.Equal(t, 0.03, getPaymentMethodFeeRate("alipay"))
	require.Equal(t, 0.0, getPaymentMethodFeeRate("wxpay"))
	require.Equal(t, 0.05, getPaymentMethodFeeRate("custom"))
}
