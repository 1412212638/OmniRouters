package model

import (
	"errors"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

// UpdateOrganizationMember serializes role and budget changes with ownership
// transfers so a stale manager check cannot demote the newly assigned owner.
func UpdateOrganizationMember(organizationId, actorUserId, targetUserId int, role string, monthlyQuotaLimit *int) error {
	if role != "" && role != OrganizationMemberAdmin && role != OrganizationMemberMember {
		return errors.New("invalid organization role")
	}
	if monthlyQuotaLimit != nil && (*monthlyQuotaLimit < 0 || *monthlyQuotaLimit > common.MaxWalletQuota) {
		return errors.New("invalid monthly quota limit")
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		if _, err := lockOrganization(tx, organizationId); err != nil {
			return err
		}
		var actor, target OrganizationMember
		if err := tx.Where("organization_id = ? AND user_id = ? AND status = ?", organizationId, actorUserId, OrganizationMemberActive).First(&actor).Error; err != nil {
			return err
		}
		if !OrganizationCanManage(actor.Role) {
			return errors.New("organization manager permission required")
		}
		if err := tx.Where("organization_id = ? AND user_id = ? AND status = ?", organizationId, targetUserId, OrganizationMemberActive).First(&target).Error; err != nil {
			return err
		}
		if target.Role == OrganizationMemberOwner {
			return errors.New("organization owner cannot be changed")
		}
		if actor.Role != OrganizationMemberOwner && (target.Role == OrganizationMemberAdmin || role == OrganizationMemberAdmin) {
			return errors.New("only owner can manage admins")
		}
		updates := map[string]interface{}{"updated_at": common.GetTimestamp()}
		if role != "" {
			updates["role"] = role
		}
		if monthlyQuotaLimit != nil {
			updates["monthly_quota_limit"] = *monthlyQuotaLimit
		}
		return tx.Model(&target).Updates(updates).Error
	})
}
