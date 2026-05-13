package operation_setting

import "github.com/QuantumNous/new-api/setting/config"

type PaymentSetting struct {
	AmountOptions  []int           `json:"amount_options"`
	AmountDiscount map[int]float64 `json:"amount_discount"`
	FeeRate        float64         `json:"fee_rate"`

	CorporateTransferEnabled       bool   `json:"corporate_transfer_enabled"`
	CorporateTransferTitle         string `json:"corporate_transfer_title"`
	CorporateTransferNotice        string `json:"corporate_transfer_notice"`
	CorporateTransferAccountName   string `json:"corporate_transfer_account_name"`
	CorporateTransferBankName      string `json:"corporate_transfer_bank_name"`
	CorporateTransferBankAccount   string `json:"corporate_transfer_bank_account"`
	CorporateTransferOperatorPhone string `json:"corporate_transfer_operator_phone"`
	CorporateTransferSupportEmail  string `json:"corporate_transfer_support_email"`
}

var paymentSetting = PaymentSetting{
	AmountOptions:  []int{10, 20, 50, 100, 200, 500},
	AmountDiscount: map[int]float64{},
	FeeRate:        0,

	CorporateTransferEnabled: false,
	CorporateTransferTitle:   "Corporate Bank Transfer",
	CorporateTransferNotice:  "",
}

func init() {
	config.GlobalConfig.Register("payment_setting", &paymentSetting)
}

func GetPaymentSetting() *PaymentSetting {
	return &paymentSetting
}
