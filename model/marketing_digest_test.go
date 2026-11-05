package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupMarketingDigestModelTest(t *testing.T) {
	t.Helper()
	originalDB := DB
	originalMainType := common.MainDatabaseType()
	originalLogType := common.LogDatabaseType()
	common.SetDatabaseTypes(common.DatabaseTypeSQLite, common.DatabaseTypeSQLite)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	DB = db
	require.NoError(t, DB.AutoMigrate(&User{}, &TopUp{}, &MarketingDigest{}, &MarketingDigestPromotion{}, &MarketingDigestRecipient{}))
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
		DB = originalDB
		common.SetDatabaseTypes(originalMainType, originalLogType)
	})
}

func createMarketingDigestTestUser(t *testing.T, id int, username, group string, status int, lastLogin int64, emailPreference *bool) {
	t.Helper()
	user := User{
		Id: id, Username: username, Password: "password123", DisplayName: username,
		Role: common.RoleCommonUser, Status: status, Group: group, LastLoginAt: lastLogin,
		Email: username + "@example.com", AffCode: "aff" + username,
	}
	if emailPreference != nil {
		user.SetSetting(dto.UserSetting{MarketingEmailEnabled: emailPreference})
	}
	require.NoError(t, DB.Create(&user).Error)
}

func TestMarketingDigestAudienceFilters(t *testing.T) {
	setupMarketingDigestModelTest(t)
	const now = int64(2_000_000_000)
	createMarketingDigestTestUser(t, 1, "recent_paid", "default", common.UserStatusEnabled, now-10, nil)
	createMarketingDigestTestUser(t, 2, "old_paid", "default", common.UserStatusEnabled, now-40*24*60*60, nil)
	createMarketingDigestTestUser(t, 3, "recent_unpaid", "premium", common.UserStatusEnabled, now-20, nil)
	createMarketingDigestTestUser(t, 4, "disabled", "default", common.UserStatusDisabled, now-5, nil)
	createMarketingDigestTestUser(t, 5, "opted_out", "default", common.UserStatusEnabled, now-5, marketingEmailDisabled())
	require.NoError(t, DB.Create(&TopUp{UserId: 1, TradeNo: "paid-1", Status: common.TopUpStatusSuccess}).Error)
	require.NoError(t, DB.Create(&TopUp{UserId: 2, TradeNo: "paid-2", Status: common.TopUpStatusSuccess}).Error)
	require.NoError(t, DB.Create(&TopUp{UserId: 3, TradeNo: "failed-3", Status: "pending"}).Error)

	all, err := GetMarketingDigestAudienceUserIDs("all", "", nil, now-30*24*60*60, 100)
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3, 5}, all)

	paid, err := GetMarketingDigestAudienceUserIDs("topup", "", nil, now-30*24*60*60, 100)
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2}, paid)

	active, err := GetMarketingDigestAudienceUserIDs("active", "", nil, now-30*24*60*60, 100)
	require.NoError(t, err)
	assert.Equal(t, []int{1, 3, 5}, active)

	group, err := GetMarketingDigestAudienceUserIDs("group", "premium", nil, now-30*24*60*60, 100)
	require.NoError(t, err)
	assert.Equal(t, []int{3}, group)
}

func TestMarketingDigestRecipientSnapshotAndRetryAreIdempotent(t *testing.T) {
	setupMarketingDigestModelTest(t)
	require.NoError(t, CreateMarketingDigestRecipients(7, []int{1, 2, 3}))
	require.Error(t, CreateMarketingDigestRecipients(7, []int{1, 2, 3}))

	claimed, err := ClaimMarketingDigestRecipients(7, 1)
	require.NoError(t, err)
	require.Len(t, claimed, 1)
	require.NoError(t, FinishMarketingDigestRecipient(claimed[0].ID, MarketingDigestRecipientSent, "a***@example.com", ""))

	counts, err := GetMarketingDigestRecipientCounts(7)
	require.NoError(t, err)
	assert.Equal(t, int64(3), counts.Total)
	assert.Equal(t, int64(1), counts.Sent)
	assert.Equal(t, int64(2), counts.Pending)

	failedID, err := ClaimMarketingDigestRecipients(7, 1)
	require.NoError(t, err)
	require.Len(t, failedID, 1)
	require.NoError(t, FinishMarketingDigestRecipient(failedID[0].ID, MarketingDigestRecipientFailed, "b***@example.com", "smtp rejected"))

	retried, err := RetryFailedMarketingDigestRecipients(7)
	require.NoError(t, err)
	assert.Equal(t, int64(1), retried)
	counts, err = GetMarketingDigestRecipientCounts(7)
	require.NoError(t, err)
	assert.Equal(t, int64(1), counts.Sent)
	assert.Equal(t, int64(2), counts.Pending)
}

func TestMarketingDigestStaleSendingRecipientsCanResumeWithoutResendingSent(t *testing.T) {
	setupMarketingDigestModelTest(t)
	require.NoError(t, CreateMarketingDigestRecipients(9, []int{1, 2, 3}))

	claimed, err := ClaimMarketingDigestRecipients(9, 2)
	require.NoError(t, err)
	require.Len(t, claimed, 2)
	require.NoError(t, FinishMarketingDigestRecipient(claimed[0].ID, MarketingDigestRecipientSent, "a***@example.com", ""))

	require.NoError(t, FinishStaleMarketingDigestRecipients(9))
	retried, err := RetryFailedMarketingDigestRecipients(9)
	require.NoError(t, err)
	assert.Equal(t, int64(1), retried)

	counts, err := GetMarketingDigestRecipientCounts(9)
	require.NoError(t, err)
	assert.Equal(t, int64(1), counts.Sent)
	assert.Equal(t, int64(2), counts.Pending)
}

func marketingEmailDisabled() *bool {
	enabled := false
	return &enabled
}

