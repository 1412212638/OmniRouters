package model

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestVisitBillingExportLogsPagesByTimeAndID(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)

	previousLogDB, previousType := LOG_DB, common.LogDatabaseType()
	LOG_DB = db
	common.SetLogDatabaseType(common.DatabaseTypeSQLite)
	t.Cleanup(func() {
		LOG_DB = previousLogDB
		common.SetLogDatabaseType(previousType)
		require.NoError(t, sqlDB.Close())
	})
	require.NoError(t, db.AutoMigrate(&Log{}))

	logs := make([]Log, billingExportLogBatchSize+1)
	for i := range logs {
		logs[i] = Log{UserId: 1, CreatedAt: 101, Type: LogTypeConsume, Quota: 1}
	}
	logs = append(logs,
		Log{UserId: 1, CreatedAt: 100, Type: LogTypeRefund, Quota: 2},
		Log{UserId: 1, CreatedAt: 102, Type: LogTypeConsume, Quota: 3},
		Log{UserId: 1, CreatedAt: 103, Type: LogTypeConsume},
		Log{UserId: 1, CreatedAt: 99, Type: LogTypeConsume},
		Log{UserId: 1, CreatedAt: 101, Type: LogTypeError},
		Log{UserId: 2, CreatedAt: 101, Type: LogTypeConsume},
		Log{UserId: 1, OrganizationId: 8, CreatedAt: 101, Type: LogTypeConsume, Quota: 4},
		Log{UserId: 2, OrganizationId: 8, CreatedAt: 101, Type: LogTypeRefund, Quota: 5},
	)
	require.NoError(t, db.CreateInBatches(&logs, 100).Error)
	require.NoError(t, db.Model(&Log{}).Where("id = ?", logs[0].Id).Update("organization_id", nil).Error)

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	var got []Log
	var batchSizes []int
	err = VisitBillingExportLogs(ctx, BillingExport{UserID: 1}, 100, 103, func(batch []Log) error {
		if len(got) == 0 {
			// Lookup/callback work can safely use the same SQLite connection.
			// This backdated insertion is outside the original read boundary.
			require.NoError(t, db.WithContext(ctx).Create(&Log{UserId: 1, CreatedAt: 102, Type: LogTypeConsume}).Error)
		}
		got = append(got, batch...)
		batchSizes = append(batchSizes, len(batch))
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, []int{billingExportLogBatchSize, 3}, batchSizes)
	require.Equal(t, LogTypeRefund, got[0].Type)
	for i := range billingExportLogBatchSize + 1 {
		require.Equal(t, logs[i].Id, got[i+1].Id)
	}
	require.Equal(t, 3, got[len(got)-1].Quota)

	visitorStopped := errors.New("visitor stopped")
	err = VisitBillingExportLogs(ctx, BillingExport{UserID: 1, OrganizationID: 8}, 100, 103, func(batch []Log) error {
		require.Len(t, batch, 2)
		require.Equal(t, 4, batch[0].Quota)
		require.Equal(t, 5, batch[1].Quota)
		return visitorStopped
	})
	require.ErrorIs(t, err, visitorStopped)
}
