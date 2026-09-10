package model

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

const (
	DefaultAuditLogRetentionDays = 90
	MaxAuditLogRetentionDays     = 3650
	auditLogCleanupBatchSize     = 100
)

func CountOldAuditLogs(ctx context.Context, before int64) (int64, error) {
	var count int64
	err := DB.WithContext(ctx).Model(&AuditLog{}).Where("created_at < ?", before).Count(&count).Error
	return count, err
}

func DeleteOldAuditLogBatch(ctx context.Context, before int64, limit int) (int64, error) {
	if limit <= 0 { limit = auditLogCleanupBatchSize }
	var ids []int64
	if err := DB.WithContext(ctx).Model(&AuditLog{}).Where("created_at < ?", before).Order("id asc").Limit(limit).Pluck("id", &ids).Error; err != nil { return 0, err }
	if len(ids) == 0 { return 0, nil }
	result := DB.WithContext(ctx).Where("id IN ?", ids).Delete(&AuditLog{})
	return result.RowsAffected, result.Error
}

const (
	AuditCategorySecurity    = "security"
	AuditCategoryAccessToken = "access_token"
)

// AuditLog is independent from usage-log retention and never stores bearer tokens.
type AuditLog struct {
	Id         int    `json:"id" gorm:"primaryKey"`
	EventId    string `json:"event_id" gorm:"type:varchar(64);uniqueIndex"`
	UserId     int    `json:"user_id" gorm:"index"`
	Username   string `json:"username" gorm:"type:varchar(64);index"`
	CreatedAt  int64  `json:"created_at" gorm:"index"`
	Category   string `json:"category" gorm:"type:varchar(24);index"`
	Action     string `json:"action" gorm:"type:varchar(128)"`
	TokenRef   string `json:"token_ref" gorm:"type:varchar(64);index"`
	Ip         string `json:"ip" gorm:"type:varchar(64)"`
	Success    bool   `json:"success"`
	Outcome    string `json:"outcome" gorm:"type:varchar(16);index"`
	RequestId  string `json:"request_id" gorm:"type:varchar(64);index"`
	Other      string `json:"other,omitempty" gorm:"type:text"`
}

func AccessTokenFingerprint(token string) string {
	token = strings.TrimSpace(token)
	if token == "" { return "" }
	digest := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", digest)
}

func RecordAuditLog(entry *AuditLog) error {
	if entry == nil || entry.UserId < 0 || (entry.UserId == 0 && entry.Category != AuditCategorySecurity) || entry.Category == "" || entry.Action == "" {
		return fmt.Errorf("invalid audit log")
	}
	if entry.CreatedAt == 0 { entry.CreatedAt = common.GetTimestamp() }
	if entry.EventId == "" { entry.EventId = common.NewRequestId() }
	return DB.Create(entry).Error
}

func RevokeUserAccessToken(userId int) error {
	if userId <= 0 { return fmt.Errorf("invalid user id") }
	return DB.Transaction(func(tx *gorm.DB) error {
		var user User
		if err := tx.Select("id", "access_token").First(&user, userId).Error; err != nil { return err }
		ref := AccessTokenFingerprint(user.GetAccessToken())
		if err := tx.Model(&User{}).Where("id = ?", userId).Updates(map[string]interface{}{"access_token": nil, "access_token_created_at": nil}).Error; err != nil { return err }
		if ref != "" { return tx.Create(&AuditLog{UserId: userId, EventId: common.NewRequestId(), CreatedAt: common.GetTimestamp(), Category: AuditCategoryAccessToken, Action: "revoke", TokenRef: ref, Success: true, Outcome: "success"}).Error }
		return nil
	})
}
