package model

import (
	"context"

	"gorm.io/gorm"
)

func GetOrganizationInvitationsDueForReminder(ctx context.Context, now int64, limit int) ([]OrganizationInvitation, error) {
	var invitations []OrganizationInvitation
	err := DB.WithContext(ctx).Where("status = ? AND expires_at > ? AND expires_at <= ? AND reminder_sent_at = 0 AND (reminder_claimed_at = 0 OR reminder_claimed_at < ?)", OrganizationInvitationPending, now, now+86400, now-600).Order("id ASC").Limit(limit).Find(&invitations).Error
	return invitations, err
}

func ClaimOrganizationInvitationReminder(ctx context.Context, id int, now int64) (bool, error) {
	result := DB.WithContext(ctx).Model(&OrganizationInvitation{}).Where("id = ? AND status = ? AND expires_at > ? AND expires_at <= ? AND reminder_sent_at = 0 AND (reminder_claimed_at = 0 OR reminder_claimed_at < ?)", id, OrganizationInvitationPending, now, now+86400, now-600).Update("reminder_claimed_at", now)
	return result.RowsAffected == 1, result.Error
}

func FinishOrganizationInvitationReminder(ctx context.Context, id int, claimAt int64, sent bool) error {
	updates := map[string]any{"reminder_claimed_at": 0}
	if sent {
		updates["reminder_sent_at"] = claimAt
	}
	result := DB.WithContext(ctx).Model(&OrganizationInvitation{}).Where("id = ? AND reminder_claimed_at = ?", id, claimAt).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
