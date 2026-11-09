package system_setting

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/config"
)

const (
	CORSModeDisabled = "disabled"
	CORSModeAllowlist = "allowlist"
	CORSModeAll       = "all"
)

var DefaultCORSOrigins = []string{
	"https://tokenplan.omnirouters.com",
	"https://omnirouters.com.cn",
	"https://www.omnirouters.com.cn",
}

type CORSSettings struct {
	Mode           string   `json:"mode"`
	AllowedOrigins []string `json:"allowed_origins"`
}

var defaultCORSSettings = CORSSettings{
	Mode:           CORSModeAllowlist,
	AllowedOrigins: append([]string(nil), DefaultCORSOrigins...),
}

func init() {
	config.GlobalConfig.Register("cors_setting", &defaultCORSSettings)
}

func GetCORSSettings() CORSSettings {
	settings := CORSSettings{
		Mode:           defaultCORSSettings.Mode,
		AllowedOrigins: append([]string(nil), defaultCORSSettings.AllowedOrigins...),
	}
	common.OptionMapRWMutex.RLock()
	mode, hasMode := common.OptionMap["cors_setting.mode"]
	originsValue, hasOrigins := common.OptionMap["cors_setting.allowed_origins"]
	common.OptionMapRWMutex.RUnlock()
	if hasMode && strings.TrimSpace(mode) != "" {
		settings.Mode = strings.TrimSpace(mode)
	}
	if hasOrigins {
		var origins []string
		if common.UnmarshalJsonStr(originsValue, &origins) == nil {
			settings.AllowedOrigins = origins
		} else {
			settings.AllowedOrigins = nil
		}
	}
	return settings
}

func ValidateCORSMode(value string) error {
	switch strings.TrimSpace(value) {
	case CORSModeDisabled, CORSModeAllowlist, CORSModeAll:
		return nil
	default:
		return fmt.Errorf("invalid API CORS mode")
	}
}

func ValidateCORSOrigins(value string) error {
	var origins []string
	if err := common.UnmarshalJsonStr(value, &origins); err != nil {
		return fmt.Errorf("API CORS origins must be a JSON array")
	}
	if len(origins) > 128 {
		return fmt.Errorf("API CORS origins cannot contain more than 128 entries")
	}
	for _, origin := range origins {
		if err := validateCORSOrigin(origin); err != nil {
			return err
		}
	}
	return nil
}

func validateCORSOrigin(origin string) error {
	origin = strings.TrimSpace(origin)
	if origin == "" || len(origin) > 2048 {
		return fmt.Errorf("API CORS origin cannot be empty or exceed 2048 characters")
	}
	parsed, err := url.Parse(origin)
	if err != nil || parsed.Host == "" || parsed.User != nil || parsed.Opaque != "" {
		return fmt.Errorf("invalid API CORS origin: %s", origin)
	}
	if !strings.EqualFold(parsed.Scheme, "http") && !strings.EqualFold(parsed.Scheme, "https") {
		return fmt.Errorf("API CORS origins must use http or https")
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" || parsed.ForceQuery || (parsed.Path != "" && parsed.Path != "/") {
		return fmt.Errorf("API CORS origins must not contain a path or query")
	}
	return nil
}
