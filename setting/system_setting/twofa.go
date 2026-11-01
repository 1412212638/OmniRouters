package system_setting

import (
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/config"
)

type TwoFASettings struct {
	EnrollmentEnabled bool `json:"enrollment_enabled"`
}

var defaultTwoFASettings = TwoFASettings{EnrollmentEnabled: true}

func init() {
	config.GlobalConfig.Register("two_fa", &defaultTwoFASettings)
}

// TwoFAEnrollmentEnabled controls enrollment only, never existing credentials.
func TwoFAEnrollmentEnabled() bool {
	common.OptionMapRWMutex.RLock()
	value, exists := common.OptionMap["two_fa.enrollment_enabled"]
	common.OptionMapRWMutex.RUnlock()
	value = strings.TrimSpace(value)
	if !exists || value == "" {
		return true
	}
	return strings.EqualFold(value, "true") || value == "1"
}
