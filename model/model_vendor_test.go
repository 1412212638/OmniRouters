package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestSetModelVendorIDsDeduplicatesAndReplaces(t *testing.T) {
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
	if err := DB.AutoMigrate(&Model{}, &Vendor{}, &ModelVendor{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := DB.Create(&Model{ModelName: "deepseek-v4.1-flash", Status: 1, SyncOfficial: 1}).Error; err != nil {
		t.Fatalf("create model: %v", err)
	}
	if err := DB.Create(&[]Vendor{{Name: "OpenAI"}, {Name: "Alibaba"}, {Name: "Tencent"}}).Error; err != nil {
		t.Fatalf("create vendors: %v", err)
	}
	var m Model
	if err := DB.Where("model_name = ?", "deepseek-v4.1-flash").First(&m).Error; err != nil {
		t.Fatalf("load model: %v", err)
	}
	if err := SetModelVendorIDs(m.Id, []int{3, 1, 2, 2, 1}); err != nil {
		t.Fatalf("set vendors: %v", err)
	}
	got, err := GetModelVendorIDs(m.Id)
	if err != nil {
		t.Fatalf("get vendors: %v", err)
	}
	if len(got) != 3 || got[0] != 3 || got[1] != 1 || got[2] != 2 {
		t.Fatalf("unexpected vendors: %#v", got)
	}
	if err := SetModelVendorIDs(m.Id, []int{3}); err != nil {
		t.Fatalf("replace vendors: %v", err)
	}
	got, err = GetModelVendorIDs(m.Id)
	if err != nil {
		t.Fatalf("get replaced vendors: %v", err)
	}
	if len(got) != 1 || got[0] != 3 {
		t.Fatalf("unexpected replaced vendors: %#v", got)
	}
}
