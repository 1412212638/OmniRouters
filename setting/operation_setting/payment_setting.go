package operation_setting

import (
	"errors"
	"github.com/Calcium-Ion/go-epay/epay"
	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/config"
	"strings"
)

type PaymentSetting struct {
	EpaySignType           string          `json:"epay_sign_type"`
	EpayMerchantPrivateKey string          `json:"epay_merchant_private_key"`
	EpayPlatformPublicKey  string          `json:"epay_platform_public_key"`
	AmountOptions          []int           `json:"amount_options"`
	AmountDiscount         map[int]float64 `json:"amount_discount"`
	FeeRate                float64         `json:"fee_rate"`

	CorporateTransferEnabled       bool   `json:"corporate_transfer_enabled"`
	CorporateTransferTitle         string `json:"corporate_transfer_title"`
	CorporateTransferNotice        string `json:"corporate_transfer_notice"`
	CorporateTransferAccountName   string `json:"corporate_transfer_account_name"`
	CorporateTransferBankName      string `json:"corporate_transfer_bank_name"`
	CorporateTransferBankAccount   string `json:"corporate_transfer_bank_account"`
	CorporateTransferOperatorPhone string `json:"corporate_transfer_operator_phone"`
	CorporateTransferSupportEmail  string `json:"corporate_transfer_support_email"`
}

const (
	EpaySignTypeMD5                 = "MD5"
	EpaySignTypeRSA                 = "RSA"
	EpaySignTypeOptionKey           = "payment_setting.epay_sign_type"
	EpayMerchantPrivateKeyOptionKey = "payment_setting.epay_merchant_private_key"
	EpayPlatformPublicKeyOptionKey  = "payment_setting.epay_platform_public_key"
)

var paymentSetting = PaymentSetting{
	EpaySignType:   EpaySignTypeMD5,
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

// ValidateEpaySignOption checks a change to the Epay signing settings against
// the saved ones. While RSA is active, both keys must parse; under MD5 only a
// newly entered key is checked.
func ValidateEpaySignOption(key, value string) error {
	if key != EpaySignTypeOptionKey && key != EpayMerchantPrivateKeyOptionKey && key != EpayPlatformPublicKeyOptionKey {
		return nil
	}
	// Saved values are written under OptionMapRWMutex by the option update path.
	common.OptionMapRWMutex.RLock()
	signType := paymentSetting.EpaySignType
	privateKey := paymentSetting.EpayMerchantPrivateKey
	publicKey := paymentSetting.EpayPlatformPublicKey
	common.OptionMapRWMutex.RUnlock()
	switch key {
	case EpaySignTypeOptionKey:
		if value != EpaySignTypeMD5 && value != EpaySignTypeRSA {
			return errors.New("Epay signature type must be MD5 or RSA")
		}
		signType = value
	case EpayMerchantPrivateKeyOptionKey:
		privateKey = value
	case EpayPlatformPublicKeyOptionKey:
		publicKey = value
	}
	rsaActive := signType == EpaySignTypeRSA
	if rsaActive || (key == EpayMerchantPrivateKeyOptionKey && strings.TrimSpace(value) != "") {
		if _, err := epay.ParseRSAPrivateKey(privateKey); err != nil {
			return errors.New("The Epay merchant private key is missing or invalid")
		}
	}
	if rsaActive || (key == EpayPlatformPublicKeyOptionKey && strings.TrimSpace(value) != "") {
		if _, err := epay.ParseRSAPublicKey(publicKey); err != nil {
			return errors.New("The Epay platform public key is missing or invalid")
		}
	}
	return nil
}
