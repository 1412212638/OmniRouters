package controller

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func openMarketingTestDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	store, err := db.DB()
	require.NoError(t, err)
	store.SetMaxOpenConns(1)
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.MarketingEmailSuppression{}, &model.Option{}))
	previous := model.DB
	redisEnabled := common.RedisEnabled
	model.DB = db
	common.RedisEnabled = false
	t.Cleanup(func() {
		model.DB = previous
		common.RedisEnabled = redisEnabled
		require.NoError(t, store.Close())
	})
}

func TestMarketingOneClickUnsubscribe(t *testing.T) {
	openMarketingTestDB(t)
	user := &model.User{Username: "reader", Status: common.UserStatusEnabled}
	require.NoError(t, model.DB.Create(user).Error)
	r := gin.New()
	r.GET("/unsubscribe", ShowMarketingEmailUnsubscribe)
	r.POST("/unsubscribe", ApplyMarketingEmailUnsubscribe)
	id := strconv.Itoa(user.Id)
	query := url.Values{"user_id": {id}, "token": {common.GenerateHMAC("marketing-unsubscribe:" + id)}}
	path := "/unsubscribe?" + query.Encode()
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	require.Equal(t, http.StatusOK, rec.Code)
	user, err := model.GetUserById(user.Id, false)
	require.NoError(t, err)
	require.True(t, model.IsMarketingEmailUserEligible(user))

	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader("List-Unsubscribe=One-Click"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	user, err = model.GetUserById(user.Id, false)
	require.NoError(t, err)
	require.False(t, model.IsMarketingEmailUserEligible(user))
}

func TestMarketingDeliverySettingsRoundTripAndPreservePassword(t *testing.T) {
	openMarketingTestDB(t)
	originalOptions := common.OptionMap
	common.OptionMap = map[string]string{}
	t.Cleanup(func() { common.OptionMap = originalOptions })
	r := gin.New()
	r.GET("/delivery", GetMarketingSMTPSettings)
	r.PUT("/delivery", UpdateMarketingSMTPSettings)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/delivery", nil))
	require.Contains(t, rec.Body.String(), `"mode":"inherit"`)

	cfg := common.MarketingSMTPConfig{
		Mode: "smtp", Server: "smtp.example.com", Port: 465,
		Account: "marketing@example.com", Password: "smtp-password-marker",
		From: "news@example.com", PerSecond: 5,
	}
	save := func() string {
		data, err := common.Marshal(cfg)
		require.NoError(t, err)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/delivery", bytes.NewReader(data)))
		require.Equal(t, http.StatusOK, rec.Code)
		require.NotContains(t, rec.Body.String(), "smtp-password-marker")
		return rec.Body.String()
	}
	require.Contains(t, save(), `"password_configured":true`)
	cfg.From, cfg.Password = "announcements@example.com", ""
	require.Contains(t, save(), `"from":"announcements@example.com"`)
	stored, err := model.GetMarketingSMTPConfig()
	require.NoError(t, err)
	require.Equal(t, "smtp-password-marker", stored.Password)
	cfg.Server = "another.example.com"
	require.Contains(t, save(), `"success":false`)
	stored, err = model.GetMarketingSMTPConfig()
	require.NoError(t, err)
	require.Equal(t, "smtp.example.com", stored.Server)
}

func TestMarketingSESCallbackSuppressesOnlyMatchingAddress(t *testing.T) {
	openMarketingTestDB(t)
	r := gin.New()
	r.POST("/ses", MarketingSESEvent)
	const sender = "old-sender@example.com"
	const original = "reader@example.net"
	const messageID = "abc@example.com"
	proof := service.MarketingSESProof(messageID, sender, original)
	post := func(email, event, bounceType string) int {
		t.Helper()
		data, err := common.Marshal(map[string]string{
			"event": event, "bounceType": bounceType, "messageId": "<" + messageID + ">",
			"X-Tencentcloudses-Cb-Proof": proof, "email": email, "from": sender,
		})
		require.NoError(t, err)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/ses?key="+marketingSESCallbackKey(), strings.NewReader(string(data))))
		return rec.Code
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/ses", nil))
	require.Equal(t, http.StatusForbidden, rec.Code)
	require.Equal(t, http.StatusOK, post(original, "bounce", "soft_bounce"))
	suppressed, err := model.IsMarketingEmailSuppressed(original)
	require.NoError(t, err)
	require.False(t, suppressed)
	require.Equal(t, http.StatusForbidden, post("other@example.net", "spamreport", ""))
	require.Equal(t, http.StatusOK, post(original, "bounce", "hard_bounce"))
	require.Equal(t, http.StatusOK, post(original, "spamreport", ""))
	suppressed, err = model.IsMarketingEmailSuppressed("READER@example.net")
	require.NoError(t, err)
	require.True(t, suppressed)
}
