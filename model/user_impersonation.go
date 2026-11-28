package model

import (
	"context"
	"errors"
	"time"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

const UserSessionLoginMethodImpersonation = "admin_impersonation"

var ErrImpersonationUnavailable = errors.New("this account cannot be impersonated")

func (session *UserSession) AllowsRole(role int) bool {
	return session.LoginMethod != UserSessionLoginMethodImpersonation || role == common.RoleCommonUser
}

// SwitchAdminUserSession commits the new user session, the administrator
// session revocation, and the audit event together before returning credentials.
func SwitchAdminUserSession(ctx context.Context, actor AuthSessionIdentity, next *UserSession, audit *AuditLog) (*User, error) {
	if next.UserID <= 0 || next.UserID == actor.UserID {
		return nil, ErrImpersonationUnavailable
	}
	var target User
	var previous *UserSession
	err := DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		admin, session, err := ResolveAuthSessionWithTx(tx, actor)
		if err != nil {
			return err
		}
		if admin.Role != common.RoleAdminUser && admin.Role != common.RoleRootUser {
			return ErrImpersonationUnavailable
		}
		if err := lockForUpdate(tx).First(&target, next.UserID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrImpersonationUnavailable
			}
			return err
		}
		if target.Role != common.RoleCommonUser || target.Status != common.UserStatusEnabled || target.AuthVersion <= 0 {
			return ErrImpersonationUnavailable
		}
		if err := CheckUserSessionLimits(tx, target.Id, next.CreatedAt); err != nil {
			return err
		}
		next.UserAuthVersion = target.AuthVersion
		if err := tx.Create(next).Error; err != nil {
			return err
		}
		session.Status = UserSessionStatusRevoked
		session.RevokedAt = time.Now().Unix()
		session.RevokedReason = UserSessionLoginMethodImpersonation
		if err := tx.Model(session).Updates(map[string]any{
			"status": session.Status, "revoked_at": session.RevokedAt, "revoked_reason": session.RevokedReason,
		}).Error; err != nil {
			return err
		}
		previous = session
		audit.Username = admin.Username
		return recordAuditLogWithTx(tx, audit)
	})
	if err != nil {
		return nil, err
	}
	// Authentication also reads the authoritative session row. Publishing the
	// revocation invalidates any cached refresh/verification observations.
	if err := writeUserSessionDenyFence(previous, previous.Status, previous.RevokedAt, previous.RevokedReason); err != nil {
		common.SysError("publish administrator session revocation: " + err.Error())
	}
	return &target, nil
}
