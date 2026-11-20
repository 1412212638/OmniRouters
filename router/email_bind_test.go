package router

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestEmailBindWithDashboardSession(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })

	oldDB, oldLogDB, oldRedis := model.DB, model.LOG_DB, common.RedisEnabled
	oldMainType, oldLogType := common.MainDatabaseType(), common.LogDatabaseType()
	model.DB, model.LOG_DB, common.RedisEnabled = db, db, false
	common.SetDatabaseTypes(common.DatabaseTypeSQLite, common.DatabaseTypeSQLite)
	t.Cleanup(func() {
		model.DB, model.LOG_DB, common.RedisEnabled = oldDB, oldLogDB, oldRedis
		common.SetDatabaseTypes(oldMainType, oldLogType)
	})
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.UserSession{}, &model.AuditLog{}))

	const oldEmail = "old@example.com"
	const newEmail = "new@example.com"
	const code = "abc123"
	personalToken := "test-personal-access-token"
	user := model.User{
		Username: "email-bind-user", Email: oldEmail, Role: common.RoleCommonUser,
		Status: common.UserStatusEnabled, AuthVersion: 1, AccessToken: &personalToken,
	}
	require.NoError(t, db.Create(&user).Error)
	login, err := service.CreateLoginSession(user.Id, "password", "", "")
	require.NoError(t, err)
	common.RegisterVerificationCodeWithKey(newEmail, code, common.EmailVerificationPurpose)
	t.Cleanup(func() { common.DeleteKey(newEmail, common.EmailVerificationPurpose) })

	engine := gin.New()
	engine.Use(gin.Recovery(), sessions.Sessions("session", cookie.NewStore([]byte("email-bind-test"))))
	SetApiRouter(engine)
	bind := func(token, verificationCode string) *httptest.ResponseRecorder {
		t.Helper()
		body, err := common.Marshal(map[string]string{"email": "New@example.com", "code": verificationCode})
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodPost, "/api/oauth/email/bind", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("New-Api-User", strconv.Itoa(user.Id))
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		resp := httptest.NewRecorder()
		engine.ServeHTTP(resp, req)
		return resp
	}
	require.Equal(t, http.StatusUnauthorized, bind("", code).Code)
	require.Equal(t, http.StatusForbidden, bind(personalToken, code).Code)

	invalid := bind(login.AccessToken, "incorrect")
	require.Equal(t, http.StatusOK, invalid.Code, invalid.Body.String())
	require.Contains(t, invalid.Body.String(), `"success":false`)
	var stored model.User
	require.NoError(t, db.First(&stored, user.Id).Error)
	require.Equal(t, oldEmail, stored.Email)

	bound := bind(login.AccessToken, code)
	require.Equal(t, http.StatusOK, bound.Code, bound.Body.String())
	require.Contains(t, bound.Body.String(), `"success":true`)
	require.NoError(t, db.First(&stored, user.Id).Error)
	require.Equal(t, newEmail, stored.Email)
	require.NoError(t, db.Where("user_id = ? AND action = ? AND success = ?", user.Id, "email.bind", true).
		First(&model.AuditLog{}).Error)
}
