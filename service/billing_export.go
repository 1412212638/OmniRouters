package service

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"sort"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/pkg/billexport"
	"github.com/shopspring/decimal"
)

const maxBillingExportMonths = 12

var ErrBillingExportPeriod = errors.New("invalid billing export period")
var ErrBillingExportLanguage = errors.New("invalid billing export language")
var errBillingArchiveTooLarge = errors.New("billing export archive exceeds the size limit")

type billingExportHandler struct{}

func (billingExportHandler) Type() string { return model.BillingExportTaskType }

func init() {
	RegisterSystemTaskHandler(billingExportHandler{})
	RegisterSystemTaskHandler(billingExportCleanup{})
}

func billingExportPeriod(startMonth, endMonth string, asOf time.Time) (time.Time, time.Time, error) {
	start, startErr := time.ParseInLocation("2006-01", startMonth, billexport.Timezone)
	endMonthStart, endErr := time.ParseInLocation("2006-01", endMonth, billexport.Timezone)
	if startErr != nil || endErr != nil || start.Year() < 2000 || endMonthStart.Before(start) ||
		!endMonthStart.Before(start.AddDate(0, maxBillingExportMonths, 0)) || endMonthStart.After(asOf) {
		return time.Time{}, time.Time{}, ErrBillingExportPeriod
	}
	end := endMonthStart.AddDate(0, 1, 0)
	// The creation second is excluded so subsequent writes with the same
	// timestamp do not change the end boundary.
	if end.After(asOf) {
		end = asOf
	}
	return start, end, nil
}

func EnqueueBillingExport(ctx context.Context, scope model.BillingExport) error {
	switch scope.Language {
	case "", model.BillingExportLanguageEnglish, model.BillingExportLanguageChinese:
		scope.Language = scope.EffectiveLanguage()
	default:
		return ErrBillingExportLanguage
	}
	if _, _, err := billingExportPeriod(scope.StartMonth, scope.EndMonth, time.Now()); err != nil {
		return err
	}
	scope.Display = billingExportDisplaySnapshot()
	if err := model.CreateBillingExport(ctx, scope); err != nil {
		return err
	}
	notifySystemTaskRunner()
	return nil
}

func (billingExportHandler) Run(ctx context.Context, task *model.SystemTask, runnerID string) {
	err := runBillingExport(ctx, task.TaskID, runnerID)
	if err == nil {
		return
	}
	logger.LogWarn(ctx, fmt.Sprintf("billing export %s failed: %v", task.TaskID, err))
	reason := "generation_failed"
	if errors.Is(err, model.ErrBillingExportTooManyRows) || errors.Is(err, errBillingArchiveTooLarge) {
		reason = "limit_exceeded"
	}
	failSystemTask(task, runnerID, errors.New(reason))
}

func runBillingExport(ctx context.Context, taskID, runnerID string) error {
	scope, err := model.GetBillingExport(ctx, taskID)
	if err != nil {
		return err
	}
	if scope.RequestedBy != scope.UserID {
		if err := AuthorizeBillingExportUser(scope.RequestedBy, scope.UserID); err != nil {
			return err
		}
	}
	requesterID := scope.RequestedBy
	if requesterID == 0 {
		requesterID = scope.UserID
	}
	if err := AuthorizeBillingExport(requesterID, scope.OrganizationID); err != nil {
		return err
	}
	statement, err := loadBillingStatement(ctx, scope)
	if err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	archive, err := billexport.WriteArchive(statement)
	if err != nil {
		return err
	}
	if len(archive) > model.MaxBillingExportBytes {
		return errBillingArchiveTooLarge
	}
	return model.FinishBillingExport(ctx, taskID, runnerID, archive)
}

var ErrBillingExportPermission = errors.New("organization manager permission required")

func AuthorizeBillingExport(userID, organizationID int) error {
	if userID <= 0 || organizationID < 0 {
		return ErrBillingExportPermission
	}
	if organizationID == 0 {
		return nil
	}
	if !common.OrganizationEnabled {
		return ErrBillingExportPermission
	}
	member, err := model.GetOrganizationMember(organizationID, userID)
	if err != nil || member.Status != model.OrganizationMemberActive || !model.OrganizationCanManage(member.Role) {
		return ErrBillingExportPermission
	}
	return nil
}

func AuthorizeBillingExportUser(viewerID, targetUserID int) error {
	if viewerID <= 0 || targetUserID <= 0 {
		return ErrBillingExportPermission
	}
	viewer, err := model.GetUserById(viewerID, false)
	if err != nil || viewer.Role < common.RoleAdminUser {
		return ErrBillingExportPermission
	}
	_, err = model.GetUserById(targetUserID, false)
	return err
}

type billingExportLogOther struct {
	Source           string `json:"billing_source"`
	SubscriptionMode string `json:"subscription_billing_mode"`
	SubscriptionID   int    `json:"subscription_id"`
	SettlementFailed bool   `json:"billing_settlement_failed"`
	WalletToolQuota  int    `json:"wallet_tool_quota"`
	TaskID           string `json:"task_id"`
	IsTask           bool   `json:"is_task"`
	ActualQuota      *int   `json:"actual_quota"`
	OriginalQuota    *int   `json:"original_quota"`
	CacheRead        int    `json:"cache_tokens"`
	CacheWrite       int    `json:"cache_write_tokens"`
	CacheCreation    int    `json:"cache_creation_tokens"`
	CacheCreation5m  int    `json:"cache_creation_tokens_5m"`
	CacheCreation1h  int    `json:"cache_creation_tokens_1h"`
	CacheCreation12h int    `json:"cache_creation_tokens_12h"`
}

func loadBillingStatement(ctx context.Context, scope model.BillingExport) (billexport.Statement, error) {
	display := resolveBillingQuotaDisplay(scope.Display)
	statement := billexport.Statement{
		StartMonth:     scope.StartMonth,
		EndMonth:       scope.EndMonth,
		Language:       scope.EffectiveLanguage(),
		AsOf:           time.Unix(scope.CreatedAt, 0),
		Organization:   scope.OrganizationID > 0,
		DisplayUnit:    display.unit,
		CurrencySymbol: display.symbol,
		ExchangeRate:   display.rate.String(),
	}
	label, accountID := "Personal account", scope.UserID
	if statement.Organization {
		label, accountID = "Organization", scope.OrganizationID
	}
	if statement.Language == model.BillingExportLanguageChinese {
		label = "个人账户"
		if statement.Organization {
			label = "组织"
		}
	}
	statement.Account = fmt.Sprintf("%s #%d", label, accountID)
	start, end, err := billingExportPeriod(scope.StartMonth, scope.EndMonth, statement.AsOf)
	if err != nil {
		return statement, err
	}
	data, err := model.LoadBillingExportData(ctx, scope, start.Unix(), end.Unix())
	if err != nil {
		return statement, err
	}
	if len(data.Payments)+len(data.Subscriptions)+len(data.Funding)+len(data.Transfers) > model.BillingExportRowLimit {
		return statement, model.ErrBillingExportTooManyRows
	}
	statement.Events = billingFundingEvents(scope, display, data)
	usage := billingUsageAccumulator{indexes: make(map[billingUsageGroup]int)}
	err = model.VisitBillingExportLogs(ctx, scope, start.Unix(), end.Unix(), func(logs []model.Log) error {
		events, err := normalizeBillingEvents(ctx, scope, display, logs)
		if err != nil {
			return err
		}
		return usage.add(events)
	})
	if err == nil {
		statement.Events = append(statement.Events, usage.events...)
		sort.SliceStable(statement.Events, func(i, j int) bool {
			return statement.Events[i].PostedAt < statement.Events[j].PostedAt
		})
	}
	return statement, err
}

func normalizeBillingEvents(ctx context.Context, scope model.BillingExport, display billingQuotaDisplay, logs []model.Log) ([]billexport.Event, error) {
	metadata := make([]billingExportLogOther, len(logs))
	taskIDs := make(map[string]struct{})
	for i, log := range logs {
		if log.Other != "" {
			// A historical or malformed metadata field can omit optional details.
			// Decode into this allowlist only. Never export the source JSON.
			_ = common.UnmarshalJsonStr(log.Other, &metadata[i])
		}
		if id := metadata[i].TaskID; id != "" {
			taskIDs[id] = struct{}{}
		}
	}
	tasks, err := model.BillingTaskSources(ctx, scope, slices.Collect(maps.Keys(taskIDs)))
	if err != nil {
		return nil, err
	}
	subIDs := make(map[int]struct{})
	for i, other := range metadata {
		if task, ok := tasks[other.TaskID]; ok && task.UserId == logs[i].UserId {
			if other.Source == "" {
				metadata[i].Source = task.PrivateData.BillingSource
			}
			if other.SubscriptionID == 0 {
				metadata[i].SubscriptionID = task.PrivateData.SubscriptionId
			}
		}
		if id := metadata[i].SubscriptionID; id > 0 {
			subIDs[id] = struct{}{}
		}
	}
	modes, err := model.BillingSubscriptionModes(ctx, scope.UserID, slices.Collect(maps.Keys(subIDs)))
	if err != nil {
		return nil, err
	}
	return billingUsageEvents(scope, display, logs, metadata, tasks, modes), nil
}

func billingFundingEvents(scope model.BillingExport, display billingQuotaDisplay, data model.BillingExportData) []billexport.Event {
	events := make([]billexport.Event, 0, len(data.Payments)+len(data.Subscriptions)+len(data.Funding)+len(data.Transfers))
	for _, payment := range data.Payments {
		events = append(events, paymentEvent(payment.CompleteTime, payment.UserId, "wallet_topup"))
	}
	for _, order := range data.Subscriptions {
		events = append(events, paymentEvent(order.CompleteTime, order.UserId, "subscription_order"))
	}
	for _, entry := range data.Funding {
		kind := billexport.Funding
		if entry.Kind == model.FundingAdjustment {
			kind = billexport.Adjustment
		}
		event := billexport.Event{PostedAt: entry.CreatedAt, UserID: entry.UserId, Kind: kind, Source: "wallet_" + entry.Kind, Unit: display.unit, Amount: display.amount(entry.Amount)}
		events = append(events, event)
	}
	for _, entry := range data.Transfers {
		event := billexport.Event{PostedAt: entry.CreatedAt, UserID: entry.UserId, Kind: billexport.Funding, Source: "organization_" + entry.Type, Unit: display.unit, Amount: display.amount(int64(entry.Amount))}
		if scope.OrganizationID == 0 {
			event.Kind = billexport.Adjustment
			event.Source = "wallet_transfer_out"
			event.Amount = event.Amount.Neg()
		}
		events = append(events, event)
	}
	return events
}

func billingUsageEvents(scope model.BillingExport, display billingQuotaDisplay, logs []model.Log, metadata []billingExportLogOther, tasks map[string]model.Task, modes map[int]string) []billexport.Event {
	events := make([]billexport.Event, 0, len(logs))
	type taskKey struct {
		taskID string
		userID int
	}
	type apiKey struct {
		id   int
		name string
	}
	taskKeys := make(map[taskKey]apiKey)
	for index, log := range logs {
		other := metadata[index]
		event := logBillingEvent(scope, display, log, other, tasks[other.TaskID], modes[other.SubscriptionID])
		if event.Kind == billexport.Precharge && event.TaskID != "" && event.KeyID > 0 {
			taskKeys[taskKey{event.TaskID, event.UserID}] = apiKey{event.KeyID, event.KeyName}
		}
		events = append(events, event)
		if other.WalletToolQuota > 0 && !other.SettlementFailed {
			// One request can use subscription tokens and incur a separate
			// wallet fee. Count its usage only on the main event.
			events = append(events, billexport.Event{PostedAt: event.PostedAt, UsedAt: event.UsedAt, CompletedAt: event.CompletedAt, Kind: billexport.Charge, Source: "wallet_tool_fee", Unit: display.unit, Model: event.Model, KeyID: event.KeyID, KeyName: event.KeyName, UserID: event.UserID, Amount: display.amount(int64(other.WalletToolQuota))})
		}
	}
	for i := range events {
		event := &events[i]
		if key, ok := taskKeys[taskKey{event.TaskID, event.UserID}]; ok {
			if event.KeyID == 0 {
				event.KeyID = key.id
			}
			if event.KeyID == key.id && event.KeyName == "" {
				event.KeyName = key.name
			}
		}
	}
	return events
}

func paymentEvent(postedAt int64, userID int, source string) billexport.Event {
	// Historical orders lack a currency snapshot. The account summary counts
	// the orders but cannot add their recorded amounts across currencies.
	return billexport.Event{PostedAt: postedAt, UserID: userID, Kind: billexport.Payment, Source: source, Unit: billexport.UnverifiedCurrency}
}

func logBillingEvent(scope model.BillingExport, display billingQuotaDisplay, log model.Log, other billingExportLogOther, task model.Task, mode string) billexport.Event {
	event := billexport.Event{PostedAt: log.CreatedAt, Kind: billexport.Charge, Source: BillingSourceWallet, Unit: display.unit, Model: log.ModelName, KeyID: log.TokenId, KeyName: log.TokenName, UserID: log.UserId, TaskID: other.TaskID, Requests: 1,
		InputTokens: int64(log.PromptTokens), OutputTokens: int64(log.CompletionTokens), CacheReadTokens: int64(other.CacheRead),
		CacheWriteTokens: max(int64(other.CacheWrite), int64(other.CacheCreation), int64(other.CacheCreation5m)+int64(other.CacheCreation1h)+int64(other.CacheCreation12h)), Amount: display.amount(int64(log.Quota))}
	if task.TaskID != "" && task.UserId == log.UserId {
		event.UsedAt = task.StartTime
		if event.UsedAt == 0 {
			event.UsedAt = task.SubmitTime
		}
		if task.FinishTime > 0 && task.FinishTime >= event.UsedAt && task.FinishTime <= log.CreatedAt {
			event.CompletedAt = task.FinishTime
		}
		if event.KeyID == 0 {
			event.KeyID = task.PrivateData.TokenId
		}
	}
	if event.UsedAt == 0 && other.TaskID == "" && !other.IsTask {
		event.UsedAt = max(0, log.CreatedAt-int64(max(log.UseTime, 0)))
		event.CompletedAt = log.CreatedAt
	}
	switch {
	case scope.OrganizationID > 0:
		event.Source = BillingSourceOrganization
	case other.Source == BillingSourceSubscription || other.SubscriptionID > 0:
		event.Source = BillingSourceSubscription
		if other.SubscriptionMode != "" {
			mode = other.SubscriptionMode
		}
		switch mode {
		case model.SubscriptionBillingModeToken:
			event.Unit, event.Amount = billexport.UnitTokens, decimal.NewFromInt(int64(log.Quota))
		case model.SubscriptionBillingModeQuota:
		default:
			event.Unit, event.Amount = "Unverified subscription units", decimal.NewFromInt(int64(log.Quota))
		}
	case other.TaskID != "" && other.Source != BillingSourceWallet:
		event.Source = "unverified_funding_source"
		event.Unit, event.Amount = "Unverified usage units", decimal.NewFromInt(int64(log.Quota))
	}
	switch {
	case log.Type == model.LogTypeRefund:
		event.Kind, event.Requests, event.Amount = billexport.Refund, 0, event.Amount.Abs().Neg()
	case other.TaskID != "" && other.ActualQuota != nil:
		event.Kind, event.Requests = billexport.AdditionalCharge, 0
	case other.IsTask:
		event.Kind = billexport.Precharge
	}
	if other.SettlementFailed {
		event.Kind = billexport.UnconfirmedUsage
	}
	if other.OriginalQuota != nil && event.Unit == display.unit && event.Kind == billexport.Charge {
		original := display.amount(int64(*other.OriginalQuota))
		event.Original = &original
	}
	return event
}

type billingExportCleanup struct{}

func (billingExportCleanup) Type() string            { return "billing_export_cleanup" }
func (billingExportCleanup) Enabled() bool           { return true }
func (billingExportCleanup) Interval() time.Duration { return 6 * time.Hour }
func (billingExportCleanup) NewPayload() any         { return nil }
func (billingExportCleanup) Run(ctx context.Context, task *model.SystemTask, runnerID string) {
	if err := model.DeleteExpiredBillingExports(ctx, time.Now().Unix()); err != nil {
		failSystemTask(task, runnerID, err)
		return
	}
	if err := model.FinishSystemTask(task.TaskID, runnerID, model.SystemTaskStatusSucceeded, nil, ""); err != nil {
		logSystemTaskLockError(ctx, task, err)
	}
}
