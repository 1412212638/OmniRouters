package model

import (
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func resetReferralRewardRows(t *testing.T) {
	t.Helper()
	require.NoError(t, DB.AutoMigrate(&ReferralReward{}))
	require.NoError(t, DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&ReferralReward{}).Error)
	t.Cleanup(func() {
		require.NoError(t, DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&ReferralReward{}).Error)
	})
}

func TestReferralCashbackRateAndQuotaUseBasisPoints(t *testing.T) {
	rateBps, err := referralCashbackRateBps(5.25)
	require.NoError(t, err)
	assert.Equal(t, 525, rateBps)

	quota, err := referralCashbackQuota(common.QuotaPerUnit*10, rateBps)
	require.NoError(t, err)
	assert.Equal(t, common.QuotaPerUnit*10*525/10000, quota)
}

func TestReferralCashbackRateRejectsInvalidValues(t *testing.T) {
	for _, rate := range []float64{-0.01, 100.01} {
		_, err := referralCashbackRateBps(rate)
		assert.Error(t, err)
	}

	_, err := referralCashbackRateBps(0)
	require.NoError(t, err)
	_, err = referralCashbackQuota(100, 10001)
	assert.ErrorIs(t, err, ErrReferralRewardInvalid)
}

func newReferralReward(inviterId, inviteeId int, sourceTopUpId *int) *ReferralReward {
	return &ReferralReward{
		InviterId:     inviterId,
		InviteeId:     inviteeId,
		RewardType:    ReferralRewardTypeFirstTopUpCashback,
		SourceTopUpId: sourceTopUpId,
		BaseQuota:     1000,
		BaseMoney:     2.5,
		RewardRateBps: 500,
		RewardQuota:   50,
	}
}

func TestCreateReferralRewardIfAbsentIsIdempotentByInviteeAndType(t *testing.T) {
	resetReferralRewardRows(t)

	first := newReferralReward(1, 2, nil)
	created, got, err := CreateReferralRewardIfAbsent(nil, first)
	require.NoError(t, err)
	assert.True(t, created)
	assert.Equal(t, first.Id, got.Id)

	second := newReferralReward(1, 2, nil)
	second.RewardQuota = 99
	created, got, err = CreateReferralRewardIfAbsent(nil, second)
	require.NoError(t, err)
	assert.False(t, created)
	assert.Equal(t, first.Id, got.Id)
	assert.Equal(t, 50, got.RewardQuota)

	loaded, err := GetReferralRewardByInviteeType(nil, 2, ReferralRewardTypeFirstTopUpCashback)
	require.NoError(t, err)
	assert.Equal(t, first.Id, loaded.Id)
}

func TestCreateReferralRewardIfAbsentIsIdempotentBySourceTopUp(t *testing.T) {
	resetReferralRewardRows(t)
	source := 42

	first := newReferralReward(1, 2, &source)
	created, _, err := CreateReferralRewardIfAbsent(nil, first)
	require.NoError(t, err)
	assert.True(t, created)

	secondSource := source
	second := newReferralReward(1, 3, &secondSource)
	created, got, err := CreateReferralRewardIfAbsent(nil, second)
	require.NoError(t, err)
	assert.False(t, created)
	assert.Equal(t, first.Id, got.Id)

	loaded, err := GetReferralRewardBySourceTopUp(nil, source, ReferralRewardTypeFirstTopUpCashback)
	require.NoError(t, err)
	assert.Equal(t, first.Id, loaded.Id)
}

func TestCreateReferralRewardIfAbsentValidatesIdentityAndStatus(t *testing.T) {
	resetReferralRewardRows(t)

	invalid := newReferralReward(1, 1, nil)
	_, _, err := CreateReferralRewardIfAbsent(nil, invalid)
	assert.ErrorIs(t, err, ErrReferralRewardInvalid)

	invalid = newReferralReward(1, 2, nil)
	invalid.Status = "unknown"
	_, _, err = CreateReferralRewardIfAbsent(nil, invalid)
	assert.ErrorIs(t, err, ErrReferralRewardStatusInvalid)

	invalid = newReferralReward(1, 2, nil)
	invalid.RewardType = "other"
	_, _, err = CreateReferralRewardIfAbsent(nil, invalid)
	assert.ErrorIs(t, err, ErrReferralRewardTypeInvalid)
}

func insertReferralTestUsers(t *testing.T, inviterID, inviteeID int) {
	t.Helper()
	inviter := &User{Id: inviterID, Username: "referral-inviter", Password: "password", Status: common.UserStatusEnabled, AffQuota: 123}
	invitee := &User{Id: inviteeID, Username: "referral-invitee", Password: "password", Status: common.UserStatusEnabled, InviterId: inviterID}
	require.NoError(t, DB.Create(inviter).Error)
	require.NoError(t, DB.Create(invitee).Error)
}

func insertSuccessfulReferralTopUp(t *testing.T, userID int, tradeNo string) *TopUp {
	t.Helper()
	topUp := &TopUp{
		UserId: userID, Amount: 1, Money: 1, TradeNo: tradeNo,
		PaymentMethod: PaymentMethodStripe, PaymentProvider: PaymentProviderStripe,
		CreateTime: time.Now().Unix(), CompleteTime: time.Now().Unix(), Status: common.TopUpStatusSuccess,
	}
	require.NoError(t, DB.Create(topUp).Error)
	return topUp
}

func applyReferralCashbackForTest(t *testing.T, topUp *TopUp, quota int) {
	t.Helper()
	require.NoError(t, DB.Transaction(func(tx *gorm.DB) error {
		return ApplyFirstTopUpReferralCashbackTx(tx, topUp, quota)
	}))
}

func TestApplyFirstTopUpReferralCashbackIsIdempotent(t *testing.T) {
	resetReferralRewardRows(t)
	truncateTables(t)
	insertReferralTestUsers(t, 8101, 8102)
	oldRate := common.ReferralFirstTopUpCashbackRate
	common.ReferralFirstTopUpCashbackRate = 10
	t.Cleanup(func() { common.ReferralFirstTopUpCashbackRate = oldRate })

	topUp := insertSuccessfulReferralTopUp(t, 8102, "referral-first-topup")
	applyReferralCashbackForTest(t, topUp, 500000)
	applyReferralCashbackForTest(t, topUp, 500000)

	reward, err := GetReferralRewardByInviteeType(nil, 8102, ReferralRewardTypeFirstTopUpCashback)
	require.NoError(t, err)
	assert.Equal(t, 50000, reward.RewardQuota)
	assert.Equal(t, ReferralRewardStatusPending, reward.Status)
	var count int64
	require.NoError(t, DB.Model(&ReferralReward{}).Where("invitee_id = ?", 8102).Count(&count).Error)
	assert.EqualValues(t, 1, count)
}

func TestApplyFirstTopUpReferralCashbackLocksFirstTopUpWhenDisabled(t *testing.T) {
	resetReferralRewardRows(t)
	truncateTables(t)
	insertReferralTestUsers(t, 8201, 8202)
	oldRate := common.ReferralFirstTopUpCashbackRate
	common.ReferralFirstTopUpCashbackRate = 0
	t.Cleanup(func() { common.ReferralFirstTopUpCashbackRate = oldRate })

	topUp := insertSuccessfulReferralTopUp(t, 8202, "referral-disabled-first-topup")
	applyReferralCashbackForTest(t, topUp, 500000)

	reward, err := GetReferralRewardByInviteeType(nil, 8202, ReferralRewardTypeFirstTopUpCashback)
	require.NoError(t, err)
	assert.Equal(t, 0, reward.RewardQuota)
	assert.Equal(t, ReferralRewardStatusNoReward, reward.Status)
	assert.Equal(t, 0, reward.RewardRateBps)
}

func TestApplyFirstTopUpReferralCashbackSkipsPreviouslyRechargedInvitee(t *testing.T) {
	resetReferralRewardRows(t)
	truncateTables(t)
	insertReferralTestUsers(t, 8301, 8302)
	oldRate := common.ReferralFirstTopUpCashbackRate
	common.ReferralFirstTopUpCashbackRate = 10
	t.Cleanup(func() { common.ReferralFirstTopUpCashbackRate = oldRate })

	insertSuccessfulReferralTopUp(t, 8302, "referral-existing-topup")
	firstEligibleTopUp := insertSuccessfulReferralTopUp(t, 8302, "referral-second-topup")
	applyReferralCashbackForTest(t, firstEligibleTopUp, 500000)

	var count int64
	require.NoError(t, DB.Model(&ReferralReward{}).Where("invitee_id = ?", 8302).Count(&count).Error)
	assert.Zero(t, count)
}
