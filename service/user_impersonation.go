package service

import (
	"context"
	"time"

	"github.com/QuantumNous/new-api/model"
)

const impersonationSessionTTL = 15 * time.Minute

func SwitchToUserSession(ctx context.Context, actor AuthIdentity, targetID int, ip, userAgent string, audit *model.AuditLog) (*AuthBundle, *model.User, error) {
	session, refreshToken, err := newLoginSession(targetID, model.UserSessionLoginMethodImpersonation, ip, userAgent, time.Now().Unix(), impersonationSessionTTL)
	if err != nil {
		return nil, nil, err
	}
	session.ImpersonatorID = actor.UserID
	target, err := model.SwitchAdminUserSession(ctx, actor, session, audit)
	if err != nil {
		return nil, nil, err
	}
	bundle, err := issueAuthBundle(session, refreshToken, true)
	if err != nil {
		_, _ = model.RevokeUserSession(targetID, session.SID, "token_issue_failed")
		return nil, nil, err
	}
	return bundle, target, nil
}
