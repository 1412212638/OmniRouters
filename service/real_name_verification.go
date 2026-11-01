package service

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
)

const (
	RealNameEnabledOption          = "real_name_verification.enabled"
	RealNameRequiredOption         = "real_name_verification.enforcement_enabled"
	RealNamePersonalEnabledOption  = "real_name_verification.personal_enabled"
	RealNamePersonalAppCodeOption  = "real_name_verification.personal_app_code"
	RealNameEnterpriseEnabledOption = "real_name_verification.enterprise_enabled"
	RealNameEnterpriseAppCodeOption = "real_name_verification.enterprise_app_code"
	RealNamePersonalConfiguredOption = "real_name_verification.personal_app_code_configured"
	RealNameEnterpriseConfiguredOption = "real_name_verification.enterprise_app_code_configured"

	RealNameTypePersonal   = "personal"
	RealNameTypeEnterprise = "enterprise"
	RealNameStatusVerified = "verified"

	personalVerifyURL   = "https://zidv2.market.alicloudapi.com/idcheck/Post"
	enterpriseVerifyURL = "https://qysysjk.market.alicloudapi.com/government/safety/9151"
)

type RealNameConfig struct {
	Enabled           bool
	Required          bool
	EncryptionConfigured bool
	PersonalEnabled   bool
	PersonalAppCode   string
	EnterpriseEnabled bool
	EnterpriseAppCode string
}

func optionBool(key string) bool {
	common.OptionMapRWMutex.RLock()
	defer common.OptionMapRWMutex.RUnlock()
	value := strings.TrimSpace(common.OptionMap[key])
	return value == "true" || value == "1"
}

func optionString(key string) string {
	common.OptionMapRWMutex.RLock()
	defer common.OptionMapRWMutex.RUnlock()
	return strings.TrimSpace(common.OptionMap[key])
}

func GetRealNameConfig() RealNameConfig {
	return RealNameConfig{
		Enabled:           optionBool(RealNameEnabledOption),
		Required:          optionBool(RealNameRequiredOption),
		EncryptionConfigured: strings.TrimSpace(os.Getenv("REAL_NAME_ENCRYPTION_KEY")) != "",
		PersonalEnabled:   optionBool(RealNamePersonalEnabledOption),
		PersonalAppCode:   optionString(RealNamePersonalAppCodeOption),
		EnterpriseEnabled: optionBool(RealNameEnterpriseEnabledOption),
		EnterpriseAppCode: optionString(RealNameEnterpriseAppCodeOption),
	}
}

func (cfg RealNameConfig) PersonalAvailable() bool {
	return cfg.Enabled && cfg.EncryptionConfigured && cfg.PersonalEnabled && cfg.PersonalAppCode != ""
}

func (cfg RealNameConfig) EnterpriseAvailable() bool {
	return cfg.Enabled && cfg.EncryptionConfigured && cfg.EnterpriseEnabled && cfg.EnterpriseAppCode != ""
}

func RealNameRequiredForUser(userID int) (bool, error) {
	cfg := GetRealNameConfig()
	if !cfg.Required || !cfg.Enabled {
		return false, nil
	}
	verified, err := model.HasVerifiedRealName(userID)
	return !verified, err
}

func encryptionKey() ([]byte, error) {
	secret := strings.TrimSpace(os.Getenv("REAL_NAME_ENCRYPTION_KEY"))
	if secret == "" {
		return nil, errors.New("REAL_NAME_ENCRYPTION_KEY is not configured")
	}
	digest := sha256.Sum256([]byte(secret))
	return digest[:], nil
}

func EncryptRealName(value string) (string, error) {
	key, err := encryptionKey()
	if err != nil { return "", err }
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nil, nonce, []byte(value), nil)
	return base64.RawURLEncoding.EncodeToString(append(nonce, ciphertext...)), nil
}

func DecryptRealName(value string) (string, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return "", err
	}
	key, err := encryptionKey()
	if err != nil { return "", err }
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(decoded) < gcm.NonceSize() {
		return "", errors.New("invalid encrypted identity value")
	}
	return string(decoded[gcm.NonceSize():]), nil
}

type personalProviderResponse struct {
	ErrorCode int `json:"error_code"`
	Result    struct {
		IsOK bool `json:"isok"`
		OrderNo string `json:"order_no"`
	} `json:"result"`
}

type enterpriseProviderResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	SeqNo   string `json:"seqNo"`
	Data    struct {
		NameResult                 string `json:"name_result"`
		IDCardResult               string `json:"idcard_result"`
		EnterpriseNameResult       string `json:"ent_name_result"`
		RegistrationNumberResult   string `json:"registration_number_result"`
	} `json:"data"`
}

func VerifyPersonal(ctx context.Context, appCode, name, idCard string) (bool, string, error) {
	form := url.Values{"cardNo": {idCard}, "realName": {name}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, personalVerifyURL, bytes.NewBufferString(form.Encode()))
	if err != nil {
		return false, "", err
	}
	req.Header.Set("Authorization", "APPCODE "+appCode)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := (&http.Client{Timeout: 20 * time.Second}).Do(req)
	if err != nil {
		return false, "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return false, "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return false, "", fmt.Errorf("personal verification provider returned HTTP %d", resp.StatusCode)
	}
	var parsed personalProviderResponse
	if err := common.Unmarshal(body, &parsed); err != nil {
		return false, "", err
	}
	if parsed.ErrorCode != 0 {
		return false, "", fmt.Errorf("personal verification provider rejected request (%d)", parsed.ErrorCode)
	}
	return parsed.Result.IsOK, parsed.Result.OrderNo, nil
}

func VerifyEnterprise(ctx context.Context, appCode, company, registrationNumber, name, idCard string) (bool, string, error) {
	query := url.Values{
		"registration_number": {registrationNumber},
		"ent_name":             {company},
		"name":                 {name},
		"idcard":               {idCard},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, enterpriseVerifyURL+"?"+query.Encode(), nil)
	if err != nil {
		return false, "", err
	}
	req.Header.Set("Authorization", "APPCODE "+appCode)
	resp, err := (&http.Client{Timeout: 20 * time.Second}).Do(req)
	if err != nil {
		return false, "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return false, "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return false, "", fmt.Errorf("enterprise verification provider returned HTTP %d", resp.StatusCode)
	}
	var parsed enterpriseProviderResponse
	if err := common.Unmarshal(body, &parsed); err != nil {
		return false, "", err
	}
	if parsed.Code != "10000" {
		return false, parsed.SeqNo, fmt.Errorf("enterprise verification provider rejected request (%s)", parsed.Code)
	}
	matched := parsed.Data.NameResult == "1" && parsed.Data.IDCardResult == "1" && parsed.Data.EnterpriseNameResult == "1" && parsed.Data.RegistrationNumberResult == "1"
	return matched, parsed.SeqNo, nil
}
