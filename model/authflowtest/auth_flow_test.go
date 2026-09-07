package authflowtest

import (
	"errors"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestAuthFlowDatabaseContracts(t *testing.T) {
	for _, engine := range []string{"sqlite", "mysql", "postgres"} {
		t.Run(engine, func(t *testing.T) {
			var dialector gorm.Dialector
			switch engine {
			case "sqlite":
				dialector = sqlite.Open(":memory:")
			case "mysql":
				dsn := os.Getenv("TEST_MYSQL_DSN")
				if dsn == "" { t.Skip("TEST_MYSQL_DSN not configured") }
				dialector = mysql.Open(dsn)
			case "postgres":
				dsn := os.Getenv("TEST_POSTGRES_DSN")
				if dsn == "" { t.Skip("TEST_POSTGRES_DSN not configured") }
				dialector = postgres.Open(dsn)
			}
			db, err := gorm.Open(dialector, &gorm.Config{})
			require.NoError(t, err)
			sqlDB, err := db.DB()
			require.NoError(t, err)
			if engine == "sqlite" { sqlDB.SetMaxOpenConns(1) }
			previous := model.DB
			model.DB = db
			t.Cleanup(func() { model.DB = previous; _ = sqlDB.Close() })
			require.NoError(t, db.AutoMigrate(&model.AuthFlow{}))
			require.NoError(t, db.AutoMigrate(&model.AuthFlow{}))
			testAuthFlowContracts(t)
		})
	}
}

func testAuthFlowContracts(t *testing.T) {
	newFlow := func() (string, *model.AuthFlow) {
		t.Helper()
		token, flow, err := model.CreateAuthFlow(model.AuthFlowCreate{
			Purpose: model.AuthFlowPurposeOAuth, Provider: "telegram", Intent: model.AuthFlowIntentBind,
			UserId: 42, SessionId: "session-original", Payload: `{"code_verifier":"server-only"}`,
			ExpiresAt: time.Now().Add(10*time.Minute),
		})
		require.NoError(t, err)
		t.Cleanup(func() { model.DB.Delete(&model.AuthFlow{}, flow.Id) })
		require.NotEqual(t, token, flow.TokenHash)
		require.Len(t, flow.TokenHash, 64)
		return token, flow
	}
	match := model.AuthFlowMatch{Purpose: model.AuthFlowPurposeOAuth, Provider: "telegram", Intent: model.AuthFlowIntentBind, UserId: 42, SessionId: "session-original"}
	token, _ := newFlow()
	for _, field := range []string{"purpose", "provider", "intent", "user", "session"} {
		wrong := match
		switch field {
		case "purpose": wrong.Purpose = "different"
		case "provider": wrong.Provider = "different"
		case "intent": wrong.Intent = model.AuthFlowIntentLogin
		case "user": wrong.UserId = 43
		case "session": wrong.SessionId = "session-other"
		}
		_, err := model.ConsumeAuthFlow(token, wrong)
		require.ErrorIs(t, err, model.ErrAuthFlowInvalid)
	}
	_, err := model.GetAuthFlow(token, match)
	require.NoError(t, err)

	rollback := errors.New("binding failed")
	_, err = model.ConsumeAuthFlowWithAction(token, match, func(tx *gorm.DB, flow *model.AuthFlow) error {
		require.NoError(t, tx.Model(flow).Update("payload", "changed").Error)
		return rollback
	})
	require.ErrorIs(t, err, rollback)
	stored, err := model.GetAuthFlow(token, match)
	require.NoError(t, err)
	require.Equal(t, `{"code_verifier":"server-only"}`, stored.Payload)
	_, err = model.ConsumeAuthFlow(token, match)
	require.NoError(t, err)
	_, err = model.ConsumeAuthFlow(token, match)
	require.ErrorIs(t, err, model.ErrAuthFlowConsumed)

	expiredToken, expired := newFlow()
	require.NoError(t, model.DB.Model(expired).Update("expires_at", time.Now().Add(-time.Hour)).Error)
	_, err = model.ConsumeAuthFlow(expiredToken, match)
	require.ErrorIs(t, err, model.ErrAuthFlowExpired)

	concurrentToken, _ := newFlow()
	var winners atomic.Int32
	var callbacks atomic.Int32
	var wg sync.WaitGroup
	start := make(chan struct{})
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, err := model.ConsumeAuthFlowWithAction(concurrentToken, match, func(*gorm.DB, *model.AuthFlow) error {
				callbacks.Add(1)
				return nil
			})
			if err == nil { winners.Add(1) } else if !errors.Is(err, model.ErrAuthFlowConsumed) { t.Errorf("unexpected consume error: %v", err) }
		}()
	}
	close(start)
	wg.Wait()
	require.EqualValues(t, 1, winners.Load())
	require.EqualValues(t, 1, callbacks.Load())

	assertionToken, _ := newFlow()
	assertionHash := common.GenerateHMACWithKey([]byte("auth-flow-v1:"+common.SessionSecret), "external:" + model.AuthFlowPurposeTelegramAssertion + ":" + assertionToken)
	t.Cleanup(func() { model.DB.Where("token_hash = ?", assertionHash).Delete(&model.AuthFlow{}) })
	require.NoError(t, model.ClaimExternalAuthAssertion(model.AuthFlowPurposeTelegramAssertion, assertionToken, time.Now().Add(time.Minute)))
	require.ErrorIs(t, model.ClaimExternalAuthAssertion(model.AuthFlowPurposeTelegramAssertion, assertionToken, time.Now().Add(time.Minute)), model.ErrAuthFlowConsumed)
	_, _, err = model.CreateAuthFlow(model.AuthFlowCreate{Purpose: model.AuthFlowPurposeOAuth, ExpiresAt: time.Now().Add(-time.Minute)})
	require.ErrorIs(t, err, model.ErrAuthFlowInvalid)
}
