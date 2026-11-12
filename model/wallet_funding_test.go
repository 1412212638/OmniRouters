package model

import (
	"errors"
	"fmt"
	"math/big"
	"sync"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupWalletFundingTest(t *testing.T) *User {
	t.Helper()
	previousDB, previousLogDB := DB, LOG_DB
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	DB, LOG_DB = db, db
	require.NoError(t, db.AutoMigrate(&User{}, &WalletFundingEntry{}, &Redemption{}, &Checkin{}, &TopUp{}, &Log{}))
	t.Cleanup(func() {
		DB, LOG_DB = previousDB, previousLogDB
		_ = sqlDB.Close()
	})
	user := &User{Username: "funding-test", Quota: 100, Status: common.UserStatusEnabled}
	require.NoError(t, DB.Create(user).Error)
	return user
}

func TestFundingCreditIdempotencyAndRefundExclusion(t *testing.T) {
	user := setupWalletFundingTest(t)
	var wg sync.WaitGroup
	errs := make([]error, 5)
	for i := range errs {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			errs[index] = CreditUserFunding(user.Id, 200, FundingGift, "invitation", "invitee:test", "", 0, "")
		}(i)
	}
	wg.Wait()
	for _, err := range errs {
		require.NoError(t, err)
	}
	// The existing refund primitive must never create a new funding credit.
	require.NoError(t, IncreaseUserQuota(user.Id, 30, true))
	require.NoError(t, DB.First(user, user.Id).Error)
	assert.Equal(t, 330, user.Quota)
	rows, total, summary, err := GetWalletFunding(user.Id, 0, 20, "")
	require.NoError(t, err)
	assert.Len(t, rows, 1)
	assert.EqualValues(t, 1, total)
	assert.Equal(t, "200", summary.Gift)
	assert.Equal(t, "0", summary.Paid)
}

func TestFundingCreditRollbackAtWalletLimit(t *testing.T) {
	user := setupWalletFundingTest(t)
	require.NoError(t, DB.Model(user).Update("quota", common.MaxWalletQuota).Error)
	require.Error(t, CreditUserFunding(user.Id, 1, FundingPaid, "admin", "overflow:test", "", 1, ""))
	var count int64
	require.NoError(t, DB.Model(&WalletFundingEntry{}).Count(&count).Error)
	assert.Zero(t, count, "a failed credit must roll back its idempotency record")
	require.NoError(t, DB.Model(user).Update("quota", 100).Error)
	require.NoError(t, CreditUserFunding(user.Id, 1, FundingPaid, "admin", "overflow:test", "", 1, ""))
}

func TestFundingEventCannotBeReusedWithDifferentCredit(t *testing.T) {
	user := setupWalletFundingTest(t)
	require.NoError(t, CreditUserFunding(user.Id, 200, FundingGift, "invitation", "unique:event", "", 0, ""))
	require.Error(t, CreditUserFunding(user.Id, 300, FundingGift, "invitation", "unique:event", "", 0, ""))
	other := &User{Username: "other-funding-user", Quota: 20}
	require.NoError(t, DB.Create(other).Error)
	require.Error(t, CreditUserFunding(other.Id, 200, FundingGift, "invitation", "unique:event", "", 0, ""))
	require.NoError(t, DB.First(user, user.Id).Error)
	require.NoError(t, DB.First(other, other.Id).Error)
	assert.Equal(t, 300, user.Quota)
	assert.Equal(t, 20, other.Quota)
}

func TestFundingDuplicateTopupRollsBackBalance(t *testing.T) {
	user := setupWalletFundingTest(t)
	order := &TopUp{Id: 1, UserId: user.Id, TradeNo: "paid-order"}
	credit := func(tx *gorm.DB) error { return creditTopUpWallet(tx, order, 200, nil) }
	require.NoError(t, DB.Transaction(credit))
	require.Error(t, DB.Transaction(credit))
	require.NoError(t, DB.First(user, user.Id).Error)
	assert.Equal(t, 300, user.Quota)
	_, total, summary, err := GetWalletFunding(user.Id, 0, 20, "")
	require.NoError(t, err)
	assert.EqualValues(t, 1, total)
	assert.Equal(t, "200", summary.Paid)
}

func TestFundingLedgerFailureRollsBackTopup(t *testing.T) {
	user := setupWalletFundingTest(t)
	require.NoError(t, DB.Callback().Create().Before("gorm:create").Register("funding_test_failure", func(tx *gorm.DB) {
		if tx.Statement.Table == "wallet_funding_entries" {
			tx.AddError(errors.New("ledger unavailable"))
		}
	}))
	err := DB.Transaction(func(tx *gorm.DB) error {
		return creditTopUpWallet(tx, &TopUp{Id: 1, UserId: user.Id, TradeNo: "paid-order"}, 200, nil)
	})
	require.Error(t, err)
	require.NoError(t, DB.First(user, user.Id).Error)
	assert.Equal(t, 100, user.Quota)
}

func TestFundingAdminClassificationAndAdjustments(t *testing.T) {
	user := setupWalletFundingTest(t)
	_, err := AdjustUserWalletFunding(user.Id, 9, "add", 200, FundingPaid, "offline payment")
	require.NoError(t, err)
	_, err = AdjustUserWalletFunding(user.Id, 9, "add", 100, FundingGift, "promotion")
	require.NoError(t, err)
	_, err = AdjustUserWalletFunding(user.Id, 9, "subtract", 50, FundingPaid, "correction")
	require.NoError(t, err)
	before, err := AdjustUserWalletFunding(user.Id, 9, "override", 100, FundingGift, "replace balance")
	require.NoError(t, err)
	assert.Equal(t, 350, before)
	rows, total, summary, err := GetWalletFunding(user.Id, 0, 20, FundingAdjustment)
	require.NoError(t, err)
	assert.EqualValues(t, 2, total)
	assert.Len(t, rows, 2)
	assert.Equal(t, "200", summary.Paid)
	assert.Equal(t, "100", summary.Gift)
	assert.Equal(t, "-300", summary.Adjustment)
	assert.Equal(t, 9, rows[0].OperatorId)
	assert.Equal(t, "replace balance", rows[0].Note)
	_, err = AdjustUserWalletFunding(user.Id, 9, "add", 100, "invalid", "")
	require.Error(t, err)
}

func TestFundingAdjustmentBounds(t *testing.T) {
	user := setupWalletFundingTest(t)
	for _, mode := range []string{"add", "subtract", "override"} {
		if mode == "override" {
			_, err := AdjustUserWalletFunding(user.Id, 9, mode, -1, FundingAdjustment, "")
			require.Error(t, err)
		}
		_, err := AdjustUserWalletFunding(user.Id, 9, mode, -common.MaxWalletQuota-1, FundingAdjustment, "")
		require.Error(t, err)
		_, err = AdjustUserWalletFunding(user.Id, 9, mode, common.MaxWalletQuota+1, FundingAdjustment, "")
		require.Error(t, err)
	}
	require.NoError(t, DB.First(user, user.Id).Error)
	assert.Equal(t, 100, user.Quota)
	_, count, _, err := GetWalletFunding(user.Id, 0, 20, "")
	require.NoError(t, err)
	assert.Zero(t, count)
}

func TestFundingLifetimeTotalsRemainExactBeyondInt64(t *testing.T) {
	user := setupWalletFundingTest(t)
	entries := make([]WalletFundingEntry, 1100)
	for i := range entries {
		entries[i] = WalletFundingEntry{UserId: user.Id, Amount: int64(common.MaxWalletQuota), Kind: FundingPaid, Source: "topup", EventKey: fmt.Sprintf("large-total:%d", i)}
	}
	require.NoError(t, DB.CreateInBatches(&entries, 50).Error)
	_, total, summary, err := GetWalletFunding(user.Id, 0, 20, "")
	require.NoError(t, err)
	assert.EqualValues(t, len(entries), total)
	want := new(big.Int).Mul(big.NewInt(int64(common.MaxWalletQuota)), big.NewInt(int64(len(entries))))
	assert.Equal(t, want.String(), summary.Paid)
}

func TestFundingRedemptionClassificationAndImmutability(t *testing.T) {
	user := setupWalletFundingTest(t)
	for index, kind := range []string{FundingPaid, FundingGift, ""} {
		code := &Redemption{UserId: 9, Key: fmt.Sprintf("funding-redemption-%d", index), Name: "funding", Quota: 50, FundingKind: kind}
		require.NoError(t, code.Insert())
		_, err := Redeem(code.Key, user.Id)
		require.NoError(t, err)
		require.NoError(t, DB.First(code, code.Id).Error)
		code.FundingKind = FundingAdjustment
		require.Error(t, code.Update(), "redeemed credits must retain their original nature")
	}
	_, _, summary, err := GetWalletFunding(user.Id, 0, 20, "")
	require.NoError(t, err)
	assert.Equal(t, "50", summary.Paid)
	assert.Equal(t, "50", summary.Gift)
	assert.Equal(t, "50", summary.Unclassified)
}

func TestFundingRegistrationAndCheckin(t *testing.T) {
	setupWalletFundingTest(t)
	previousGift := common.QuotaForNewUser
	common.QuotaForNewUser = 40
	t.Cleanup(func() { common.QuotaForNewUser = previousGift })
	user := &User{Username: "funding-registration", Status: common.UserStatusEnabled}
	require.NoError(t, DB.Transaction(func(tx *gorm.DB) error { return user.InsertWithTx(tx, 0) }))
	checkin := &Checkin{UserId: user.Id, CheckinDate: "2026-11-12", QuotaAwarded: 10}
	_, err := userCheckinWithTransaction(checkin, user.Id, 10)
	require.NoError(t, err)
	_, err = userCheckinWithTransaction(&Checkin{UserId: user.Id, CheckinDate: "2026-11-12", QuotaAwarded: 10}, user.Id, 10)
	require.Error(t, err)
	require.NoError(t, DB.First(user, user.Id).Error)
	assert.Equal(t, 50, user.Quota)
	_, total, summary, err := GetWalletFunding(user.Id, 0, 20, "")
	require.NoError(t, err)
	assert.EqualValues(t, 2, total)
	assert.Equal(t, "50", summary.Gift)
}

func TestFundingDoesNotInferHistoricalBalance(t *testing.T) {
	user := setupWalletFundingTest(t)
	_, total, summary, err := GetWalletFunding(user.Id, 0, 20, "")
	require.NoError(t, err)
	assert.Zero(t, total)
	assert.Equal(t, "0", summary.Paid)
	assert.Equal(t, "0", summary.Gift)
	require.NoError(t, DB.First(user, user.Id).Error)
	assert.Equal(t, 100, user.Quota)
}
