package authflowtest

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func testUserSessionContracts(t *testing.T) {
	require.NoError(t, model.DB.AutoMigrate(&model.User{}, &model.UserSession{}))
	require.NoError(t, model.DB.AutoMigrate(&model.User{}, &model.UserSession{}))
	user := model.User{Username: fmt.Sprintf("session-%d", time.Now().UnixNano()), Password: "unused", Status: common.UserStatusEnabled}
	require.NoError(t, model.DB.Create(&user).Error)
	require.EqualValues(t, 1, user.AuthVersion)
	t.Cleanup(func() { model.DB.Unscoped().Delete(&model.User{}, user.Id) })
	for _, useRedis := range []bool{false, true} {
		t.Run(fmt.Sprintf("session-redis-%t", useRedis), func(t *testing.T) {
			previousEnabled, previousClient := common.RedisEnabled, common.RDB
			common.RedisEnabled = useRedis
			t.Cleanup(func() { common.RedisEnabled = previousEnabled; common.RDB = previousClient })
			if useRedis {
				address := os.Getenv("TEST_REDIS_ADDR")
				if address == "" { t.Skip("TEST_REDIS_ADDR not configured") }
				client := redis.NewClient(&redis.Options{Addr: address})
				require.NoError(t, client.Ping(context.Background()).Err())
				common.RDB = client
				t.Cleanup(func() { _ = client.Close() })
			}
			newSession := func() *model.UserSession {
				session := &model.UserSession{
					SID: fmt.Sprintf("test-%d", time.Now().UnixNano()), UserID: user.Id, UserAuthVersion: 1,
					RefreshHash: strings.Repeat("a", 64), ExpiresAt: time.Now().Add(time.Hour).Unix(),
				}
				require.NoError(t, model.CreateUserSession(session))
				t.Cleanup(func() { model.DB.Where("sid = ?", session.SID).Delete(&model.UserSession{}) })
				return session
			}
			session := newSession()
			identity := model.AuthSessionIdentity{UserID: user.Id, SessionID: session.SID, UserAuthVersion: 1, SessionVersion: 1}
			validate := func(id model.AuthSessionIdentity) error {
				return model.DB.Transaction(func(tx *gorm.DB) error { return model.ValidateAuthSessionWithTx(tx, id) })
			}
			require.NoError(t, validate(identity))
			wrong := identity
			wrong.SessionVersion++
			require.ErrorIs(t, validate(wrong), model.ErrUserSessionInactive)
			_, err := model.GetUserSessionCached(session.SID)
			require.NoError(t, err)
			revoked, err := model.RevokeUserSession(user.Id+1, session.SID, "wrong-owner")
			require.NoError(t, err)
			require.False(t, revoked)
			require.NoError(t, validate(identity))
			revoked, err = model.RevokeUserSession(user.Id, session.SID, "test")
			require.NoError(t, err)
			require.True(t, revoked)
			require.ErrorIs(t, validate(identity), model.ErrUserSessionInactive)
			_, err = model.GetUserSessionCached(session.SID)
			require.ErrorIs(t, err, model.ErrUserSessionInactive)

			// Revocation discovered in the binding transaction must roll back consumption.
			state, flow, err := model.CreateAuthFlow(model.AuthFlowCreate{Purpose: model.AuthFlowPurposeOAuth, Provider: "telegram", Intent: model.AuthFlowIntentBind, UserId: user.Id, SessionId: session.SID, ExpiresAt: time.Now().Add(time.Minute)})
			require.NoError(t, err)
			t.Cleanup(func() { model.DB.Delete(&model.AuthFlow{}, flow.Id) })
			match := model.AuthFlowMatch{Purpose: model.AuthFlowPurposeOAuth, Provider: "telegram", Intent: model.AuthFlowIntentBind, UserId: user.Id, SessionId: session.SID}
			_, err = model.ConsumeAuthFlowWithAction(state, match, func(tx *gorm.DB, _ *model.AuthFlow) error { return model.ValidateAuthSessionWithTx(tx, identity) })
			require.ErrorIs(t, err, model.ErrUserSessionInactive)
			_, err = model.GetAuthFlow(state, match)
			require.NoError(t, err)

			refresh := newSession()
			now := time.Now().Unix()
			_, err = model.RotateUserSessionRefresh(user.Id, refresh.SID, strings.Repeat("x", 64), strings.Repeat("b", 64), now, time.Second)
			require.ErrorIs(t, err, model.ErrUserSessionRefreshInvalid)
			_, err = model.RotateUserSessionRefresh(user.Id, refresh.SID, refresh.RefreshHash, strings.Repeat("b", 64), now, time.Second)
			require.NoError(t, err)
			_, err = model.RotateUserSessionRefresh(user.Id, refresh.SID, refresh.RefreshHash, strings.Repeat("c", 64), now, time.Second)
			require.ErrorIs(t, err, model.ErrUserSessionRefreshRace)
			_, err = model.RotateUserSessionRefresh(user.Id, refresh.SID, refresh.RefreshHash, strings.Repeat("c", 64), now+2, time.Second)
			require.ErrorIs(t, err, model.ErrUserSessionRefreshReuse)
			_, err = model.GetUserSessionCached(refresh.SID)
			require.ErrorIs(t, err, model.ErrUserSessionInactive)

			versioned := newSession()
			identity.SessionID = versioned.SID
			// Version writes are explicit; stale legacy User.Save cannot roll them back.
			require.NoError(t, model.DB.Table("users").Where("id = ?", user.Id).Update("auth_version", 2).Error)
			require.NoError(t, model.DB.Save(&user).Error)
			require.ErrorIs(t, validate(identity), model.ErrUserSessionInactive)
			require.NoError(t, model.DB.Table("users").Where("id = ?", user.Id).Update("auth_version", 1).Error)
			require.NoError(t, model.DB.Model(versioned).Where("sid = ?", versioned.SID).Update("expires_at", time.Now().Add(-time.Minute).Unix()).Error)
			require.ErrorIs(t, validate(identity), model.ErrUserSessionInactive)
		})
	}
}
