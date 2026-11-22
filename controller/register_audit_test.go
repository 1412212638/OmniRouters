package controller

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestPasswordRegistrationAuditUsesCreatedAccount(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)

	previousDB, previousLogDB := model.DB, model.LOG_DB
	registerEnabled, passwordRegisterEnabled := common.RegisterEnabled, common.PasswordRegisterEnabled
	emailVerificationEnabled, redisEnabled := common.EmailVerificationEnabled, common.RedisEnabled
	generateDefaultToken := constant.GenerateDefaultToken
	t.Cleanup(func() {
		model.DB, model.LOG_DB = previousDB, previousLogDB
		common.RegisterEnabled, common.PasswordRegisterEnabled = registerEnabled, passwordRegisterEnabled
		common.EmailVerificationEnabled, common.RedisEnabled = emailVerificationEnabled, redisEnabled
		constant.GenerateDefaultToken = generateDefaultToken
		_ = sqlDB.Close()
	})
	model.DB, model.LOG_DB = db, db
	common.RegisterEnabled, common.PasswordRegisterEnabled = true, true
	common.EmailVerificationEnabled, common.RedisEnabled = false, false
	constant.GenerateDefaultToken = false
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.AuditLog{}, &model.Log{}, &model.WalletFundingEntry{}))

	engine := gin.New()
	engine.Use(middleware.AuditRequests())
	engine.POST("/api/user/register", Register)
	body := `{"username":"audit-signup","password":"private-test-pass","verification_code":"private-test-code"}`
	for _, expected := range []bool{true, false} {
		request := httptest.NewRequest(http.MethodPost, "/api/user/register", strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		engine.ServeHTTP(response, request)
		var result struct {
			Success bool `json:"success"`
		}
		require.NoError(t, common.Unmarshal(response.Body.Bytes(), &result))
		require.Equal(t, expected, result.Success, response.Body.String())
	}

	var user model.User
	require.NoError(t, db.Where("username = ?", "audit-signup").First(&user).Error)
	var entries []model.AuditLog
	require.NoError(t, db.Find(&entries).Error)
	require.Len(t, entries, 1)
	entry := entries[0]
	require.Equal(t, user.Id, entry.UserId)
	require.Equal(t, user.Username, entry.Username)
	require.Equal(t, "user.register", entry.Action)
	require.Equal(t, model.AuditCategorySecurity, entry.Category)
	require.True(t, entry.Success)
	require.Empty(t, entry.TokenRef)
	require.NotContains(t, entry.Other, "private-test")
	var details model.AuditDetails
	require.NoError(t, common.UnmarshalJsonStr(entry.Other, &details))
	require.Equal(t, "password", details.AuthMethod)
	require.Equal(t, "/api/user/register", details.Route)
}
