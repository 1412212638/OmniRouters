package model

import (
	"context"
	"errors"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

const BillingExportRowLimit = 25000
const billingExportLogBatchSize = 2000

var ErrBillingExportTooManyRows = errors.New("billing export exceeds the row limit")

type BillingExportData struct {
	Payments      []TopUp
	Subscriptions []SubscriptionOrder
	Funding       []WalletFundingEntry
	Transfers     []OrganizationWalletTransaction
}

func LoadBillingExportData(ctx context.Context, scope BillingExport, start, end int64) (BillingExportData, error) {
	var data BillingExportData
	personal := scope.OrganizationID == 0
	payments := DB.WithContext(ctx).Model(&TopUp{}).
		Select("user_id, complete_time").
		Where("status = ? AND complete_time >= ? AND complete_time < ?", common.TopUpStatusSuccess, start, end)
	transfers := DB.WithContext(ctx).Model(&OrganizationWalletTransaction{}).
		Select("user_id, created_at, amount, type").
		Where("created_at >= ? AND created_at < ?", start, end)

	if personal {
		payments = payments.Where("top_ups.user_id = ? AND COALESCE(top_ups.organization_id, 0) = 0", scope.UserID).
			Where("NOT EXISTS (?)", DB.Model(&SubscriptionOrder{}).Select("1").Where("subscription_orders.trade_no = top_ups.trade_no"))
		transfers = transfers.Where("user_id = ? AND type = ?", scope.UserID, "transfer_in")
		err := findBillingRecords(DB.WithContext(ctx).Model(&SubscriptionOrder{}).
			Select("user_id, complete_time").
			Where("user_id = ? AND status = ? AND complete_time >= ? AND complete_time < ?", scope.UserID, common.TopUpStatusSuccess, start, end), &data.Subscriptions)
		if err != nil {
			return data, err
		}
		err = findBillingRecords(DB.WithContext(ctx).Model(&WalletFundingEntry{}).
			Select("user_id, created_at, amount, kind").
			Where("user_id = ? AND created_at >= ? AND created_at < ?", scope.UserID, start, end), &data.Funding)
		if err != nil {
			return data, err
		}
	} else {
		payments = payments.Where("organization_id = ?", scope.OrganizationID)
		transfers = transfers.Where("organization_id = ? AND type IN ?", scope.OrganizationID, []string{"topup", "transfer_in"})
	}
	if err := findBillingRecords(payments, &data.Payments); err != nil {
		return data, err
	}
	if err := findBillingRecords(transfers, &data.Transfers); err != nil {
		return data, err
	}
	return data, nil
}

// VisitBillingExportLogs reads the scoped billing records in bounded batches.
// The caller consumes a batch before returning from visit.
func VisitBillingExportLogs(ctx context.Context, scope BillingExport, start, end int64, visit func([]Log) error) error {
	query := LOG_DB.WithContext(ctx).Model(&Log{}).
		Select("id, user_id, created_at, type, token_id, token_name, model_name, quota, prompt_tokens, completion_tokens, use_time, other").
		Where("created_at >= ? AND created_at < ? AND type IN ?", start, end, []int{LogTypeConsume, LogTypeRefund})
	if scope.OrganizationID == 0 {
		query = query.Where("user_id = ? AND COALESCE(organization_id, 0) = 0", scope.UserID)
	} else {
		query = query.Where("organization_id = ?", scope.OrganizationID)
	}
	if common.UsingLogDatabase(common.DatabaseTypeClickHouse) {
		// ClickHouse log IDs may be zero. Stream the rows from the separate log
		// database without relying on an ID or request ID being unique.
		return visitClickHouseBillingExportLogs(query, visit)
	}

	// Fix the upper ID before the first page so new rows appended during the
	// export do not enter later pages, even if their posting time is backdated.
	var lastLog Log
	last := LOG_DB.WithContext(ctx).Model(&Log{}).Select("id").Order("id DESC").Limit(1).Find(&lastLog)
	if last.Error != nil || last.RowsAffected == 0 {
		return last.Error
	}
	query = query.Where("id <= ?", lastLog.Id).
		Order("created_at ASC, id ASC").Limit(billingExportLogBatchSize).
		Session(&gorm.Session{})

	var batch []Log
	var cursor *Log
	for {
		page := query
		if cursor != nil {
			page = page.Where("created_at >= ? AND (created_at > ? OR id > ?)", cursor.CreatedAt, cursor.CreatedAt, cursor.Id)
		}
		if err := page.Find(&batch).Error; err != nil {
			return err
		}
		if len(batch) == 0 {
			return nil
		}
		lastRow := batch[len(batch)-1]
		if err := visit(batch); err != nil {
			return err
		}
		if len(batch) < billingExportLogBatchSize {
			return nil
		}
		cursor = &lastRow
	}
}

func visitClickHouseBillingExportLogs(query *gorm.DB, visit func([]Log) error) error {
	rows, err := query.Rows()
	if err != nil {
		return err
	}
	defer rows.Close()

	batch := make([]Log, 0, billingExportLogBatchSize)
	for rows.Next() {
		batch = batch[:len(batch)+1]
		if err := query.ScanRows(rows, &batch[len(batch)-1]); err != nil {
			return err
		}
		if len(batch) == billingExportLogBatchSize {
			if err := visit(batch); err != nil {
				return err
			}
			batch = batch[:0]
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if len(batch) > 0 {
		return visit(batch)
	}
	return nil
}

func findBillingRecords[T any](query *gorm.DB, rows *[]T) error {
	if err := query.Limit(BillingExportRowLimit + 1).Find(rows).Error; err != nil {
		return err
	}
	if len(*rows) > BillingExportRowLimit {
		return ErrBillingExportTooManyRows
	}
	return nil
}

func BillingTaskSources(ctx context.Context, scope BillingExport, taskIDs []string) (map[string]Task, error) {
	tasks := make(map[string]Task, len(taskIDs))
	for start := 0; start < len(taskIDs); start += 400 {
		end := min(start+400, len(taskIDs))
		query := DB.WithContext(ctx).Model(&Task{}).Select("task_id, user_id, submit_time, start_time, finish_time, private_data").Where("task_id IN ?", taskIDs[start:end])
		if scope.OrganizationID == 0 {
			query = query.Where("user_id = ? AND COALESCE(organization_id, 0) = 0", scope.UserID)
		} else {
			query = query.Where("organization_id = ?", scope.OrganizationID)
		}
		var batch []Task
		if err := query.Find(&batch).Error; err != nil {
			return nil, err
		}
		for _, task := range batch {
			tasks[task.TaskID] = task
		}
	}
	return tasks, nil
}

func BillingSubscriptionModes(ctx context.Context, userID int, subscriptionIDs []int) (map[int]string, error) {
	modes := make(map[int]string, len(subscriptionIDs))
	for start := 0; start < len(subscriptionIDs); start += 400 {
		end := min(start+400, len(subscriptionIDs))
		var subscriptions []UserSubscription
		err := DB.WithContext(ctx).Model(&UserSubscription{}).Select("id, billing_mode").
			Where("user_id = ? AND id IN ?", userID, subscriptionIDs[start:end]).Find(&subscriptions).Error
		if err != nil {
			return nil, err
		}
		for _, subscription := range subscriptions {
			modes[subscription.Id] = subscription.BillingMode
		}
	}
	return modes, nil
}
