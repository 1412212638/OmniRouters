package controller

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestManageMultiKeysRecoveryPersistsRoutingState(t *testing.T) {
	gin.SetMode(gin.TestMode)
	previousDB, previousLogDB, previousCache, previousRedis := model.DB, model.LOG_DB, common.MemoryCacheEnabled, common.RedisEnabled
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	raw, err := db.DB()
	require.NoError(t, err)
	raw.SetMaxOpenConns(1)
	model.DB, model.LOG_DB, common.MemoryCacheEnabled, common.RedisEnabled = db, db, true, false
	t.Cleanup(func() {
		model.DB, model.LOG_DB, common.MemoryCacheEnabled, common.RedisEnabled = previousDB, previousLogDB, previousCache, previousRedis
		if previousDB != nil && previousCache {
			model.InitChannelCache()
		}
		_ = raw.Close()
	})
	require.NoError(t, db.AutoMigrate(&model.Channel{}, &model.Ability{}, &model.Log{}, &model.User{}))

	for _, tc := range []struct {
		name, action, reason string
		status, want         int
		keyIndex             string
	}{
		{"enable one", "enable_key", "All keys are disabled", common.ChannelStatusAutoDisabled, common.ChannelStatusEnabled, `,"key_index":0`},
		{"enable all", "enable_all_keys", "All keys are disabled", common.ChannelStatusAutoDisabled, common.ChannelStatusEnabled, ""},
		{"manual disable", "enable_key", "All keys are disabled", common.ChannelStatusManuallyDisabled, common.ChannelStatusManuallyDisabled, `,"key_index":0`},
		{"other auto disable", "enable_all_keys", "other failure", common.ChannelStatusAutoDisabled, common.ChannelStatusAutoDisabled, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, db.Exec("DELETE FROM abilities").Error)
			require.NoError(t, db.Exec("DELETE FROM channels").Error)
			channel := &model.Channel{Key: "a\nb", Status: tc.status, Models: "test-model", Group: "default", ChannelInfo: model.ChannelInfo{IsMultiKey: true, MultiKeySize: 2, MultiKeyStatusList: map[int]int{0: common.ChannelStatusAutoDisabled, 1: common.ChannelStatusAutoDisabled}}}
			channel.SetOtherInfo(map[string]interface{}{"status_reason": tc.reason})
			require.NoError(t, channel.Insert())
			model.InitChannelCache()
			before, err := model.GetRandomSatisfiedChannel("default", "test-model", 0, nil)
			require.NoError(t, err)
			require.Nil(t, before)

			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest(http.MethodPost, "/api/channel/multi-key", bytes.NewBufferString(fmt.Sprintf(`{"channel_id":%d,"action":%q%s}`, channel.Id, tc.action, tc.keyIndex)))
			ctx.Request.Header.Set("Content-Type", "application/json")
			ManageMultiKeys(ctx)
			require.Equal(t, http.StatusOK, recorder.Code)
			var result struct {
				Success bool   `json:"success"`
				Message string `json:"message"`
			}
			require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &result))
			require.True(t, result.Success, result.Message)

			stored, err := model.GetChannelById(channel.Id, true)
			require.NoError(t, err)
			require.Equal(t, tc.want, stored.Status)
			require.NotContains(t, stored.ChannelInfo.MultiKeyStatusList, 0)
			var ability model.Ability
			require.NoError(t, db.Where("channel_id = ?", channel.Id).First(&ability).Error)
			require.Equal(t, tc.want == common.ChannelStatusEnabled, ability.Enabled)
			cached, err := model.CacheGetChannel(channel.Id)
			require.NoError(t, err)
			require.Equal(t, tc.want, cached.Status)
			routed, err := model.GetRandomSatisfiedChannel("default", "test-model", 0, nil)
			require.NoError(t, err)
			if tc.want == common.ChannelStatusEnabled {
				require.NotNil(t, routed)
			} else {
				require.Nil(t, routed)
			}
		})
	}
}

func TestRestoreAutoDisabledMultiKeyChannel(t *testing.T) {
	cases := []struct {
		name     string
		status   int
		reason   string
		keys     string
		multi    bool
		disabled map[int]int
		want     int
	}{
		{"one key enabled", common.ChannelStatusAutoDisabled, "All keys are disabled", "a\nb", true, map[int]int{1: common.ChannelStatusAutoDisabled}, common.ChannelStatusEnabled},
		{"all keys enabled", common.ChannelStatusAutoDisabled, "All keys are disabled", "a\nb", true, nil, common.ChannelStatusEnabled},
		{"all keys remain disabled", common.ChannelStatusAutoDisabled, "All keys are disabled", "a\nb", true, map[int]int{0: common.ChannelStatusAutoDisabled, 1: common.ChannelStatusAutoDisabled}, common.ChannelStatusAutoDisabled},
		{"different auto disable reason", common.ChannelStatusAutoDisabled, "other failure", "a", true, nil, common.ChannelStatusAutoDisabled},
		{"manual disable preserved", common.ChannelStatusManuallyDisabled, "All keys are disabled", "a", true, nil, common.ChannelStatusManuallyDisabled},
		{"single key channel preserved", common.ChannelStatusAutoDisabled, "All keys are disabled", "a", false, nil, common.ChannelStatusAutoDisabled},
		{"empty key list preserved", common.ChannelStatusAutoDisabled, "All keys are disabled", "", true, nil, common.ChannelStatusAutoDisabled},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			channel := &model.Channel{Key: tc.keys, Status: tc.status, ChannelInfo: model.ChannelInfo{IsMultiKey: tc.multi, MultiKeyStatusList: tc.disabled}}
			channel.SetOtherInfo(map[string]interface{}{"status_reason": tc.reason})
			restoreAutoDisabledMultiKeyChannel(channel)
			require.Equal(t, tc.want, channel.Status)
			if tc.want == common.ChannelStatusEnabled {
				require.Equal(t, "", channel.GetOtherInfo()["status_reason"])
			} else {
				require.Equal(t, tc.reason, channel.GetOtherInfo()["status_reason"])
			}
		})
	}
	restoreAutoDisabledMultiKeyChannel(nil)
}
