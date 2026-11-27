package controller

import (
	"bytes"
	"fmt"
	"net/http"
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

func TestManageWalletAdjustmentAppearsInAffectedUsersLogs(t *testing.T) {
	previousDB, previousLogDB := model.DB, model.LOG_DB
	previousRedisEnabled := common.RedisEnabled
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	model.DB, model.LOG_DB = db, db
	common.RedisEnabled = false
	t.Cleanup(func() {
		model.DB, model.LOG_DB = previousDB, previousLogDB
		common.RedisEnabled = previousRedisEnabled
		_ = sqlDB.Close()
	})
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.WalletFundingEntry{}, &model.Log{}))
	user := model.User{Username: "adjusted-user", Role: common.RoleCommonUser, Quota: 100, AffCode: "adj-user"}
	admin := model.User{Username: "adjusting-admin", Role: common.RoleAdminUser, AffCode: "adj-admin"}
	require.NoError(t, db.Create(&user).Error)
	require.NoError(t, db.Create(&admin).Error)

	for _, tc := range []struct {
		mode    string
		value   int
		success bool
	}{
		{"add", 200, true},
		{"subtract", 50, true},
		{"override", 0, true},
		{"add", 0, false},
	} {
		body, err := common.Marshal(ManageRequest{
			Id: user.Id, Action: "add_quota", Mode: tc.mode, Value: tc.value,
			FundingKind: model.FundingGift, Note: "private adjustment note",
		})
		require.NoError(t, err)
		response := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(response)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/user/manage", bytes.NewReader(body))
		c.Set("id", admin.Id)
		c.Set("username", admin.Username)
		c.Set("role", admin.Role)
		ManageUser(c)
		var result struct {
			Success bool `json:"success"`
		}
		require.NoError(t, common.Unmarshal(response.Body.Bytes(), &result))
		require.Equal(t, tc.success, result.Success, response.Body.String())
	}

	logs, total, err := model.GetUserLogs(user.Id, model.LogTypeSystem, 0, 0, "", "", 0, 10, "", "", "")
	require.NoError(t, err)
	require.EqualValues(t, 3, total)
	require.Len(t, logs, 3)
	for i, mode := range []string{"override", "subtract", "add"} {
		require.Equal(t, user.Username, logs[i].Username)
		require.Zero(t, logs[i].Quota)
		require.Contains(t, logs[i].Other, "user.quota_"+mode)
		require.NotContains(t, logs[i].Other, "private adjustment note")
		require.NotContains(t, logs[i].Other, admin.Username)
		require.Empty(t, logs[i].Ip)
	}

	_, total, err = model.GetUserLogs(user.Id, model.LogTypeSystem, 0, 0, "", "", 0, 10, "", "", "", 123)
	require.NoError(t, err)
	require.Zero(t, total)
	_, total, err = model.GetAllLogs(model.LogTypeSystem, 0, 0, "", user.Username, "", 0, 10, 0, "", "", "")
	require.NoError(t, err)
	require.EqualValues(t, 3, total)
	_, total, err = model.GetAllLogs(model.LogTypeManage, 0, 0, "", admin.Username, "", 0, 10, 0, "", "", "")
	require.NoError(t, err)
	require.EqualValues(t, 3, total)
}

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
	user := model.User{Username: "funding-private-user", Role: common.RoleCommonUser, AffCode: "fund-user"}
	admin := model.User{Username: "funding-private-admin", Role: common.RoleAdminUser, AffCode: "fund-adm"}
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

func TestWalletFundingAdminShowsRecordedOrderPayments(t *testing.T) {
	previousDB := model.DB
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	conn, err := db.DB()
	require.NoError(t, err)
	conn.SetMaxOpenConns(1)
	model.DB = db
	t.Cleanup(func() {
		model.DB = previousDB
		_ = conn.Close()
	})
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.WalletFundingEntry{}, &model.TopUp{}))
	user := model.User{Username: "paid-ledger-user", Role: common.RoleCommonUser, AffCode: "pay-u"}
	other := model.User{Username: "paid-ledger-other", Role: common.RoleCommonUser, AffCode: "pay-o"}
	require.NoError(t, db.Create(&user).Error)
	require.NoError(t, db.Create(&other).Error)

	entries := []struct {
		reference  string
		kind       string
		source     string
		credit     int64
		order      *model.TopUp
		paidAmount string
	}{
		{reference: "gift", kind: model.FundingGift, order: &model.TopUp{Money: 80}},
		{reference: "manual", source: "admin", order: &model.TopUp{Money: 79}},
		{reference: "discount", credit: 500_000_000, order: &model.TopUp{Money: 950, PaymentProvider: model.PaymentProviderEpay}, paidAmount: "950.00"},
		{reference: "zero", order: &model.TopUp{}, paidAmount: "0.00"},
		{reference: "rounded", order: &model.TopUp{Money: 12.125}, paidAmount: "12.12"},
		{reference: "missing"},
		{reference: "other-user", order: &model.TopUp{UserId: other.Id, Money: 75}},
		{reference: "organization", order: &model.TopUp{OrganizationId: 9, Money: 74}},
		{reference: "pending", order: &model.TopUp{Status: common.TopUpStatusPending, Money: 73}},
		{reference: "stripe", order: &model.TopUp{PaymentMethod: model.PaymentMethodStripe, Money: 72}},
	}
	for index, test := range entries {
		entry := model.WalletFundingEntry{
			UserId: user.Id, Amount: test.credit, Kind: test.kind, Source: test.source,
			Reference: test.reference, EventKey: "test:" + test.reference, CreatedAt: int64(len(entries) - index),
		}
		if entry.Kind == "" {
			entry.Kind = model.FundingPaid
		}
		if entry.Source == "" {
			entry.Source = "topup"
		}
		if entry.Amount == 0 {
			entry.Amount = 50_000_000
		}
		if order := test.order; order != nil {
			order.TradeNo = entry.Reference
			if order.UserId == 0 {
				order.UserId = user.Id
			}
			if order.Status == "" {
				order.Status = common.TopUpStatusSuccess
			}
			if order.PaymentMethod == "" {
				order.PaymentMethod = "alipay"
			}
			require.NoError(t, db.Create(order).Error)
			if entry.Source == "topup" && entry.Kind == model.FundingPaid {
				entry.EventKey = fmt.Sprintf("topup:%d", order.Id)
			}
		}
		require.NoError(t, model.RecordWalletFundingTx(db, entry))
	}

	router := gin.New()
	router.GET("/api/user/self/wallet/funding", func(c *gin.Context) {
		c.Set("id", user.Id)
		GetSelfWalletFunding(c)
	})
	router.GET("/api/user/:id/wallet/funding", func(c *gin.Context) {
		c.Set("role", common.RoleAdminUser)
		GetUserWalletFunding(c)
	})
	type ledgerPage struct {
		Items   []map[string]any           `json:"items"`
		Total   int                        `json:"total"`
		Summary model.WalletFundingSummary `json:"summary"`
	}
	getLedger := func(query string) ledgerPage {
		response := httptest.NewRecorder()
		path := fmt.Sprintf("/api/user/%d/wallet/funding%s", user.Id, query)
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		var result struct {
			Success bool       `json:"success"`
			Data    ledgerPage `json:"data"`
		}
		require.NoError(t, common.Unmarshal(response.Body.Bytes(), &result))
		require.True(t, result.Success, response.Body.String())
		return result.Data
	}

	all := getLedger("?page_size=20")
	require.Len(t, all.Items, len(entries))
	assert.Equal(t, len(entries), all.Total)
	for index, test := range entries {
		item := all.Items[index]
		assert.Equal(t, test.reference, item["reference"])
		if test.paidAmount == "" {
			assert.NotContains(t, item, "paid_amount", test.reference)
		} else {
			assert.Equal(t, test.paidAmount, item["paid_amount"], test.reference)
		}
	}
	filtered := getLedger("?kind=paid&p=2&page_size=1")
	require.Len(t, filtered.Items, 1)
	assert.Equal(t, len(entries)-1, filtered.Total)
	assert.Equal(t, all.Summary, filtered.Summary)
	assert.Equal(t, "discount", filtered.Items[0]["reference"])
	assert.Equal(t, "500000000", filtered.Items[0]["amount"])
	assert.Equal(t, "950.00", filtered.Items[0]["paid_amount"])

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/user/self/wallet/funding", nil))
	assert.Contains(t, response.Body.String(), `"paid":"900000000"`)
	assert.NotContains(t, response.Body.String(), "paid_amount")
}
