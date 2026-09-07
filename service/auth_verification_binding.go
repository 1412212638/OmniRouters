package service

import (
	"encoding/json"
	"errors"

	"github.com/QuantumNous/new-api/common"
)

// VerificationBinding is the server-derived scope/context binding carried by
// a security proof. It intentionally contains no raw operation parameters.
type VerificationBinding struct {
	Scope       string `json:"scope"`
	ContextHash string `json:"context_hash"`
}

const (
	VerificationScopeChannelKeyRead  = "channel.key.read"
	VerificationScopePasskeyRegister = "passkey.register"
	VerificationScopePasskeyDelete   = "passkey.delete"
	VerificationScopeTwoFASetup      = "2fa.setup"
)

var ErrVerificationContextInvalid = errors.New("the action details are invalid")

// BindVerificationOperation normalizes the small set of supported operation
// contexts before hashing; callers never place raw request JSON in a proof.
func BindVerificationOperation(operation VerificationOperation) (VerificationBinding, error) {
	var fields map[string]json.RawMessage
	if len(operation.Context) > 0 {
		if common.GetJsonType(operation.Context) != "object" || common.Unmarshal(operation.Context, &fields) != nil {
			return VerificationBinding{}, ErrVerificationContextInvalid
		}
	}
	var normalized any
	switch operation.Scope {
	case VerificationScopeChannelKeyRead:
		var id int
		if len(fields) != 1 || common.Unmarshal(fields["channel_id"], &id) != nil || id <= 0 {
			return VerificationBinding{}, ErrVerificationContextInvalid
		}
		normalized = struct{ ChannelID int `json:"channel_id"` }{id}
	case VerificationScopePasskeyRegister, VerificationScopePasskeyDelete, VerificationScopeTwoFASetup:
		if len(fields) != 0 {
			return VerificationBinding{}, ErrVerificationContextInvalid
		}
		normalized = struct{}{}
	default:
		return VerificationBinding{}, ErrProofScope
	}
	payload, err := common.Marshal(struct {
		Scope string `json:"scope"`
		Context any `json:"context"`
	}{operation.Scope, normalized})
	if err != nil { return VerificationBinding{}, err }
	return VerificationBinding{Scope: operation.Scope, ContextHash: common.GenerateHMACWithKey(authSigningKey("verification-context"), string(payload))}, nil
}

type VerificationOperation struct {
	Scope string `json:"scope"`
	Context json.RawMessage `json:"context,omitempty"`
}
