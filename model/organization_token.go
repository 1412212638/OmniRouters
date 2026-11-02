package model

import (
	"errors"
	"fmt"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"gorm.io/gorm"
)

func GetOrganizationTokens(organizationId int) ([]*Token, error) {
	var tokens []*Token
	err := DB.Where("organization_id = ?", organizationId).Order("id desc").Find(&tokens).Error
	return tokens, err
}

func GetOrganizationToken(organizationId, tokenId int) (*Token, error) {
	var token Token
	err := DB.Where("organization_id = ? AND id = ?", organizationId, tokenId).First(&token).Error
	return &token, err
}

func CreateOrganizationToken(organizationId, userId int, name string) (*Token, error) {
	name = strings.TrimSpace(name)
	if organizationId <= 0 || userId <= 0 || (name == "" || len(name) > 50) {
		return nil, errors.New("invalid organization token arguments")
	}
	member, err := GetOrganizationMember(organizationId, userId)
	if err != nil || member.Status != OrganizationMemberActive {
		return nil, errors.New("organization membership is not active")
	}
	organization, err := GetOrganizationById(organizationId)
	if err != nil {
		return nil, err
	}
	var count int64
	if err := DB.Model(&Token{}).Where("organization_id = ? AND user_id = ?", organizationId, userId).Count(&count).Error; err != nil {
		return nil, err
	}
	if count >= int64(operation_setting.GetMaxUserTokens()) {
		return nil, errors.New("organization token count limit reached")
	}
	key, err := common.GenerateKey()
	if err != nil {
		return nil, err
	}
	token := &Token{UserId: userId, OrganizationId: organizationId, Name: name, Key: key,
		Status: common.TokenStatusEnabled, CreatedTime: common.GetTimestamp(), AccessedTime: 0,
		ExpiredTime: -1, UnlimitedQuota: true, Group: organization.Group, KeyType: TokenKeyTypeGeneral,
		AllowUserQuotaQuery: false}
	if err := token.Insert(); err != nil {
		return nil, err
	}
	return token, nil
}

func UpdateOrganizationToken(organizationId int, tokenId int, name *string, status *int) (*Token, error) {
	token, err := GetOrganizationToken(organizationId, tokenId)
	if err != nil {
		return nil, err
	}
	if name != nil {
		value := strings.TrimSpace(*name)
		if value == "" {
			return nil, errors.New("token name cannot be empty")
		}
		token.Name = value
	}
	if status != nil {
		if *status != common.TokenStatusEnabled && *status != common.TokenStatusDisabled {
			return nil, errors.New("invalid token status")
		}
		if *status == common.TokenStatusEnabled {
			member, membershipErr := GetOrganizationMember(organizationId, token.UserId)
			if membershipErr != nil || member.Status != OrganizationMemberActive {
				return nil, errors.New("token creator is no longer an active organization member")
			}
		}
		token.Status = *status
	}
	if err := invalidateTokenCacheForMutation(token.Key); err != nil {
		return nil, err
	}
	if err := DB.Model(&Token{}).Where("id = ? AND organization_id = ?", tokenId, organizationId).Updates(map[string]interface{}{"name": token.Name, "status": token.Status}).Error; err != nil {
		return nil, err
	}
	return token, nil
}

func DeleteOrganizationToken(organizationId, tokenId int) error {
	token, err := GetOrganizationToken(organizationId, tokenId)
	if err != nil {
		return err
	}
	return token.Delete()
}

func RemoveOrganizationMembership(organizationId, actorUserId, userId int) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		if _, err := lockOrganization(tx, organizationId); err != nil {
			return err
		}
		var actor OrganizationMember
		if err := tx.Where("organization_id = ? AND user_id = ? AND status = ?", organizationId, actorUserId, OrganizationMemberActive).First(&actor).Error; err != nil {
			return err
		}
		var member OrganizationMember
		if err := tx.Where("organization_id = ? AND user_id = ? AND status = ?", organizationId, userId, OrganizationMemberActive).First(&member).Error; err != nil {
			return err
		}
		if member.Role == OrganizationMemberOwner {
			return errors.New("organization owner cannot be removed")
		}
		if actorUserId != userId && (!OrganizationCanManage(actor.Role) || (member.Role == OrganizationMemberAdmin && actor.Role != OrganizationMemberOwner)) {
			return errors.New("organization member management permission required")
		}
		var tokens []Token
		if err := tx.Where("organization_id = ? AND user_id = ?", organizationId, userId).Find(&tokens).Error; err != nil {
			return err
		}
		if err := invalidateTokensCache(tokens); err != nil {
			return err
		}
		if err := tx.Model(&member).Updates(map[string]interface{}{"status": OrganizationMemberRemoved, "updated_at": common.GetTimestamp()}).Error; err != nil {
			return err
		}
		return tx.Model(&Token{}).Where("organization_id = ? AND user_id = ?", organizationId, userId).Update("status", common.TokenStatusDisabled).Error
	})
}

func TransferUserQuotaToOrganization(organizationId, userId, quota int) error {
	if quota <= 0 {
		return errors.New("quota must be positive")
	}
	if err := common.ValidateWalletQuota(quota); err != nil {
		return err
	}
	member, err := GetOrganizationMember(organizationId, userId)
	if err != nil || member.Status != OrganizationMemberActive || !OrganizationCanManage(member.Role) {
		return errors.New("organization manager permission required")
	}
	// Reserve only the live cache balance here. The durable debit belongs to
	// the same transaction as the organization credit and its wallet ledger;
	// TryReserveUserQuota would persist/enqueue the debit before that transaction.
	cacheReserved, err := reserveOrganizationTransferCache(userId, quota)
	if err != nil {
		return err
	}
	err = DB.Transaction(func(tx *gorm.DB) error {
		if _, err := lockOrganization(tx, organizationId); err != nil {
			return err
		}
		var activeMember OrganizationMember
		if err := tx.Where("organization_id = ? AND user_id = ? AND status = ?", organizationId, userId, OrganizationMemberActive).First(&activeMember).Error; err != nil {
			return err
		}
		if !OrganizationCanManage(activeMember.Role) {
			return errors.New("organization manager permission required")
		}
		personalDebit := tx.Model(&User{}).Where("id = ? AND quota >= ?", userId, quota).
			Update("quota", gorm.Expr("quota - ?", quota))
		if personalDebit.Error != nil {
			return personalDebit.Error
		}
		if personalDebit.RowsAffected != 1 {
			return errors.New("personal quota is insufficient")
		}
		result := tx.Model(&Organization{}).Where("id = ? AND quota <= ?", organizationId, common.MaxWalletQuota-quota).Updates(map[string]interface{}{"quota": gorm.Expr("quota + ?", quota), "updated_at": common.GetTimestamp()})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return errors.New("organization balance limit reached")
		}
		var organization Organization
		if err := tx.First(&organization, organizationId).Error; err != nil {
			return err
		}
		return AddOrganizationWalletTransaction(tx, organizationId, userId, quota, organization.Quota, "transfer_in", "Personal wallet transfer")
	})
	if err != nil && cacheReserved {
		result, refundErr := cacheApplyUserQuotaDelta(userId, int64(quota))
		if refundErr != nil || result != cacheQuotaOK {
			common.SysError(fmt.Sprintf("organization transfer cache release failed: user=%d result=%d error=%v", userId, result, refundErr))
		}
	}
	return err
}

// reserveOrganizationTransferCache never falls back to a stale database
// balance when Redis is configured. Pending request batches may already have
// consumed that balance even though their database deltas have not flushed.
func reserveOrganizationTransferCache(userId, quota int) (bool, error) {
	if !common.RedisEnabled {
		return false, nil
	}
	result, err := cacheTryReserveUserQuota(userId, int64(quota))
	if err != nil {
		return false, errors.New("personal balance is temporarily unavailable; please retry")
	}
	if result == cacheQuotaMiss {
		// A missing cache cannot be safely rebuilt from DB during batching:
		// another process may hold unflushed debits. Require a live balance.
		if common.BatchUpdateEnabled {
			return false, errors.New("personal balance is temporarily unavailable; please refresh and retry")
		}
		user, loadErr := GetUserById(userId, false)
		if loadErr != nil {
			return false, loadErr
		}
		if err = writeUserCache(user.ToBaseUser(), true); err != nil {
			return false, errors.New("personal balance is temporarily unavailable; please retry")
		}
		result, err = cacheTryReserveUserQuota(userId, int64(quota))
	}
	if err != nil || result == cacheQuotaMiss {
		return false, errors.New("personal balance is temporarily unavailable; please retry")
	}
	if result != cacheQuotaOK {
		return false, errors.New("personal quota is insufficient")
	}
	return true, nil
}
