package model

import (
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestInvoiceProfileValidateNormalizesAndRequiresIdentity(t *testing.T) {
	profile := &InvoiceProfile{
		Type:  "PERSONAL ",
		Name:  "  Ada Lovelace ",
		Email: " Ada@Example.com ",
	}

	require.NoError(t, profile.Validate())
	require.Equal(t, "personal", profile.Type)
	require.Equal(t, "Ada Lovelace", profile.Name)
	require.Equal(t, "ada@example.com", profile.Email)

	require.Error(t, (&InvoiceProfile{Type: "personal", Email: "ada@example.com"}).Validate())
	require.Error(t, (&InvoiceProfile{Type: "company", CompanyName: "Example", Email: "invalid"}).Validate())
	require.Error(t, (&InvoiceProfile{Type: "company", CompanyName: "Example", Email: "billing@example.com", Country: "USA"}).Validate())
}

func TestInvoiceProfileValidateSupportsCompanyDetails(t *testing.T) {
	profile := &InvoiceProfile{
		Type:        "company",
		CompanyName: "Example Ltd",
		TaxID:       "GB123456789",
		Email:       "billing@example.com",
		Country:     "gb",
	}

	require.NoError(t, profile.Validate())
	require.Equal(t, "GB", profile.Country)
}

func TestInvoiceProfileValidateCountsUnicodeCharacters(t *testing.T) {
	profile := &InvoiceProfile{
		Type:  "personal",
		Name:  strings.Repeat("中", 200),
		Email: "billing@example.com",
	}
	require.NoError(t, profile.Validate())

	profile.Name += "中"
	require.Error(t, profile.Validate())
}

func TestInvoiceProfileAndStripeCustomerPersist(t *testing.T) {
	originalDB := DB
	originalMainType := common.MainDatabaseType()
	originalLogType := common.LogDatabaseType()
	originalRedisEnabled := common.RedisEnabled
	t.Cleanup(func() {
		DB = originalDB
		common.SetDatabaseTypes(originalMainType, originalLogType)
		common.RedisEnabled = originalRedisEnabled
	})
	common.SetDatabaseTypes(common.DatabaseTypeSQLite, common.DatabaseTypeSQLite)
	common.RedisEnabled = false
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	DB = db
	require.NoError(t, DB.AutoMigrate(&User{}))

	user := &User{Username: "invoice-profile-test"}
	require.NoError(t, DB.Create(user).Error)
	profile := &InvoiceProfile{
		Type:    "personal",
		Name:    strings.Repeat("中", 100),
		Email:   "ada@example.com",
		City:    "北京市",
		Address: "朝阳区建国路",
	}
	require.NoError(t, SetUserInvoiceProfile(user.Id, profile))
	require.NoError(t, SetUserStripeCustomer(user.Id, "cus_invoice_profile"))

	loaded, err := GetUserById(user.Id, true)
	require.NoError(t, err)
	require.Equal(t, "cus_invoice_profile", loaded.StripeCustomer)
	require.Equal(t, profile, loaded.GetInvoiceProfile())

	require.NoError(t, SetUserInvoiceProfile(user.Id, nil))
	loaded, err = GetUserById(user.Id, true)
	require.NoError(t, err)
	require.Nil(t, loaded.GetInvoiceProfile())
}
