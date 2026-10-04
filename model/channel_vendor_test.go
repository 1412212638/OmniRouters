package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestChannelUpdatePersistsVendorBinding(t *testing.T) {
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
	if err := DB.AutoMigrate(&Channel{}, &Ability{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := DB.Create(&Channel{
		Name:   "test-channel",
		Key:    "test-key",
		Models: "glm-5.3",
		Group:  "default",
		Status: common.ChannelStatusEnabled,
	}).Error; err != nil {
		t.Fatalf("create channel: %v", err)
	}
	var channel Channel
	if err := DB.First(&channel).Error; err != nil {
		t.Fatalf("load channel: %v", err)
	}
	vendorID := 42
	channel.VendorID = &vendorID
	if err := channel.Update(); err != nil {
		t.Fatalf("update channel: %v", err)
	}

	var reloaded Channel
	if err := DB.First(&reloaded, channel.Id).Error; err != nil {
		t.Fatalf("reload channel: %v", err)
	}
	if reloaded.VendorID == nil || *reloaded.VendorID != vendorID {
		t.Fatalf("vendor binding was not persisted: %#v", reloaded.VendorID)
	}
}
