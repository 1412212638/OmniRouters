package service

import (
	"errors"
	"strings"

	"github.com/QuantumNous/new-api/model"
	"gorm.io/gorm"
)

var ErrTelegramOAuthCommitInvalid = errors.New("telegram oauth commit is invalid")

// CommitTelegramLogin atomically consumes the login flow, creates the user,
// and claims the Telegram subject. Caller must perform post-commit effects
// such as cache refresh, rewards, and session issuance after success.
func CommitTelegramLogin(flowToken, subject string, user *model.User, inviterID int) error {
	if strings.TrimSpace(flowToken) == "" || strings.TrimSpace(subject) == "" || user == nil || user.Id != 0 {
		return ErrTelegramOAuthCommitInvalid
	}
	_, err := model.ConsumeAuthFlowWithAction(flowToken, model.AuthFlowMatch{
		Purpose: model.AuthFlowPurposeOAuth, Provider: model.ExternalIdentityProviderTelegram,
		Intent: model.AuthFlowIntentLogin,
	}, func(tx *gorm.DB, _ *model.AuthFlow) error {
		if err := user.InsertWithTx(tx, inviterID); err != nil {
			return err
		}
		return model.BindExternalIdentityWithTx(tx, model.ExternalIdentityProviderTelegram, subject, user.Id)
	})
	return err
}

// CommitTelegramBind atomically consumes a bind flow and claims the subject
// for the flow's already authenticated user.
func CommitTelegramBind(flowToken, subject string, userID int) error {
	if strings.TrimSpace(flowToken) == "" || strings.TrimSpace(subject) == "" || userID <= 0 {
		return ErrTelegramOAuthCommitInvalid
	}
	_, err := model.ConsumeAuthFlowWithAction(flowToken, model.AuthFlowMatch{
		Purpose: model.AuthFlowPurposeOAuth, Provider: model.ExternalIdentityProviderTelegram,
		Intent: model.AuthFlowIntentBind, UserId: userID,
	}, func(tx *gorm.DB, _ *model.AuthFlow) error {
		return model.BindExternalIdentityWithTx(tx, model.ExternalIdentityProviderTelegram, subject, userID)
	})
	return err
}

func CommitTelegramExistingLogin(flowToken, subject string, userID int) error {
	if strings.TrimSpace(flowToken) == "" || strings.TrimSpace(subject) == "" || userID <= 0 {
		return ErrTelegramOAuthCommitInvalid
	}
	_, err := model.ConsumeAuthFlowWithAction(flowToken, model.AuthFlowMatch{
		Purpose: model.AuthFlowPurposeOAuth, Provider: model.ExternalIdentityProviderTelegram,
		Intent: model.AuthFlowIntentLogin,
	}, func(tx *gorm.DB, _ *model.AuthFlow) error {
		return model.BindExternalIdentityWithTx(tx, model.ExternalIdentityProviderTelegram, subject, userID)
	})
	return err
}
