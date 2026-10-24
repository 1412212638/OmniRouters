package controller

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/stretchr/testify/require"
)

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
