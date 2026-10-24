package controller

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/stretchr/testify/require"
)

func TestFormatWaffoPancakeAmount_UsesDisplayPriceString(t *testing.T) {
	testCases := []struct {
		name     string
		amount   float64
		expected string
	}{
		{name: "whole amount", amount: 29, expected: "29.00"},
		{name: "decimal amount", amount: 29.9, expected: "29.90"},
		{name: "round half up to cents", amount: 29.999, expected: "30.00"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, formatWaffoPancakeAmount(tc.amount))
		})
	}
}

func TestGetWaffoPancakeBillingDetailUsesSupportedCompanyFields(t *testing.T) {
	profile := &model.InvoiceProfile{
		Type:        "company",
		CompanyName: "Example Ltd",
		TaxID:       "GB123456789",
		Email:       "billing@example.com",
		Country:     "gb",
		State:       "London",
		Postcode:    "SW1A 1AA",
		Address:     "10 Downing Street",
	}
	profileBytes, err := common.Marshal(profile)
	require.NoError(t, err)

	detail := getWaffoPancakeBillingDetail(&model.User{InvoiceProfile: string(profileBytes)})
	require.NotNil(t, detail)
	require.Equal(t, "GB", detail.Country)
	require.True(t, detail.IsBusiness)
	require.Equal(t, "London", detail.State)
	require.Equal(t, "SW1A 1AA", detail.Postcode)
	require.Equal(t, "Example Ltd", detail.BusinessName)
	require.Equal(t, "GB123456789", detail.TaxID)

	profile.Country = ""
	profileBytes, err = common.Marshal(profile)
	require.NoError(t, err)
	require.Nil(t, getWaffoPancakeBillingDetail(&model.User{InvoiceProfile: string(profileBytes)}))
}

func TestGetWaffoPancakePayMoney(t *testing.T) {
	originalUnitPrice := setting.WaffoPancakeUnitPrice
	originalQuotaDisplayType := operation_setting.GetGeneralSetting().QuotaDisplayType
	originalDiscounts := make(map[int]float64, len(operation_setting.GetPaymentSetting().AmountDiscount))
	for k, v := range operation_setting.GetPaymentSetting().AmountDiscount {
		originalDiscounts[k] = v
	}
	originalTopupGroupRatio := common.TopupGroupRatio2JSONString()

	t.Cleanup(func() {
		setting.WaffoPancakeUnitPrice = originalUnitPrice
		operation_setting.GetGeneralSetting().QuotaDisplayType = originalQuotaDisplayType
		operation_setting.GetPaymentSetting().AmountDiscount = originalDiscounts
		require.NoError(t, common.UpdateTopupGroupRatioByJSONString(originalTopupGroupRatio))
	})

	setting.WaffoPancakeUnitPrice = 2.5
	operation_setting.GetPaymentSetting().AmountDiscount = map[int]float64{
		10:                           0.8,
		int(common.QuotaPerUnit * 3): 0.5,
		20:                           0,
	}
	require.NoError(t, common.UpdateTopupGroupRatioByJSONString(`{"default":1,"vip":1.2}`))

	testCases := []struct {
		name             string
		amount           int64
		group            string
		quotaDisplayType string
		expected         float64
	}{
		{
			name:             "currency display applies unit price group ratio and discount",
			amount:           10,
			group:            "vip",
			quotaDisplayType: operation_setting.QuotaDisplayTypeUSD,
			expected:         27.6,
		},
		{
			name:             "tokens display converts quota to display units before pricing",
			amount:           int64(common.QuotaPerUnit * 3),
			group:            "vip",
			quotaDisplayType: operation_setting.QuotaDisplayTypeTokens,
			expected:         7.2,
		},
		{
			name:             "non-positive discount falls back to no discount",
			amount:           20,
			group:            "default",
			quotaDisplayType: operation_setting.QuotaDisplayTypeUSD,
			expected:         50,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			operation_setting.GetGeneralSetting().QuotaDisplayType = tc.quotaDisplayType
			actual := getWaffoPancakePayMoney(tc.amount, tc.group)
			require.InDelta(t, tc.expected, actual, 0.000001)
		})
	}
}

func TestGetWaffoPancakePriceQuoteSeparatesMarkupFromBase(t *testing.T) {
	originalUnitPrice := setting.WaffoPancakeUnitPrice
	originalQuotaDisplayType := operation_setting.GetGeneralSetting().QuotaDisplayType
	originalDiscounts := operation_setting.GetPaymentSetting().AmountDiscount
	originalTopupGroupRatio := common.TopupGroupRatio2JSONString()
	common.OptionMapRWMutex.Lock()
	originalOptionPrice, hadOptionPrice := common.OptionMap["WaffoPancakeUnitPrice"]
	common.OptionMap["WaffoPancakeUnitPrice"] = "1.05"
	common.OptionMapRWMutex.Unlock()
	t.Cleanup(func() {
		setting.WaffoPancakeUnitPrice = originalUnitPrice
		operation_setting.GetGeneralSetting().QuotaDisplayType = originalQuotaDisplayType
		operation_setting.GetPaymentSetting().AmountDiscount = originalDiscounts
		require.NoError(t, common.UpdateTopupGroupRatioByJSONString(originalTopupGroupRatio))
		common.OptionMapRWMutex.Lock()
		if hadOptionPrice {
			common.OptionMap["WaffoPancakeUnitPrice"] = originalOptionPrice
		} else {
			delete(common.OptionMap, "WaffoPancakeUnitPrice")
		}
		common.OptionMapRWMutex.Unlock()
	})

	setting.WaffoPancakeUnitPrice = 1.05
	operation_setting.GetGeneralSetting().QuotaDisplayType = operation_setting.QuotaDisplayTypeUSD
	operation_setting.GetPaymentSetting().AmountDiscount = map[int]float64{}
	require.NoError(t, common.UpdateTopupGroupRatioByJSONString(`{"default":1}`))

	baseAmount, payAmount := getWaffoPancakePriceQuote(10, "default")
	require.Equal(t, "10.00", baseAmount.Round(2).StringFixed(2))
	require.Equal(t, "10.50", payAmount.Round(2).StringFixed(2))
	require.Equal(t, "0.50", payAmount.Sub(baseAmount).Round(2).StringFixed(2))

	operation_setting.GetPaymentSetting().AmountDiscount = map[int]float64{10: 0.95}
	baseAmount, payAmount = getWaffoPancakePriceQuote(10, "default")
	require.Equal(t, "9.50", baseAmount.Round(2).StringFixed(2))
	require.Equal(t, "10.00", payAmount.Round(2).StringFixed(2))
	require.Equal(t, "0.50", payAmount.Sub(baseAmount).Round(2).StringFixed(2))
}
