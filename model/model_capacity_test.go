package model

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestModelCapacitySaveAndClear(t *testing.T) {
	original := DB
	t.Cleanup(func() { DB = original })
	var err error
	DB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil { t.Fatal(err) }
	if err = DB.AutoMigrate(&Model{}); err != nil { t.Fatal(err) }
	m := Model{ModelName: "capacity-test", ContextLength: 1048576, MaxOutputTokens: 32768}
	m.ContextLengthDisplay, m.MaxOutputTokensDisplay = "1.05M", "128.00K"
	if err = m.Insert(); err != nil { t.Fatal(err) }
	var saved Model
	if err = DB.First(&saved, m.Id).Error; err != nil { t.Fatal(err) }
	if saved.ContextLength != m.ContextLength || saved.MaxOutputTokens != m.MaxOutputTokens {
		t.Fatal("model capacities were not persisted")
	}
	if saved.ContextLengthDisplay != "1.05M" || saved.MaxOutputTokensDisplay != "128.00K" {
		t.Fatal("capacity display text must be preserved verbatim")
	}
	m.ContextLength, m.MaxOutputTokens = 0, 0
	m.ContextLengthDisplay, m.MaxOutputTokensDisplay = "", ""
	if err = m.Update(); err != nil { t.Fatal(err) }
	if err = DB.First(&saved, m.Id).Error; err != nil { t.Fatal(err) }
	if saved.ContextLength != 0 || saved.MaxOutputTokens != 0 {
		t.Fatal("clearing model capacities must persist zero values")
	}
	if saved.ContextLengthDisplay != "" || saved.MaxOutputTokensDisplay != "" {
		t.Fatal("clearing display text must persist empty strings")
	}
}
