package router

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

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

func TestAdminUserImpersonation(t *testing.T) {
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
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.UserSession{}, &model.AuditLog{}, &model.Log{}))

	pat := "fixture-admin-pat"
	admin := model.User{Username: "switch-admin", Role: common.RoleAdminUser, AccessToken: &pat}
	target := model.User{Username: "switch-target", Role: common.RoleCommonUser}
	otherAdmin := model.User{Username: "other-admin", Role: common.RoleAdminUser}
	disabled := model.User{Username: "switch-disabled", Role: common.RoleCommonUser, Status: common.UserStatusDisabled}
	for index, user := range []*model.User{&admin, &target, &otherAdmin, &disabled} {
		user.AffCode = strconv.Itoa(index + 1000)
		require.NoError(t, db.Create(user).Error)
	}
	login := func(userID int) *service.AuthBundle {
		t.Helper()
		bundle, err := service.CreateLoginSession(userID, "password", "", "")
		require.NoError(t, err)
		return bundle
	}
	actor := login(admin.Id)
	independentAdmin := login(admin.Id)
	regular := login(target.Id)

	engine := gin.New()
	engine.Use(sessions.Sessions("session", cookie.NewStore([]byte("user-impersonation-router-test"))))
	SetApiRouter(engine)
	request := func(method, path, token string, userID int, refresh string, body string) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Origin", "http://example.com")
		if body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		if userID > 0 {
			req.Header.Set("New-Api-User", strconv.Itoa(userID))
		}
		if refresh != "" {
			req.AddCookie(&http.Cookie{Name: service.RefreshCookieName, Value: refresh})
		}
		resp := httptest.NewRecorder()
		engine.ServeHTTP(resp, req)
		return resp
	}
	start := func(targetID int, token string, userID int) *httptest.ResponseRecorder {
		return request(http.MethodPost, "/api/user/"+strconv.Itoa(targetID)+"/impersonate", token, userID, "", "")
	}
	self := func(token string, userID int) *httptest.ResponseRecorder {
		return request(http.MethodGet, "/api/user/self", token, userID, "", "")
	}
	refresh := func(cookie string) *httptest.ResponseRecorder {
		return request(http.MethodPost, "/api/user/auth/refresh", "", 0, cookie, "")
	}
	readBundle := func(response *httptest.ResponseRecorder) (string, *http.Cookie) {
		t.Helper()
		require.Equal(t, http.StatusOK, response.Code, response.Body.String())
		var result struct {
			Success bool
			Data    struct {
				AccessToken string `json:"access_token"`
				User        struct{ ID int }
				Session     service.LoginSessionView
			}
		}
		require.NoError(t, common.Unmarshal(response.Body.Bytes(), &result))
		require.True(t, result.Success)
		require.Equal(t, target.Id, result.Data.User.ID)
		require.Equal(t, model.UserSessionLoginMethodImpersonation, result.Data.Session.LoginMethod)
		require.InDelta(t, time.Now().Add(15*time.Minute).Unix(), result.Data.Session.ExpiresAt, 5)
		require.NotEmpty(t, result.Data.AccessToken)
		for _, cookie := range response.Result().Cookies() {
			if cookie.Name == service.RefreshCookieName {
				require.True(t, cookie.HttpOnly)
				require.Equal(t, "/api/user/auth", cookie.Path)
				require.NotContains(t, response.Body.String(), cookie.Value)
				return result.Data.AccessToken, cookie
			}
		}
		t.Fatal("missing refresh cookie")
		return "", nil
	}

	require.Equal(t, http.StatusUnauthorized, start(target.Id, "", 0).Code)
	require.Equal(t, http.StatusForbidden, start(target.Id, pat, admin.Id).Code)
	require.Contains(t, start(target.Id, regular.AccessToken, target.Id).Body.String(), `"success":false`)
	for _, id := range []int{admin.Id, otherAdmin.Id, disabled.Id, disabled.Id + 100} {
		require.Equal(t, http.StatusForbidden, start(id, actor.AccessToken, admin.Id).Code)
	}

	// An audit failure must leave the prior session usable and withhold the new cookie.
	require.NoError(t, db.Migrator().DropTable(&model.AuditLog{}))
	failed := start(target.Id, actor.AccessToken, admin.Id)
	require.Equal(t, http.StatusInternalServerError, failed.Code, failed.Body.String())
	require.Empty(t, failed.Header().Get("Set-Cookie"))
	require.Contains(t, self(actor.AccessToken, admin.Id).Body.String(), `"username":"switch-admin"`)
	var targetSessions int64
	require.NoError(t, db.Model(&model.UserSession{}).Where("user_id = ?", target.Id).Count(&targetSessions).Error)
	require.EqualValues(t, 1, targetSessions)
	require.NoError(t, db.AutoMigrate(&model.AuditLog{}))

	issued := start(target.Id, actor.AccessToken, admin.Id)
	require.Contains(t, issued.Header().Get("Cache-Control"), "no-store")
	access, refreshCookie := readBundle(issued)
	require.Equal(t, http.StatusUnauthorized, self(actor.AccessToken, admin.Id).Code)
	require.Equal(t, http.StatusUnauthorized, refresh(actor.RefreshToken).Code)
	require.Equal(t, http.StatusUnauthorized, start(target.Id, actor.AccessToken, admin.Id).Code)
	require.Contains(t, self(independentAdmin.AccessToken, admin.Id).Body.String(), `"username":"switch-admin"`)
	require.Contains(t, self(regular.AccessToken, target.Id).Body.String(), `"username":"switch-target"`)
	require.Contains(t, self(access, target.Id).Body.String(), `"username":"switch-target"`)
	require.Equal(t, http.StatusUnauthorized, self(access, admin.Id).Code)
	require.Contains(t, request(http.MethodGet, "/api/user/", access, target.Id, "", "").Body.String(), `"success":false`)

	var startAudit model.AuditLog
	require.NoError(t, db.Where("user_id = ? AND action = ? AND success = ?", admin.Id, "user.impersonation.start", true).First(&startAudit).Error)
	require.Contains(t, startAudit.Other, `"target_user_id":`+strconv.Itoa(target.Id))
	require.NotContains(t, startAudit.Other, access)
	var userAudit model.AuditLog
	require.NoError(t, db.Where("user_id = ? AND other LIKE ?", target.Id, `%"impersonator_id":`+strconv.Itoa(admin.Id)+`%`).First(&userAudit).Error)
	require.Contains(t, userAudit.Other, `"route":"/api/user/self"`)

	access, refreshCookie = readBundle(refresh(refreshCookie.Value))
	require.Contains(t, self(access, target.Id).Body.String(), `"username":"switch-target"`)
	logout := request(http.MethodPost, "/api/user/auth/logout", access, target.Id, refreshCookie.Value, "")
	require.Equal(t, http.StatusOK, logout.Code, logout.Body.String())
	require.Equal(t, http.StatusUnauthorized, self(access, target.Id).Code)
	require.Equal(t, http.StatusUnauthorized, refresh(refreshCookie.Value).Code)

	actor = login(admin.Id)
	access, refreshCookie = readBundle(start(target.Id, actor.AccessToken, admin.Id))
	require.NoError(t, db.Model(&model.User{}).Where("id = ?", target.Id).UpdateColumn("role", common.RoleAdminUser).Error)
	require.Equal(t, http.StatusUnauthorized, self(access, target.Id).Code)
	require.Equal(t, http.StatusUnauthorized, refresh(refreshCookie.Value).Code)
	require.NoError(t, db.Model(&model.User{}).Where("id = ?", target.Id).UpdateColumn("role", common.RoleCommonUser).Error)

	actor = login(admin.Id)
	access, refreshCookie = readBundle(start(target.Id, actor.AccessToken, admin.Id))
	identity, err := service.ParseAccessToken(access)
	require.NoError(t, err)
	require.NoError(t, db.Model(&model.UserSession{}).Where("sid = ?", identity.SessionID).UpdateColumn("expires_at", time.Now().Add(-time.Second).Unix()).Error)
	require.Equal(t, http.StatusUnauthorized, self(access, target.Id).Code)
	require.Equal(t, http.StatusUnauthorized, refresh(refreshCookie.Value).Code)
}
