package service

import (
	"archive/zip"
	"bytes"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/pkg/billexport"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/glebarez/sqlite"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestBillingExportLanguagePipeline(t *testing.T) {
	settings := operation_setting.GetGeneralSetting()
	previousSettings, previousRate, previousQuota := *settings, operation_setting.USDExchangeRate, common.QuotaPerUnit
	t.Cleanup(func() {
		*settings, operation_setting.USDExchangeRate, common.QuotaPerUnit = previousSettings, previousRate, previousQuota
	})
	settings.QuotaDisplayType, operation_setting.USDExchangeRate, common.QuotaPerUnit = operation_setting.QuotaDisplayTypeCNY, 7.3, 500_000
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	originalDB, originalLogDB := model.DB, model.LOG_DB
	model.DB, model.LOG_DB = db, db
	t.Cleanup(func() {
		model.DB, model.LOG_DB = originalDB, originalLogDB
		require.NoError(t, sqlDB.Close())
	})
	require.NoError(t, db.AutoMigrate(
		&model.BillingExport{},
		&model.SystemTask{},
		&model.SystemTaskLock{},
		&model.Log{},
		&model.Task{},
		&model.TopUp{},
		&model.SubscriptionOrder{},
		&model.WalletFundingEntry{},
		&model.OrganizationWalletTransaction{},
	))

	ctx := t.Context()
	now := time.Now().In(billexport.Timezone)
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, billexport.Timezone).AddDate(0, -1, 0)
	month := periodStart.Format("2006-01")
	scope := model.BillingExport{
		UserID: 1, StartMonth: month, EndMonth: month, Language: "unsupported",
	}
	require.ErrorIs(t, EnqueueBillingExport(ctx, scope), ErrBillingExportLanguage)
	scope.Language = model.BillingExportLanguageChinese
	require.NoError(t, EnqueueBillingExport(ctx, scope))
	// A settings change after enqueue must affect new reports only.
	settings.QuotaDisplayType, settings.CustomCurrencySymbol, settings.CustomCurrencyExchangeRate = operation_setting.QuotaDisplayTypeCustom, " HK$ ", 7.8
	common.QuotaPerUnit = 1_000_000

	postedAt := periodStart.Add(time.Hour).Unix()
	require.NoError(t, model.LOG_DB.Create(&[]model.Log{
		{
			UserId: 1, CreatedAt: postedAt, Type: model.LogTypeConsume, ModelName: "video",
			Quota: 500000, TokenId: 22, TokenName: "video key", Other: `{"is_task":true,"task_id":"task_precharged","billing_source":"wallet"}`,
		},
		{
			UserId: 1, CreatedAt: postedAt + 120, Type: model.LogTypeRefund, ModelName: "video",
			Quota: 500000, Other: `{"task_id":"task_precharged","billing_source":"wallet","pre_consumed_quota":500000,"actual_quota":0}`,
		},
		{
			UserId: 1, CreatedAt: postedAt + 240, Type: model.LogTypeConsume, ModelName: "chat",
			Quota: 250000, UseTime: 3,
		},
		{
			UserId: 1, CreatedAt: postedAt + 250, Type: model.LogTypeConsume, ModelName: "token-plan",
			Quota: 2500, Other: `{"billing_source":"subscription","subscription_billing_mode":"token"}`,
		},
		{
			UserId: 2, CreatedAt: postedAt + 240, Type: model.LogTypeConsume, ModelName: "other-account",
			Quota: 1000000,
		},
	}).Error)
	list, err := model.ListBillingExports(ctx, scope.UserID, 0, now.Unix())
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.Equal(t, model.BillingExportLanguageChinese, list[0].Language)
	require.Equal(t, model.SystemTaskStatusPending, list[0].Status)

	report, err := model.GetBillingExport(ctx, list[0].TaskID)
	require.NoError(t, err)
	statement, err := loadBillingStatement(ctx, report)
	require.NoError(t, err)
	require.Equal(t, model.BillingExportLanguageChinese, statement.Language)
	require.Equal(t, "个人账户 #1", statement.Account)
	require.Equal(t, billexport.UnitCNY, statement.DisplayUnit)
	require.Equal(t, "7.3", statement.ExchangeRate)
	require.Len(t, statement.Events, 4)
	precharge, refund, chat := statement.Events[0], statement.Events[1], statement.Events[2]
	require.Equal(t, billexport.Precharge, precharge.Kind)
	require.Equal(t, billexport.Refund, refund.Kind)
	require.True(t, refund.Amount.Equal(decimal.RequireFromString("-7.3")))
	require.True(t, precharge.Amount.Add(refund.Amount).IsZero())
	require.Equal(t, 1, precharge.Requests+refund.Requests)
	require.Equal(t, 22, refund.KeyID)
	require.Equal(t, "video key", refund.KeyName)
	require.Equal(t, chat.PostedAt-3, chat.UsedAt)
	require.Equal(t, chat.PostedAt, chat.CompletedAt)
	require.True(t, chat.Amount.Equal(decimal.RequireFromString("3.65")))
	require.Equal(t, billexport.UnitTokens, statement.Events[3].Unit)
	require.True(t, statement.Events[3].Amount.Equal(decimal.NewFromInt(2500)))

	task, err := model.GetSystemTaskByTaskID(report.TaskID)
	require.NoError(t, err)
	const runnerID = "billing-export-test"
	_, claimed, err := model.ClaimSystemTask(task.ID, task.Type, runnerID, now.Add(time.Minute).Unix())
	require.NoError(t, err)
	require.True(t, claimed)
	require.NoError(t, runBillingExport(ctx, task.TaskID, runnerID))
	archive, err := model.ReadBillingExportArchive(ctx, task.TaskID, scope.UserID, 0, now.Unix())
	require.NoError(t, err)
	zipReader, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	require.NoError(t, err)
	require.Len(t, zipReader.File, 4)
	_, err = model.ReadBillingExportArchive(ctx, task.TaskID, 2, 0, now.Unix())
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)

	// An older caller can still omit the field. Historical empty values also
	// render and list as English.
	scope.Language = ""
	require.NoError(t, EnqueueBillingExport(ctx, scope))
	var legacy model.BillingExport
	require.NoError(t, db.Where("task_id <> ?", report.TaskID).Take(&legacy).Error)
	require.Equal(t, model.BillingExportLanguageEnglish, legacy.Language)
	require.NoError(t, db.Model(&legacy).Update("language", "").Error)
	list, err = model.ListBillingExports(ctx, scope.UserID, 0, now.Unix())
	require.NoError(t, err)
	require.Len(t, list, 2)
	for _, item := range list {
		if item.TaskID == legacy.TaskID {
			require.Equal(t, model.BillingExportLanguageEnglish, item.Language)
		}
	}
	legacy.Language = ""
	statement, err = loadBillingStatement(ctx, legacy)
	require.NoError(t, err)
	require.Equal(t, model.BillingExportLanguageEnglish, statement.Language)
	require.Equal(t, "Personal account #1", statement.Account)
	require.Equal(t, billexport.UnitCustomCurrency, statement.DisplayUnit)
	require.Equal(t, "HK$", statement.CurrencySymbol)
	require.True(t, statement.Events[0].Amount.Equal(decimal.RequireFromString("3.9")))

	// Raw history can exceed the former row limit when it collapses into a
	// small set of daily billing groups for the selected account.
	const requests = 28_001
	frequent := make([]model.Log, requests)
	for i := range frequent {
		frequent[i] = model.Log{
			UserId: 1, CreatedAt: postedAt + 500, Type: model.LogTypeConsume,
			ModelName: "frequent", Quota: 250, PromptTokens: 100, CompletionTokens: 10,
		}
	}
	require.NoError(t, model.LOG_DB.CreateInBatches(&frequent, 150).Error)
	statement, err = loadBillingStatement(ctx, legacy)
	require.NoError(t, err)
	require.Len(t, statement.Events, 5)
	group := statement.Events[4]
	require.Equal(t, "frequent", group.Model)
	require.Equal(t, requests, group.RecordCount())
	require.Equal(t, requests, group.Requests)
	require.EqualValues(t, requests*100, group.InputTokens)
	require.True(t, group.Amount.Equal(decimal.RequireFromString("54.60195")))
	archive, err = billexport.WriteArchive(statement)
	require.NoError(t, err)
	require.Less(t, len(archive), model.MaxBillingExportBytes)
}

func TestBillingExportLegacyQuotaUnit(t *testing.T) {
	display := resolveBillingQuotaDisplay(model.BillingExportDisplay{
		Mode: operation_setting.QuotaDisplayTypeTokens, QuotaPerUnit: 500_000, Rate: 1,
	})
	require.Equal(t, billexport.UnitQuota, display.unit)
	require.NotEqual(t, billexport.UnitTokens, display.unit)
	require.True(t, display.amount(1234).Equal(decimal.NewFromInt(1234)))
}

func TestAuthorizeBillingExportUserRequiresAdmin(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	previousDB := model.DB
	model.DB = db
	t.Cleanup(func() { model.DB = previousDB })
	require.NoError(t, db.AutoMigrate(&model.User{}))

	user := model.User{Username: "billing-user", Role: common.RoleCommonUser}
	admin := model.User{Username: "billing-admin", Role: common.RoleAdminUser}
	require.NoError(t, db.Create(&user).Error)
	require.NoError(t, db.Create(&admin).Error)

	require.ErrorIs(t, AuthorizeBillingExportUser(user.Id, admin.Id), ErrBillingExportPermission)
	require.NoError(t, AuthorizeBillingExportUser(admin.Id, user.Id))
	require.ErrorIs(t, AuthorizeBillingExportUser(admin.Id, 9999), gorm.ErrRecordNotFound)
}
