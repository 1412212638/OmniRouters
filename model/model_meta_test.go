package model

import (
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestIsModelNameDuplicatedIsCaseSensitive(t *testing.T) {
	originalDB := DB
	originalMainType := common.MainDatabaseType()
	originalLogType := common.LogDatabaseType()
	t.Cleanup(func() {
		DB = originalDB
		common.SetDatabaseTypes(originalMainType, originalLogType)
	})

	common.SetDatabaseTypes(common.DatabaseTypeSQLite, common.DatabaseTypeSQLite)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	DB = db
	if err := DB.AutoMigrate(&Model{}); err != nil {
		t.Fatalf("migrate models: %v", err)
	}

	if err := (&Model{ModelName: "GPT-image-2", Status: 1, SyncOfficial: 1}).Insert(); err != nil {
		t.Fatalf("insert model: %v", err)
	}

	duplicated, err := IsModelNameDuplicated(0, "GPT-image-2")
	if err != nil {
		t.Fatalf("check exact duplicate: %v", err)
	}
	if !duplicated {
		t.Fatal("expected exact model name to be duplicated")
	}

	duplicated, err = IsModelNameDuplicated(0, "gpt-image-2")
	if err != nil {
		t.Fatalf("check case-variant duplicate: %v", err)
	}
	if duplicated {
		t.Fatal("expected case-variant model name to be allowed")
	}
}

func TestWhereModelNameExactUsesBinaryComparisonOnMySQL(t *testing.T) {
	originalMainType := common.MainDatabaseType()
	originalLogType := common.LogDatabaseType()
	t.Cleanup(func() {
		common.SetDatabaseTypes(originalMainType, originalLogType)
	})

	common.SetDatabaseTypes(common.DatabaseTypeMySQL, common.DatabaseTypeSQLite)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{DryRun: true})
	if err != nil {
		t.Fatalf("open dry-run db: %v", err)
	}

	stmt := WhereModelNameExact(db.Model(&Model{}), "GPT-image-2").
		Find(&[]Model{}).
		Statement
	if sql := stmt.SQL.String(); !strings.Contains(sql, "BINARY model_name = BINARY ?") {
		t.Fatalf("expected MySQL binary comparison, got SQL: %s", sql)
	}
}
