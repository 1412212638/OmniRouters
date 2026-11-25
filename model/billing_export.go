package model

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

const BillingExportTaskType = "billing_export"
const BillingExportRetention = 7 * 24 * time.Hour
const MaxBillingExportBytes = 2 * 1024 * 1024
const BillingExportLanguageEnglish = "en"
const BillingExportLanguageChinese = "zh"

var ErrBillingExportAlreadyQueued = errors.New("billing export is already queued")

type BillingExport struct {
	TaskID         string               `json:"task_id" gorm:"type:varchar(64);primaryKey"`
	UserID         int                  `json:"-" gorm:"index:idx_billing_export_owner,priority:1;not null"`
	RequestedBy    int                  `json:"-" gorm:"not null;default:0"`
	OrganizationID int                  `json:"-" gorm:"index:idx_billing_export_owner,priority:2;not null"`
	StartMonth     string               `json:"start_month" gorm:"type:varchar(7);not null"`
	EndMonth       string               `json:"end_month" gorm:"type:varchar(7);not null"`
	Language       string               `json:"language" gorm:"size:2;not null;default:en"`
	Display        BillingExportDisplay `json:"-" gorm:"embedded;embeddedPrefix:display_"`
	CreatedAt      int64                `json:"created_at" gorm:"index:idx_billing_export_owner,priority:3"`
	ExpiresAt      int64                `json:"expires_at" gorm:"index;not null"`
	Archive        []byte               `json:"-"`
}

// BillingExportDisplay captures the presentation settings used by the job.
type BillingExportDisplay struct {
	Mode         string `gorm:"size:8"`
	Symbol       string `gorm:"size:32"`
	Rate         float64
	QuotaPerUnit float64
}

type BillingExportItem struct {
	BillingExport
	Status SystemTaskStatus `json:"status"`
	Error  string           `json:"error,omitempty"`
}

func (report BillingExport) EffectiveLanguage() string {
	if report.Language == BillingExportLanguageChinese {
		return BillingExportLanguageChinese
	}
	return BillingExportLanguageEnglish
}

func CreateBillingExport(ctx context.Context, report BillingExport) error {
	// The runner serializes work across nodes. The active key prevents one
	// account from queuing another export until its current job finishes.
	activeKey := fmt.Sprintf("billing_export:%d", report.UserID)
	err := DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		task, err := createSystemTask(tx, BillingExportTaskType, activeKey, nil, nil)
		if err != nil {
			return err
		}
		report.TaskID = task.TaskID
		report.CreatedAt = task.CreatedAt
		report.ExpiresAt = task.CreatedAt + int64(BillingExportRetention/time.Second)
		return tx.Create(&report).Error
	})
	if err != nil {
		var existing SystemTask
		if lookupErr := DB.WithContext(ctx).Select("task_id").Where("active_key = ?", activeKey).Take(&existing).Error; lookupErr == nil {
			return ErrBillingExportAlreadyQueued
		}
	}
	return err
}

func ListBillingExports(ctx context.Context, userID, organizationID int, now int64) ([]BillingExportItem, error) {
	items := make([]BillingExportItem, 0)
	err := DB.WithContext(ctx).Model(&BillingExport{}).
		Select("billing_exports.task_id, billing_exports.start_month, billing_exports.end_month, billing_exports.language, billing_exports.created_at, billing_exports.expires_at, system_tasks.status, system_tasks.error").
		Joins("JOIN system_tasks ON system_tasks.task_id = billing_exports.task_id").
		Where("billing_exports.user_id = ? AND billing_exports.organization_id = ? AND billing_exports.expires_at > ?", userID, organizationID, now).
		Order("billing_exports.created_at DESC, billing_exports.task_id DESC").Limit(30).Scan(&items).Error
	for i := range items {
		items[i].Language = items[i].EffectiveLanguage()
	}
	return items, err
}

func GetBillingExport(ctx context.Context, taskID string) (BillingExport, error) {
	var report BillingExport
	err := DB.WithContext(ctx).Omit("archive").Where("task_id = ?", taskID).Take(&report).Error
	return report, err
}

func ReadBillingExportArchive(ctx context.Context, taskID string, userID, organizationID int, now int64) ([]byte, error) {
	var report BillingExport
	err := DB.WithContext(ctx).Model(&BillingExport{}).Select("billing_exports.archive").
		Joins("JOIN system_tasks ON system_tasks.task_id = billing_exports.task_id").
		Where("billing_exports.task_id = ? AND billing_exports.user_id = ? AND billing_exports.organization_id = ? AND billing_exports.expires_at > ? AND system_tasks.status = ?", taskID, userID, organizationID, now, SystemTaskStatusSucceeded).
		Take(&report).Error
	return report.Archive, err
}

func FinishBillingExport(ctx context.Context, taskID, runnerID string, archive []byte) error {
	if len(archive) == 0 || len(archive) > MaxBillingExportBytes {
		return fmt.Errorf("billing archive size out of range")
	}
	return DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := finishSystemTask(tx, taskID, runnerID, SystemTaskStatusSucceeded, nil, ""); err != nil {
			return err
		}
		result := tx.Model(&BillingExport{}).Where("task_id = ?", taskID).Update("archive", archive)
		if result.Error == nil && result.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}
		return result.Error
	})
}

func DeleteExpiredBillingExports(ctx context.Context, now int64) error {
	return DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var ids []string
		if err := tx.Model(&BillingExport{}).Where("expires_at <= ?", now).Limit(200).Pluck("task_id", &ids).Error; err != nil || len(ids) == 0 {
			return err
		}
		// Leave live job metadata available to the runner until it finishes.
		if err := tx.Where("task_id IN ? AND status IN ?", ids, []SystemTaskStatus{SystemTaskStatusSucceeded, SystemTaskStatusFailed}).Delete(&SystemTask{}).Error; err != nil {
			return err
		}
		remaining := tx.Model(&SystemTask{}).Select("task_id").Where("task_id IN ?", ids)
		return tx.Where("task_id IN ? AND task_id NOT IN (?)", ids, remaining).Delete(&BillingExport{}).Error
	})
}
