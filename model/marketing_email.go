package model

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func GetMarketingSMTPConfig() (common.MarketingSMTPConfig, error) {
	var option Option
	err := DB.Where(&Option{Key: common.MarketingSMTPOption}).Take(&option).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return common.DecodeMarketingSMTPConfig("")
	}
	if err != nil {
		return common.MarketingSMTPConfig{}, err
	}
	return common.DecodeMarketingSMTPConfig(option.Value)
}

type MarketingEmailSuppression struct {
	AddressHash string `gorm:"type:varchar(64);primaryKey"`
	Reason      string `gorm:"type:varchar(32)"`
	CreatedAt   int64
}

func marketingEmailAddressHash(address string) string {
	hash := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(address))))
	return hex.EncodeToString(hash[:])
}

func IsMarketingEmailSuppressed(address string) (bool, error) {
	var count int64
	err := DB.Model(&MarketingEmailSuppression{}).Where("address_hash = ?", marketingEmailAddressHash(address)).Count(&count).Error
	return count != 0, err
}

func SuppressMarketingEmail(address, reason string) error {
	return DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&MarketingEmailSuppression{
		AddressHash: marketingEmailAddressHash(address), Reason: reason,
	}).Error
}

func GetMarketingEmailAudienceUserIDs(audienceType, group string, userIDs []int, max int) ([]int, error) {
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
		query = query.Where("last_login_at >= ?", common.GetTimestamp()-30*24*60*60)
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
	if err := query.Order("id ASC").Limit(max+1).Pluck("id", &result).Error; err != nil {
		return nil, err
	}
	if len(result) > max {
		return nil, errors.New("marketing email audience exceeds recipient limit")
	}
	return result, nil
}

func UpdateMarketingEmailPreference(userID int, enabled bool) error {
	user, err := GetUserById(userID, false)
	if err != nil {
		return err
	}
	setting := user.GetSetting()
	setting.MarketingEmailEnabled = &enabled
	return UpdateUserSetting(userID, setting)
}

func IsMarketingEmailUserEligible(user *User) bool {
	if user == nil || user.Status != common.UserStatusEnabled || user.DeletedAt.Valid {
		return false
	}
	setting := user.GetSetting()
	return setting.MarketingEmailEnabled == nil || *setting.MarketingEmailEnabled
}
