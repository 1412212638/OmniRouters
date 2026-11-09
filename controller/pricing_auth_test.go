package controller

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/config"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// Exercise the public pricing route with the same dashboard Bearer credentials
// as the default frontend, not a manually injected user ID or legacy cookie.
func TestPublicPricingUsesDashboardCustomerMultiplier(t *testing.T) {
	previousDB, previousLogDB := model.DB, model.LOG_DB
	previousRedis := common.RedisEnabled
	previousMainType, previousLogType := common.MainDatabaseType(), common.LogDatabaseType()
	previousGroups := setting.UserUsableGroups2JSONString()
	saved := map[string]string{}
	require.NoError(t, config.GlobalConfig.SaveToDB(func(key, value string) error {
		saved[key] = value
		return nil
	}))
	common.OptionMapRWMutex.Lock()
	previousOptions := common.OptionMap
	common.OptionMap = map[string]string{"HeaderNavModules": `{"pricing":{"enabled":true,"requireAuth":false}}`}
	common.OptionMapRWMutex.Unlock()
	t.Cleanup(func() {
		model.DB, model.LOG_DB = previousDB, previousLogDB
		common.RedisEnabled = previousRedis
		common.SetDatabaseTypes(previousMainType, previousLogType)
		require.NoError(t, config.GlobalConfig.LoadFromDB(saved))
		require.NoError(t, setting.UpdateUserUsableGroupsByJSONString(previousGroups))
		common.OptionMapRWMutex.Lock()
		common.OptionMap = previousOptions
		common.OptionMapRWMutex.Unlock()
		model.InvalidatePricingCache()
	})
	db := setupModelListControllerTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.UserSession{}, &model.QuotaData{}))
	require.NoError(t, config.GlobalConfig.LoadFromDB(map[string]string{
		"group_ratio_setting.group_ratio":              `{"volcengine":1}`,
		"group_ratio_setting.group_group_ratio":        `{}`,
		"group_ratio_setting.group_model_ratio":        `{"volcengine":{"doubao-seedream-5.0-pro":0.8}}`,
		"group_ratio_setting.group_model_user_ratio":   `{"volcengine":{"doubao-seedream-5.0-pro":{"9740":0.6}}}`,
		"group_ratio_setting.group_model_ratio_expiry": `{}`,
	}))
	require.NoError(t, setting.UpdateUserUsableGroupsByJSONString(`{"volcengine":"Volcengine"}`))
	const modelName = "doubao-seedream-5.0-pro"
	require.NoError(t, db.Create(&model.Channel{
		Id: 1, Type: constant.ChannelTypeOpenAI, Key: "test-key", Name: "test", Status: common.ChannelStatusEnabled,
	}).Error)
	require.NoError(t, db.Create(&model.Ability{
		Group: "volcengine", Model: modelName, ChannelId: 1, Enabled: true,
	}).Error)
	tokens := map[int]string{}
	for _, id := range []int{9740, 9889} {
		require.NoError(t, db.Create(&model.User{
			Id: id, Username: fmt.Sprintf("pricing-user-%d", id), Password: "test",
			AffCode: fmt.Sprintf("pricing-%d", id),
			Group: "volcengine", Role: common.RoleCommonUser, Status: common.UserStatusEnabled, AuthVersion: 1,
		}).Error)
		sid := fmt.Sprintf("pricing-session-%d", id)
		require.NoError(t, db.Create(&model.UserSession{
			SID: sid, UserID: id, Version: 1, UserAuthVersion: 1, Status: model.UserSessionStatusActive,
			RefreshHash: "test-refresh-hash", LoginMethod: "password", ExpiresAt: time.Now().Add(time.Hour).Unix(),
		}).Error)
		token, _, err := service.IssueAccessToken(service.AuthIdentity{
			UserID: id, SessionID: sid, UserAuthVersion: 1, SessionVersion: 1,
		})
		require.NoError(t, err)
		tokens[id] = token
	}
	model.InvalidatePricingCache()
	router := gin.New()
	router.Use(sessions.Sessions("session", cookie.NewStore([]byte("pricing-auth-test"))))
	router.GET("/api/pricing", middleware.HeaderNavModuleAuth("pricing"), GetPricing)
	requestPricing := func(token string, userID int) *httptest.ResponseRecorder {
		t.Helper()
		request := httptest.NewRequest(http.MethodGet, "/api/pricing", nil)
		if token != "" {
			request.Header.Set("Authorization", "Bearer "+token)
		}
		if userID != 0 {
			request.Header.Set("New-Api-User", strconv.Itoa(userID))
		}
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)
		return recorder
	}
	for _, tc := range []struct {
		name  string
		id    int
		token string
		ratio float64
	}{
		{"anonymous", 0, "", 0.8},
		{"customer", 9740, tokens[9740], 0.6},
		{"other customer", 9889, tokens[9889], 0.8},
		{"customer after other request", 9740, tokens[9740], 0.6},
		{"anonymous after customer", 0, "", 0.8},
		{"user ID header alone is not authentication", 9740, "", 0.8},
	} {
		t.Run(tc.name, func(t *testing.T) {
			recorder := requestPricing(tc.token, tc.id)
			require.Equal(t, http.StatusOK, recorder.Code)
			require.Equal(t, "private, no-store", recorder.Header().Get("Cache-Control"))
			require.Contains(t, recorder.Header().Values("Vary"), "Authorization, Cookie, New-Api-User")
			var response struct {
				Success         bool                         `json:"success"`
				Data            []model.Pricing              `json:"data"`
				GroupModelRatio map[string]map[string]float64 `json:"group_model_ratio"`
			}
			require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &response))
			require.True(t, response.Success)
			require.Len(t, response.Data, 1)
			require.Equal(t, tc.ratio, response.GroupModelRatio["volcengine"][modelName])
			require.Equal(t, tc.ratio, response.Data[0].ModelGroupRatio["volcengine"])
		})
	}
	t.Run("mismatched user header", func(t *testing.T) {
		require.Equal(t, http.StatusUnauthorized, requestPricing(tokens[9740], 9889).Code)
	})
	t.Run("missing user header", func(t *testing.T) {
		require.Equal(t, http.StatusUnauthorized, requestPricing(tokens[9740], 0).Code)
	})
	t.Run("revoked session", func(t *testing.T) {
		require.NoError(t, db.Model(&model.UserSession{}).Where("user_id = ?", 9740).
			Update("status", model.UserSessionStatusRevoked).Error)
		require.Equal(t, http.StatusUnauthorized, requestPricing(tokens[9740], 9740).Code)
	})
}
