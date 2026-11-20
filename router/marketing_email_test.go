package router

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestMarketingUnsubscribePostAllowsOpaqueOriginWithValidSignature(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	oldDB, oldRedis := model.DB, common.RedisEnabled
	model.DB, common.RedisEnabled = db, false
	t.Cleanup(func() {
		model.DB, common.RedisEnabled = oldDB, oldRedis
		require.NoError(t, sqlDB.Close())
	})
	require.NoError(t, db.AutoMigrate(&model.User{}))
	user := model.User{Username: "unsubscribe-reader", Status: common.UserStatusEnabled}
	require.NoError(t, db.Create(&user).Error)

	engine := gin.New()
	SetApiRouter(engine)
	const path = "/api/marketing_email/unsubscribe"
	id := strconv.Itoa(user.Id)
	token := common.GenerateHMAC("marketing-unsubscribe:" + id)
	post := func(signature string) *httptest.ResponseRecorder {
		form := url.Values{"user_id": {id}, "token": {signature}}
		req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Origin", "null")
		res := httptest.NewRecorder()
		engine.ServeHTTP(res, req)
		return res
	}

	require.Equal(t, http.StatusBadRequest, post("invalid").Code)
	stored, err := model.GetUserById(user.Id, false)
	require.NoError(t, err)
	require.True(t, model.IsMarketingEmailUserEligible(stored))

	res := post(token)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())
	require.Contains(t, res.Body.String(), "Unsubscribed")
	require.Empty(t, res.Header().Get("Access-Control-Allow-Origin"))
	stored, err = model.GetUserById(user.Id, false)
	require.NoError(t, err)
	require.False(t, model.IsMarketingEmailUserEligible(stored))

	req := httptest.NewRequest(http.MethodGet, "/api/about", nil)
	req.Header.Set("Origin", "null")
	res = httptest.NewRecorder()
	engine.ServeHTTP(res, req)
	require.Equal(t, http.StatusForbidden, res.Code)
	require.Contains(t, res.Body.String(), "CORS_ORIGIN_FORBIDDEN")
}

func TestMarketingCustomAddressUnsubscribe(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	oldDB := model.DB
	model.DB = db
	t.Cleanup(func() { model.DB = oldDB })
	require.NoError(t, db.AutoMigrate(&model.MarketingEmailSuppression{}))

	const address = "external@example.net"
	hash := model.MarketingEmailAddressHash(address)
	token := common.GenerateHMAC("marketing-address-unsubscribe:" + hash)
	path := "/api/marketing_email/unsubscribe"
	form := url.Values{"email_hash": {hash}, "token": {token}}
	engine := gin.New()
	SetApiRouter(engine)
	res := httptest.NewRecorder()
	engine.ServeHTTP(res, httptest.NewRequest(http.MethodGet, path+"?"+form.Encode(), nil))
	require.Equal(t, http.StatusOK, res.Code)
	require.Contains(t, res.Body.String(), `name="email_hash"`)
	require.NotContains(t, res.Body.String(), address)
	suppressed, err := model.IsMarketingEmailSuppressed(address)
	require.NoError(t, err)
	require.False(t, suppressed)

	for _, oneClick := range []bool{false, true} {
		target, body := path, form.Encode()
		if oneClick {
			target, body = path+"?"+form.Encode(), "List-Unsubscribe=One-Click"
		}
		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Origin", "null")
		res := httptest.NewRecorder()
		engine.ServeHTTP(res, req)
		require.Equal(t, http.StatusOK, res.Code, res.Body.String())
	}
	suppressed, err = model.IsMarketingEmailSuppressed("EXTERNAL@example.net")
	require.NoError(t, err)
	require.True(t, suppressed)
}
