package operation_setting

import (
	"strings"
	"sync/atomic"
	"time"
)

const (
	DefaultClientGoneDrainTimeoutSeconds = 10
	MinClientGoneDrainTimeoutSeconds     = 1
	MaxClientGoneDrainTimeoutSeconds     = 300
)

var clientGoneDrainTimeoutSeconds atomic.Int64

func init() {
	clientGoneDrainTimeoutSeconds.Store(DefaultClientGoneDrainTimeoutSeconds)
}

// GetClientGoneDrainTimeout returns the bounded wait for provider terminal
// usage after a downstream streaming client disconnects.
func GetClientGoneDrainTimeout() time.Duration {
	return time.Duration(clientGoneDrainTimeoutSeconds.Load()) * time.Second
}

func GetClientGoneDrainTimeoutSeconds() int {
	return int(clientGoneDrainTimeoutSeconds.Load())
}

func SetClientGoneDrainTimeoutSeconds(seconds int) bool {
	if seconds < MinClientGoneDrainTimeoutSeconds || seconds > MaxClientGoneDrainTimeoutSeconds {
		return false
	}
	clientGoneDrainTimeoutSeconds.Store(int64(seconds))
	return true
}

var DemoSiteEnabled = false
var SelfUseModeEnabled = false

var AutomaticDisableKeywords = []string{
	"Your credit balance is too low",
	"This organization has been disabled.",
	"You exceeded your current quota",
	"Permission denied",
	"The security token included in the request is invalid",
	"Operation not allowed",
	"Your account is not authorized",
}

func AutomaticDisableKeywordsToString() string {
	return strings.Join(AutomaticDisableKeywords, "\n")
}

func AutomaticDisableKeywordsFromString(s string) {
	AutomaticDisableKeywords = []string{}
	ak := strings.Split(s, "\n")
	for _, k := range ak {
		k = strings.TrimSpace(k)
		k = strings.ToLower(k)
		if k != "" {
			AutomaticDisableKeywords = append(AutomaticDisableKeywords, k)
		}
	}
}
