package controller

import (
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestWalletFundingSelfPrivacyAndAdminRoleBoundary(t *testing.T) {
	previousDB := model.DB
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	model.DB = db
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.WalletFundingEntry{}))
	t.Cleanup(func() {
		model.DB = previousDB
		connection, _ := db.DB()
		_ = connection.Close()
	})
	user := model.User{Username: "funding-private-user", Role: common.RoleCommonUser}
	admin := model.User{Username: "funding-private-admin", Role: common.RoleAdminUser}
	require.NoError(t, db.Create(&user).Error)
	require.NoError(t, db.Create(&admin).Error)
	require.NoError(t, model.RecordWalletFundingTx(db, model.WalletFundingEntry{
		UserId: user.Id, Amount: 100, Kind: model.FundingPaid, Source: "admin",
		EventKey: "privacy:test", Note: "private internal note", Reference: "private-order", OperatorId: admin.Id,
	}))

	router := gin.New()
	router.GET("/self", func(c *gin.Context) {
		c.Set("id", user.Id)
		GetSelfWalletFunding(c)
	})
	router.GET("/admin/:id", func(c *gin.Context) {
		c.Set("role", common.RoleAdminUser)
		GetUserWalletFunding(c)
	})
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest("GET", "/self?user_id=999", nil))
	assert.Contains(t, response.Body.String(), `"paid":"100"`)
	assert.NotContains(t, response.Body.String(), "private")
	assert.NotContains(t, response.Body.String(), "operator_id")

	response = httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest("GET", "/admin/"+common.Interface2String(user.Id), nil))
	assert.Contains(t, response.Body.String(), "private internal note")
	response = httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest("GET", "/admin/"+common.Interface2String(admin.Id), nil))
	assert.Contains(t, response.Body.String(), `"success":false`)
	assert.NotContains(t, response.Body.String(), "items")
}
