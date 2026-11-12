package model

import (
	"errors"
	"fmt"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

// ValidateOrganizationTopUp checks permission before creating an external checkout.
// Callbacks use the order's saved organization ID, never the browser workspace.
func ValidateOrganizationTopUp(organizationId, userId, quota int) error {
	if !common.OrganizationEnabled {
		return errors.New("organization workspaces are disabled")
	}
	member, err := GetOrganizationMember(organizationId, userId)
	if err != nil || member.Status != OrganizationMemberActive || !OrganizationCanManage(member.Role) {
		return errors.New("organization manager permission required")
	}
	maxCurrent, err := topUpQuotaMaxCurrent(quota)
	if err != nil {
		return err
	}
	organization, err := GetOrganizationById(organizationId)
	if err != nil {
		return err
	}
	if organization.Quota > maxCurrent {
		return ErrTopUpQuotaLimitExceeded
	}
	return nil
}

func creditTopUpWallet(tx *gorm.DB, topUp *TopUp, quota int, updates map[string]interface{}) error {
	if topUp.OrganizationId == 0 {
		if err := creditTopUpQuota(tx, topUp.UserId, quota, updates); err != nil {
			return err
		}
		return RecordWalletFundingTx(tx, WalletFundingEntry{UserId: topUp.UserId, Amount: int64(quota), Kind: FundingPaid, Source: "topup", EventKey: fmt.Sprintf("topup:%d", topUp.Id), Reference: topUp.TradeNo})
	}
	maxCurrent, err := topUpQuotaMaxCurrent(quota)
	if err != nil {
		return err
	}
	organization, err := lockOrganization(tx, topUp.OrganizationId)
	if err != nil {
		return err
	}
	// A checkout paid after a feature toggle or membership change still belongs
	// to its original organization and must never silently credit a personal wallet.
	balanceAfter := organization.Quota + quota
	result := tx.Model(&Organization{}).Where("id = ? AND quota <= ?", organization.Id, maxCurrent).Updates(map[string]interface{}{
		"quota": gorm.Expr("quota + ?", quota), "updated_at": common.GetTimestamp(),
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return ErrTopUpQuotaLimitExceeded
	}
	return AddOrganizationWalletTransaction(tx, topUp.OrganizationId, topUp.UserId, quota, balanceAfter, "topup", topUp.TradeNo)
}

func syncTopUpWalletCache(topUp *TopUp, quota int, reason string) {
	if topUp.OrganizationId == 0 {
		syncCreditUserQuotaCache(topUp.UserId, quota, reason)
	}
}

func GetOrganizationTopUps(organizationId int, keyword string, page *common.PageInfo) ([]*TopUp, int64, error) {
	query := DB.Model(&TopUp{}).Where("organization_id = ?", organizationId)
	if keyword != "" {
		pattern, err := sanitizeLikePattern(keyword)
		if err != nil {
			return nil, 0, err
		}
		query = query.Where("trade_no LIKE ? ESCAPE '!'", pattern)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []*TopUp
	if err := query.Order("id desc").Offset(page.GetStartIdx()).Limit(page.GetPageSize()).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	if err := attachTopUpUsernames(rows); err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
