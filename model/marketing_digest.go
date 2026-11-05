package model

import (
	"errors"
	"time"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

const (
	MarketingDigestStatusPendingReview = "pending_review"
	MarketingDigestStatusSending       = "sending"
	MarketingDigestStatusSent          = "sent"
	MarketingDigestStatusSentWithError = "sent_with_errors"
	MarketingDigestStatusFailed        = "failed"

	MarketingDigestRecipientPending = "pending"
	MarketingDigestRecipientSending = "sending"
	MarketingDigestRecipientSent    = "sent"
	MarketingDigestRecipientSkipped = "skipped"
	MarketingDigestRecipientFailed  = "failed"
)

type MarketingDigest struct {
	ID                 int64  `json:"id" gorm:"primaryKey"`
	PeriodStart        int64  `json:"period_start" gorm:"not null;index"`
	PeriodEnd          int64  `json:"period_end" gorm:"not null;uniqueIndex"`
	Subject            string `json:"subject" gorm:"type:text;not null"`
	Content            string `json:"content" gorm:"type:text;not null"`
	ModelNames         string `json:"model_names" gorm:"type:text"`
	PromotionSnapshot  string `json:"promotion_snapshot" gorm:"type:text"`
	AudienceType       string `json:"audience_type" gorm:"type:varchar(32);not null;default:all"`
	AudienceGroup      string `json:"audience_group" gorm:"type:varchar(64)"`
	AudienceUserIDs    string `json:"audience_user_ids" gorm:"type:text"`
	Status             string `json:"status" gorm:"type:varchar(32);not null;index"`
	CreatedAt          int64  `json:"created_at" gorm:"not null;index"`
	UpdatedAt          int64  `json:"updated_at" gorm:"not null"`
	ReviewedBy         int    `json:"reviewed_by"`
	SentAt             int64  `json:"sent_at"`
	RecipientCount     int64  `json:"recipient_count" gorm:"-"`
	SentCount          int64  `json:"sent_count" gorm:"-"`
	SkippedCount       int64  `json:"skipped_count" gorm:"-"`
	FailedCount        int64  `json:"failed_count" gorm:"-"`
}

type MarketingDigestPromotion struct {
	ID          int64  `json:"id" gorm:"primaryKey"`
	ModelName   string `json:"model_name" gorm:"type:varchar(128);not null;index"`
	Description string `json:"description" gorm:"type:text"`
	Discount    string `json:"discount" gorm:"type:varchar(128)"`
	StartsAt    int64  `json:"starts_at" gorm:"not null;default:0"`
	EndsAt      int64  `json:"ends_at" gorm:"not null;default:0"`
	URL         string `json:"url" gorm:"type:text"`
	Enabled     bool   `json:"enabled" gorm:"not null;default:true;index"`
	CreatedAt   int64  `json:"created_at" gorm:"not null"`
	UpdatedAt   int64  `json:"updated_at" gorm:"not null"`
}

type MarketingDigestRecipient struct {
	ID          int64  `json:"id" gorm:"primaryKey"`
	DigestID    int64  `json:"digest_id" gorm:"not null;uniqueIndex:uk_marketing_digest_user,priority:1;index"`
	UserID      int    `json:"user_id" gorm:"not null;uniqueIndex:uk_marketing_digest_user,priority:2;index"`
	EmailMasked string `json:"email_masked" gorm:"type:varchar(128)"`
	Status      string `json:"status" gorm:"type:varchar(16);not null;index"`
	Reason      string `json:"reason" gorm:"type:text"`
	UpdatedAt   int64  `json:"updated_at" gorm:"not null"`
}

func (digest *MarketingDigest) BeforeCreate(_ *gorm.DB) error {
	now := common.GetTimestamp()
	if digest.CreatedAt == 0 {
		digest.CreatedAt = now
	}
	if digest.UpdatedAt == 0 {
		digest.UpdatedAt = now
	}
	if digest.AudienceType == "" {
		digest.AudienceType = "all"
	}
	if digest.Status == "" {
		digest.Status = MarketingDigestStatusPendingReview
	}
	return nil
}

func (promotion *MarketingDigestPromotion) BeforeCreate(_ *gorm.DB) error {
	now := common.GetTimestamp()
	if promotion.CreatedAt == 0 {
		promotion.CreatedAt = now
	}
	if promotion.UpdatedAt == 0 {
		promotion.UpdatedAt = now
	}
	return nil
}

func CreateMarketingDigest(digest *MarketingDigest) error {
	return DB.Create(digest).Error
}

func GetMarketingDigest(id int64) (*MarketingDigest, error) {
	var digest MarketingDigest
	err := DB.First(&digest, id).Error
	return &digest, err
}

func GetMarketingDigestByPeriodEnd(periodEnd int64) (*MarketingDigest, error) {
	var digest MarketingDigest
	err := DB.Where("period_end = ?", periodEnd).First(&digest).Error
	return &digest, err
}

func GetLatestMarketingDigest() (*MarketingDigest, error) {
	var digest MarketingDigest
	err := DB.Order("period_end DESC").First(&digest).Error
	return &digest, err
}

func ListMarketingDigests(limit int) ([]MarketingDigest, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	var digests []MarketingDigest
	err := DB.Order("created_at DESC").Limit(limit).Find(&digests).Error
	return digests, err
}

func UpdateMarketingDigestDraft(id int64, subject, content, audienceType, audienceGroup, audienceUserIDs string) error {
	result := DB.Model(&MarketingDigest{}).
		Where("id = ? AND status = ?", id, MarketingDigestStatusPendingReview).
		Updates(map[string]interface{}{
			"subject":            subject,
			"content":            content,
			"audience_type":      audienceType,
			"audience_group":     audienceGroup,
			"audience_user_ids":  audienceUserIDs,
			"updated_at":         common.GetTimestamp(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("newsletter is no longer editable")
	}
	return nil
}

func SetMarketingDigestStatus(id int64, status string, reviewerID int, sentAt int64) error {
	updates := map[string]interface{}{"status": status, "updated_at": common.GetTimestamp()}
	if reviewerID > 0 {
		updates["reviewed_by"] = reviewerID
	}
	if sentAt > 0 {
		updates["sent_at"] = sentAt
	}
	return DB.Model(&MarketingDigest{}).Where("id = ?", id).Updates(updates).Error
}

func CreateMarketingDigestRecipients(digestID int64, userIDs []int) error {
	if len(userIDs) == 0 {
		return nil
	}
	rows := make([]MarketingDigestRecipient, 0, len(userIDs))
	for _, id := range userIDs {
		if id <= 0 {
			continue
		}
		rows = append(rows, MarketingDigestRecipient{
			DigestID:  digestID,
			UserID:    id,
			Status:    MarketingDigestRecipientPending,
			UpdatedAt: common.GetTimestamp(),
		})
	}
	if len(rows) == 0 {
		return nil
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		var existing int64
		if err := tx.Model(&MarketingDigestRecipient{}).Where("digest_id = ?", digestID).Count(&existing).Error; err != nil {
			return err
		}
		if existing > 0 {
			return errors.New("newsletter recipients have already been snapshotted")
		}
		return tx.CreateInBatches(&rows, 500).Error
	})
}

func DeleteMarketingDigestRecipients(digestID int64) error {
	return DB.Where("digest_id = ?", digestID).Delete(&MarketingDigestRecipient{}).Error
}

func ClaimMarketingDigestRecipients(digestID int64, limit int) ([]MarketingDigestRecipient, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	var claimed []MarketingDigestRecipient
	err := DB.Transaction(func(tx *gorm.DB) error {
		var pending []MarketingDigestRecipient
		if err := tx.Where("digest_id = ? AND status = ?", digestID, MarketingDigestRecipientPending).
			Order("id ASC").Limit(limit).Find(&pending).Error; err != nil {
			return err
		}
		for _, recipient := range pending {
			result := tx.Model(&MarketingDigestRecipient{}).
				Where("id = ? AND status = ?", recipient.ID, MarketingDigestRecipientPending).
				Updates(map[string]interface{}{"status": MarketingDigestRecipientSending, "updated_at": common.GetTimestamp()})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 1 {
				recipient.Status = MarketingDigestRecipientSending
				claimed = append(claimed, recipient)
			}
		}
		return nil
	})
	return claimed, err
}

func FinishMarketingDigestRecipient(id int64, status, emailMasked, reason string) error {
	return DB.Model(&MarketingDigestRecipient{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":       status,
		"email_masked": emailMasked,
		"reason":       reason,
		"updated_at":   common.GetTimestamp(),
	}).Error
}

func FinishStaleMarketingDigestRecipients(digestID int64) error {
	return DB.Model(&MarketingDigestRecipient{}).
		Where("digest_id = ? AND status = ?", digestID, MarketingDigestRecipientSending).
		Updates(map[string]interface{}{
			"status":     MarketingDigestRecipientFailed,
			"reason":     "delivery outcome was not recorded; retry may send a duplicate",
			"updated_at": common.GetTimestamp(),
		}).Error
}

func RetryFailedMarketingDigestRecipients(digestID int64) (int64, error) {
	result := DB.Model(&MarketingDigestRecipient{}).
		Where("digest_id = ? AND status = ?", digestID, MarketingDigestRecipientFailed).
		Updates(map[string]interface{}{"status": MarketingDigestRecipientPending, "reason": "", "updated_at": common.GetTimestamp()})
	return result.RowsAffected, result.Error
}

type MarketingDigestRecipientCounts struct {
	Total   int64 `json:"total"`
	Sent    int64 `json:"sent"`
	Skipped int64 `json:"skipped"`
	Failed  int64 `json:"failed"`
	Pending int64 `json:"pending"`
}

func GetMarketingDigestRecipientCounts(digestID int64) (MarketingDigestRecipientCounts, error) {
	var rows []struct {
		Status string
		Count  int64
	}
	if err := DB.Model(&MarketingDigestRecipient{}).
		Select("status, COUNT(*) AS count").Where("digest_id = ?", digestID).
		Group("status").Scan(&rows).Error; err != nil {
		return MarketingDigestRecipientCounts{}, err
	}
	var counts MarketingDigestRecipientCounts
	for _, row := range rows {
		counts.Total += row.Count
		switch row.Status {
		case MarketingDigestRecipientSent:
			counts.Sent = row.Count
		case MarketingDigestRecipientSkipped:
			counts.Skipped = row.Count
		case MarketingDigestRecipientFailed:
			counts.Failed = row.Count
		case MarketingDigestRecipientPending, MarketingDigestRecipientSending:
			counts.Pending += row.Count
		}
	}
	return counts, nil
}

func ListMarketingDigestRecipients(digestID int64, status string, limit, offset int) ([]MarketingDigestRecipient, int64, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	query := DB.Model(&MarketingDigestRecipient{}).Where("digest_id = ?", digestID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []MarketingDigestRecipient
	err := query.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error
	return rows, total, err
}

func ListMarketingDigestPromotions() ([]MarketingDigestPromotion, error) {
	var promotions []MarketingDigestPromotion
	err := DB.Order("updated_at DESC").Find(&promotions).Error
	return promotions, err
}

func GetActiveMarketingDigestPromotions(startAt, endAt int64) ([]MarketingDigestPromotion, error) {
	var promotions []MarketingDigestPromotion
	err := DB.Where("enabled = ? AND (starts_at = 0 OR starts_at <= ?) AND (ends_at = 0 OR ends_at >= ?)", true, endAt, startAt).
		Order("id ASC").Find(&promotions).Error
	return promotions, err
}

func SaveMarketingDigestPromotion(promotion *MarketingDigestPromotion) error {
	promotion.UpdatedAt = common.GetTimestamp()
	if promotion.ID == 0 {
		return DB.Create(promotion).Error
	}
	return DB.Model(&MarketingDigestPromotion{}).Where("id = ?", promotion.ID).Updates(map[string]interface{}{
		"model_name": promotion.ModelName, "description": promotion.Description, "discount": promotion.Discount,
		"starts_at": promotion.StartsAt, "ends_at": promotion.EndsAt, "url": promotion.URL,
		"enabled": promotion.Enabled, "updated_at": promotion.UpdatedAt,
	}).Error
}

func DeleteMarketingDigestPromotion(id int64) error {
	return DB.Delete(&MarketingDigestPromotion{}, id).Error
}

func GetMarketingDigestAudienceUserIDs(audienceType, group string, userIDs []int, activeSince int64, max int) ([]int, error) {
	if max <= 0 {
		max = 100000
	}
	query := DB.Model(&User{}).Where("status = ?", common.UserStatusEnabled)
	switch audienceType {
	case "all":
	case "topup":
		paidUsers := DB.Model(&TopUp{}).Select("user_id").Where("organization_id = 0 AND status = ?", common.TopUpStatusSuccess).Distinct()
		query = query.Where("id IN (?)", paidUsers)
	case "active":
		query = query.Where("last_login_at >= ?", activeSince)
	case "group":
		if group == "" {
			return nil, errors.New("audience group is required")
		}
		query = query.Where(commonGroupCol+" = ?", group)
	case "manual":
		if len(userIDs) == 0 {
			return nil, errors.New("select at least one recipient")
		}
		query = query.Where("id IN ?", userIDs)
	default:
		return nil, errors.New("invalid audience type")
	}
	var result []int
	err := query.Order("id ASC").Limit(max + 1).Pluck("id", &result).Error
	if err != nil {
		return nil, err
	}
	if len(result) > max {
		return nil, errors.New("newsletter audience exceeds recipient limit")
	}
	return result, nil
}

func UpdateMarketingDigestEmailPreference(userID int, enabled bool) error {
	user, err := GetUserById(userID, true)
	if err != nil {
		return err
	}
	setting := user.GetSetting()
	setting.MarketingEmailEnabled = &enabled
	user.SetSetting(setting)
	return user.Update(false)
}

func IsMarketingDigestUserEligible(user *User) bool {
	if user == nil || user.Status != common.UserStatusEnabled || user.DeletedAt.Valid {
		return false
	}
	setting := user.GetSetting()
	return setting.MarketingEmailEnabled == nil || *setting.MarketingEmailEnabled
}

func MarketingDigestActiveSince(now time.Time) int64 {
	return now.Add(-30 * 24 * time.Hour).Unix()
}
