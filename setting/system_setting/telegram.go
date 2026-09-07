package system_setting

import (
	"strings"

	"github.com/QuantumNous/new-api/setting/config"
)

// TelegramSettings is the new OAuth configuration. Legacy bot settings are
// intentionally kept separate so the existing widget flow remains valid.
type TelegramSettings struct {
	Enabled      bool   `json:"enabled"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

var defaultTelegramSettings = TelegramSettings{}

func init() { config.GlobalConfig.Register("telegram", &defaultTelegramSettings) }

func GetTelegramSettings() *TelegramSettings { return &defaultTelegramSettings }

func (s *TelegramSettings) IsConfigured() bool {
	return s != nil && s.Enabled && strings.TrimSpace(s.ClientID) != "" && strings.TrimSpace(s.ClientSecret) != ""
}
