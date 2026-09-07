package model

import (
	"time"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

// AuthSessionIdentity binds an authentication flow to a specific session version.
type AuthSessionIdentity struct {
	UserID          int    `json:"user_id"`
	SessionID       string `json:"session_id"`
	UserAuthVersion int64  `json:"auth_version"`
	SessionVersion  int64  `json:"session_version"`
}

// AuthFlowAuthorization is server-owned state carried into a configuration flow
// after a proof has been consumed. ProofID is a database ID, never the proof token.
type AuthFlowAuthorization struct {
	AuthSessionIdentity
	ProofID     int64  `json:"proof_id"`
	Scope       string `json:"scope"`
	ContextHash string `json:"context_hash"`
	Method      string `json:"method"`
}

// ValidateAuthSessionWithTx rechecks the authoritative identity while holding the
// user/session locks until the caller's credential change or flow consumption commits.
func ValidateAuthSessionWithTx(tx *gorm.DB, identity AuthSessionIdentity) error {
	if identity.UserID <= 0 || identity.SessionID == "" || identity.UserAuthVersion <= 0 || identity.SessionVersion <= 0 {
		return ErrUserSessionInactive
	}
	var user User
	if err := lockForUpdate(tx).First(&user, identity.UserID).Error; err != nil {
		return err
	}
	if user.Status != common.UserStatusEnabled || user.AuthVersion != identity.UserAuthVersion {
		return ErrUserSessionInactive
	}
	var session UserSession
	if err := lockForUpdate(tx).Where("sid = ? AND user_id = ?", identity.SessionID, identity.UserID).First(&session).Error; err != nil {
		return err
	}
	if session.Status != UserSessionStatusActive || session.RevokedAt != 0 || session.ExpiresAt <= time.Now().Unix() || session.UserAuthVersion != identity.UserAuthVersion || session.Version != identity.SessionVersion {
		return ErrUserSessionInactive
	}
	return nil
}
