package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestGetEnabledChannelVendorIDsByModelsMap(t *testing.T) {
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
	vendorA, vendorB, vendorDisabled := 11, 12, 13
	channels := []Channel{
		{Id: 1, Name: "a", VendorID: &vendorA, Status: common.ChannelStatusEnabled},
		{Id: 2, Name: "b", VendorID: &vendorB, Status: common.ChannelStatusEnabled},
		{Id: 3, Name: "disabled", VendorID: &vendorDisabled, Status: common.ChannelStatusManuallyDisabled},
	}
	if err := DB.Create(&channels).Error; err != nil {
		t.Fatalf("create channels: %v", err)
	}
	abilities := []Ability{
		{Group: "default", Model: "glm-5.3", ChannelId: 1, Enabled: true},
		{Group: "default", Model: "glm-5.3", ChannelId: 2, Enabled: true},
		{Group: "default", Model: "glm-5.3", ChannelId: 3, Enabled: true},
		{Group: "default", Model: "other-model", ChannelId: 1, Enabled: true},
	}
	if err := DB.Create(&abilities).Error; err != nil {
		t.Fatalf("create abilities: %v", err)
	}

	got, err := GetEnabledChannelVendorIDsByModelsMap([]string{"glm-5.3"})
	if err != nil {
		t.Fatalf("query channel vendors: %v", err)
	}
	want := []int{vendorA, vendorB}
	if len(got["glm-5.3"]) != len(want) {
		t.Fatalf("unexpected vendors: %#v", got["glm-5.3"])
	}
	for i, vendorID := range want {
		if got["glm-5.3"][i] != vendorID {
			t.Fatalf("unexpected vendor order: %#v", got["glm-5.3"])
		}
	}
}
