package model

import (
	"errors"

	"github.com/QuantumNous/new-api/common"
)

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
	if err := query.Order("id ASC").Limit(max + 1).Pluck("id", &result).Error; err != nil {
		return nil, err
	}
	if len(result) > max {
		return nil, errors.New("marketing email audience exceeds recipient limit")
	}
	return result, nil
}

func UpdateMarketingEmailPreference(userID int, enabled bool) error {
	user, err := GetUserById(userID, true)
	if err != nil {
		return err
	}
	setting := user.GetSetting()
	setting.MarketingEmailEnabled = &enabled
	user.SetSetting(setting)
	return user.Update(false)
}

func IsMarketingEmailUserEligible(user *User) bool {
	if user == nil || user.Status != common.UserStatusEnabled || user.DeletedAt.Valid {
		return false
	}
	setting := user.GetSetting()
	return setting.MarketingEmailEnabled == nil || *setting.MarketingEmailEnabled
}
