package model

import (
	"path/filepath"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestResetChannelUsedQuota(t *testing.T) {
	oldDB, oldBatchUpdateEnabled := DB, common.BatchUpdateEnabled
	databasePath := filepath.Join(t.TempDir(), "channels.sqlite")
	openDatabase := func() *gorm.DB {
		db, err := gorm.Open(sqlite.Open(databasePath), &gorm.Config{})
		require.NoError(t, err)
		sqlDB, err := db.DB()
		require.NoError(t, err)
		t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
		return db
	}
	writerDB, adminDB := openDatabase(), openDatabase()
	DB = writerDB
	common.BatchUpdateEnabled = true
	drainChannelUsedQuotaBatch()
	t.Cleanup(func() {
		drainChannelUsedQuotaBatch()
		DB, common.BatchUpdateEnabled = oldDB, oldBatchUpdateEnabled
	})
	require.NoError(t, DB.AutoMigrate(&Channel{}, &Ability{}))
	channel := Channel{Id: 1, Key: "test-key", UsedQuota: 500, Balance: 12.5, BalanceUpdatedTime: 123, Models: "test-model", Group: "default"}
	other := Channel{Id: 2, Key: "other-key", UsedQuota: 750}
	require.NoError(t, DB.Create(&channel).Error)
	require.NoError(t, DB.Create(&other).Error)

	UpdateChannelUsedQuota(channel.Id, 80)
	UpdateChannelUsedQuota(other.Id, 120)
	batchTakenBeforeReset := drainChannelUsedQuotaBatch()
	UpdateChannelUsedQuota(channel.Id, 15)
	DB = adminDB
	require.NoError(t, ResetChannelUsedQuota(channel.Id))
	// Confirming a reset on an already zero counter also supersedes old deltas.
	require.NoError(t, ResetChannelUsedQuota(channel.Id))
	require.ErrorIs(t, ResetChannelUsedQuota(99), gorm.ErrRecordNotFound)
	DB = writerDB
	flushChannelUsedQuotaBatch(batchTakenBeforeReset)
	flushChannelUsedQuotaBatch(drainChannelUsedQuotaBatch())

	// A channel edit loaded before the reset must not restore either old field.
	channel.Name = "renamed"
	require.NoError(t, channel.Update())
	var got Channel
	require.NoError(t, DB.First(&got, channel.Id).Error)
	require.Equal(t, "renamed", got.Name)
	require.Zero(t, got.UsedQuota)
	require.EqualValues(t, 2, got.UsedQuotaVersion)
	require.Equal(t, 12.5, got.Balance)
	require.EqualValues(t, 123, got.BalanceUpdatedTime)
	require.NoError(t, DB.First(&other, other.Id).Error)
	require.EqualValues(t, 870, other.UsedQuota)

	UpdateChannelUsedQuota(channel.Id, 700)
	UpdateChannelUsedQuota(channel.Id, -50)
	batchUpdate()
	common.BatchUpdateEnabled = false
	UpdateChannelUsedQuota(channel.Id, 20)
	require.NoError(t, DB.First(&got, channel.Id).Error)
	require.EqualValues(t, 670, got.UsedQuota)
}
