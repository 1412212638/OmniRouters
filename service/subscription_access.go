package service

import (
	"errors"
	"fmt"

	relaycommon "github.com/QuantumNous/new-api/relay/common"
	relayconstant "github.com/QuantumNous/new-api/relay/constant"
)

// IsStandaloneToolRequest identifies the independent Search/Fetch endpoints.
// They are per-call wallet charges even when the caller uses a token key.
func IsStandaloneToolRequest(info *relaycommon.RelayInfo) bool {
	if info == nil {
		return false
	}
	return info.RelayMode == relayconstant.RelayModeToolsSearch || info.RelayMode == relayconstant.RelayModeToolsFetch
}

// IsTokenSubscriptionKey identifies an authenticated key bound to a token
// subscription, including calls to the independent Web Tools endpoints.
func IsTokenSubscriptionKey(info *relaycommon.RelayInfo) bool {
	return info != nil && info.TokenKeyType == "subscription" && info.SubscriptionBillingMode == "token"
}

// IsTokenSubscriptionRequest identifies LLM traffic billed by raw input and
// output tokens from a token-based subscription. Standalone Web Tools use the
// same key but are billed separately from the wallet.
func IsTokenSubscriptionRequest(info *relaycommon.RelayInfo) bool {
	return IsTokenSubscriptionKey(info) && !IsStandaloneToolRequest(info)
}

// ValidateSubscriptionRequest applies the restrictions carried by a
// subscription key after the request model and relay mode are known.
func ValidateSubscriptionRequest(info *relaycommon.RelayInfo) error {
	if info == nil || info.TokenKeyType != "subscription" {
		return nil
	}
	if info.SubscriptionBillingMode != "token" {
		return errors.New("subscription key is bound to an unsupported plan")
	}
	switch info.RelayMode {
	case relayconstant.RelayModeChatCompletions,
		relayconstant.RelayModeCompletions,
		relayconstant.RelayModeResponses,
		relayconstant.RelayModeResponsesCompact,
		relayconstant.RelayModeToolsSearch,
		relayconstant.RelayModeToolsFetch:
		// Supported model request protocols and standalone web tools are allowed.
	default:
		return fmt.Errorf("subscription key does not support this request type")
	}
	if IsStandaloneToolRequest(info) {
		return nil
	}
	for _, allowed := range info.SubscriptionAllowedModels {
		if allowed == info.OriginModelName {
			return nil
		}
	}
	return fmt.Errorf("model %s is not included in this subscription", info.OriginModelName)
}
