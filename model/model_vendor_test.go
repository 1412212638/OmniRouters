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

func TestGetEnabledVendorIDsAndDeleteBindings(t *testing.T) {
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
	vendors := []Vendor{{Name: "Enabled", Status: 1}, {Name: "Disabled", Status: 0}}
	if err := DB.Create(&vendors).Error; err != nil {
		t.Fatalf("create vendors: %v", err)
	}
	m := Model{ModelName: "test-model", Status: 1}
	if err := DB.Create(&m).Error; err != nil {
		t.Fatalf("create model: %v", err)
	}
	if err := DB.Create(&[]ModelVendor{{ModelID: m.Id, VendorID: vendors[0].Id}, {ModelID: m.Id, VendorID: vendors[1].Id}}).Error; err != nil {
		t.Fatalf("create bindings: %v", err)
	}
	enabled, err := GetEnabledVendorIDs([]int{vendors[0].Id, vendors[1].Id})
	if err != nil {
		t.Fatalf("get enabled vendors: %v", err)
	}
	if _, ok := enabled[vendors[0].Id]; !ok {
		t.Fatal("enabled vendor was filtered out")
	}
	if _, ok := enabled[vendors[1].Id]; ok {
		t.Fatal("disabled vendor was returned")
	}
	if err := DeleteModelVendorBindingsByVendorID(vendors[1].Id); err != nil {
		t.Fatalf("delete vendor bindings: %v", err)
	}
	var count int64
	if err := DB.Model(&ModelVendor{}).Where("vendor_id = ?", vendors[1].Id).Count(&count).Error; err != nil {
		t.Fatalf("count bindings: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected deleted bindings, got %d", count)
	}
}
