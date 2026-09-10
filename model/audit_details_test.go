package model

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestAuditOutcomeMigrationPreservesHistoricalRows(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	raw, err := db.DB(); require.NoError(t, err); defer raw.Close()
	type legacyAudit struct {
		Id int `gorm:"primaryKey"`
		UserId int
		Category string
		Action string
		Success bool
	}
	require.NoError(t, db.Table("audit_logs").AutoMigrate(&legacyAudit{}))
	require.NoError(t, db.Table("audit_logs").Create(&legacyAudit{Id: 1, UserId: 7, Category: "access_token", Action: "request", Success: true}).Error)
	require.NoError(t, db.AutoMigrate(&AuditLog{}))
	var row AuditLog
	require.NoError(t, db.First(&row, 1).Error)
	require.Equal(t, 7, row.UserId)
	require.True(t, row.Success)
	require.Empty(t, row.Outcome)
	require.NoError(t, db.AutoMigrate(&AuditLog{}))
}
