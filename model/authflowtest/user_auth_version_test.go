package authflowtest

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func testUserAuthVersionWrites(t *testing.T) {
	t.Run("auth-version-writes", func(t *testing.T) {
		previousRedis := common.RedisEnabled
		common.RedisEnabled = false
		t.Cleanup(func() { common.RedisEnabled = previousRedis })
		require.NoError(t, model.DB.AutoMigrate(&model.User{}, &model.TwoFA{}, &model.TwoFABackupCode{}))
		user := model.User{Username: fmt.Sprintf("av-%d", time.Now().UnixNano()), Password: "unused", Status: common.UserStatusEnabled}
		require.NoError(t, model.DB.Create(&user).Error)
		t.Cleanup(func() {
			model.DB.Unscoped().Where("user_id = ?", user.Id).Delete(&model.TwoFA{})
			model.DB.Unscoped().Delete(&model.User{}, user.Id)
		})
		version := func() int64 {
			t.Helper()
			var stored model.User
			require.NoError(t, model.DB.First(&stored, user.Id).Error)
			return stored.AuthVersion
		}

		// This is the model path reached by POST /api/user/2fa/setup.
		factor := model.TwoFA{UserId: user.Id, Secret: "test-secret", IsEnabled: false}
		require.NoError(t, factor.Create())
		require.EqualValues(t, 2, version())
		// A stale legacy save must still leave the advanced version intact.
		require.NoError(t, model.DB.Save(&user).Error)
		require.EqualValues(t, 2, version())

		rollback := errors.New("rollback security change")
		err := model.DB.Transaction(func(tx *gorm.DB) error {
			next, err := model.IncrementUserAuthVersionWithTx(tx, user.Id)
			require.NoError(t, err)
			require.EqualValues(t, 3, next)
			return rollback
		})
		require.ErrorIs(t, err, rollback)
		require.EqualValues(t, 2, version())
		require.NoError(t, factor.Delete())
		require.EqualValues(t, 3, version())

		// Upgrade initialization must also bypass the create-only field guard.
		require.NoError(t, model.DB.Table("users").Where("id = ?", user.Id).Update("auth_version", 0).Error)
		require.NoError(t, model.InitializeUserAuthVersions())
		require.EqualValues(t, 1, version())
		require.NoError(t, model.InitializeUserAuthVersions())
		require.EqualValues(t, 1, version())
	})
}
