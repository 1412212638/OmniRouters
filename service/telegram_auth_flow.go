package service

import (
	"errors"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/oauth"
)

var ErrTelegramAuthFlowInvalid = errors.New("telegram authentication flow is invalid")

// TelegramAuthFlowState is stored server-side in AuthFlow.Payload. The
// verifier is never sent back to the browser and the flow token is opaque.
type TelegramAuthFlowState struct {
	CodeVerifier string `json:"code_verifier"`
	ClientID     string `json:"client_id"`
	RedirectURI  string `json:"redirect_uri"`
}

func CreateTelegramAuthFlow(userID int, intent string) (string, *model.AuthFlow, error) {
	if intent != model.AuthFlowIntentLogin && intent != model.AuthFlowIntentBind {
		return "", nil, ErrTelegramAuthFlowInvalid
	}
	flow, err := oauth.NewTelegramOAuthFlow()
	if err != nil {
		return "", nil, err
	}
	payload, err := common.Marshal(TelegramAuthFlowState{
		CodeVerifier: flow.CodeVerifier,
		ClientID:     flow.ClientID,
		RedirectURI:  flow.RedirectURI,
	})
	if err != nil {
		return "", nil, err
	}
	return model.CreateAuthFlow(model.AuthFlowCreate{
		Purpose:   model.AuthFlowPurposeOAuth,
		Provider:  "telegram",
		Intent:    intent,
		UserId:    userID,
		Payload:   string(payload),
		ExpiresAt: time.Now().Add(5 * time.Minute),
	})
}

func ReadTelegramAuthFlow(token, intent string) (*model.AuthFlow, *oauth.TelegramOAuthFlow, error) {
	if strings.TrimSpace(token) == "" || (intent != model.AuthFlowIntentLogin && intent != model.AuthFlowIntentBind) {
		return nil, nil, ErrTelegramAuthFlowInvalid
	}
	flowRecord, err := model.GetAuthFlow(token, model.AuthFlowMatch{
		Purpose: model.AuthFlowPurposeOAuth, Provider: "telegram", Intent: intent,
	})
	if err != nil {
		return nil, nil, err
	}
	var state TelegramAuthFlowState
	if err := common.UnmarshalJsonStr(flowRecord.Payload, &state); err != nil || state.CodeVerifier == "" || state.ClientID == "" || state.RedirectURI == "" {
		return nil, nil, ErrTelegramAuthFlowInvalid
	}
	return flowRecord, &oauth.TelegramOAuthFlow{CodeVerifier: state.CodeVerifier, ClientID: state.ClientID, RedirectURI: state.RedirectURI}, nil
}
