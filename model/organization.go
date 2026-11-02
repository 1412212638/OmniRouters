package model

import (
	"errors"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	OrganizationMemberOwner        = "owner"
	OrganizationMemberAdmin        = "admin"
	OrganizationMemberMember       = "member"
	OrganizationMemberActive       = "active"
	OrganizationMemberRemoved      = "removed"
	OrganizationInvitationPending  = "pending"
	OrganizationInvitationAccepted = "accepted"
	OrganizationInvitationRevoked  = "revoked"
)

type Organization struct {
	Id           int    `json:"id"`
	Name         string `json:"name" gorm:"type:varchar(128);not null"`
	OwnerUserId  int    `json:"owner_user_id" gorm:"index;not null"`
	Group        string `json:"group" gorm:"type:varchar(64);default:'default'"`
	Quota        int    `json:"quota" gorm:"default:0"`
	UsedQuota    int    `json:"used_quota" gorm:"default:0"`
	RequestCount int    `json:"request_count" gorm:"default:0"`
	CreatedAt    int64  `json:"created_at" gorm:"bigint"`
	UpdatedAt    int64  `json:"updated_at" gorm:"bigint"`
}

type OrganizationMember struct {
	Id                int    `json:"id"`
	OrganizationId    int    `json:"organization_id" gorm:"uniqueIndex:idx_org_member"`
	UserId            int    `json:"user_id" gorm:"uniqueIndex:idx_org_member"`
	Role              string `json:"role" gorm:"type:varchar(16);default:'member'"`
	Status            string `json:"status" gorm:"type:varchar(16);index;default:'active'"`
	MonthlyQuotaLimit int    `json:"monthly_quota_limit" gorm:"default:0"`
	CreatedAt         int64  `json:"created_at" gorm:"bigint"`
	UpdatedAt         int64  `json:"updated_at" gorm:"bigint"`
}

// OrganizationMemberUsage keeps the month in the row so async task settlement
// cannot accidentally charge a newer month after a request crosses a boundary.
type OrganizationMemberUsage struct {
	Id             int    `json:"id"`
	OrganizationId int    `json:"organization_id" gorm:"uniqueIndex:idx_org_usage_period"`
	UserId         int    `json:"user_id" gorm:"uniqueIndex:idx_org_usage_period"`
	Period         string `json:"period" gorm:"type:varchar(7);uniqueIndex:idx_org_usage_period"`
	UsedQuota      int    `json:"used_quota" gorm:"default:0"`
	UpdatedAt      int64  `json:"updated_at" gorm:"bigint"`
}

type OrganizationInvitation struct {
	Id                int    `json:"id"`
	OrganizationId    int    `json:"organization_id" gorm:"index;not null"`
	Email             string `json:"email" gorm:"type:varchar(256);index;not null"`
	Role              string `json:"role" gorm:"type:varchar(16);default:'member'"`
	InvitedBy         int    `json:"invited_by" gorm:"index;not null"`
	Status            string `json:"status" gorm:"type:varchar(16);index;default:'pending'"`
	ExpiresAt         int64  `json:"expires_at" gorm:"bigint"`
	ReminderSentAt    int64  `json:"-" gorm:"bigint;default:0"`
	ReminderClaimedAt int64  `json:"-" gorm:"bigint;default:0"`
	CreatedAt         int64  `json:"created_at" gorm:"bigint"`
}

type OrganizationWalletTransaction struct {
	Id             int    `json:"id"`
	OrganizationId int    `json:"organization_id" gorm:"index;not null"`
	UserId         int    `json:"user_id" gorm:"index;not null"`
	Amount         int    `json:"amount"`
	BalanceAfter   int    `json:"balance_after"`
	Type           string `json:"type" gorm:"type:varchar(32);index"`
	Description    string `json:"description" gorm:"type:text"`
	CreatedAt      int64  `json:"created_at" gorm:"bigint;index"`
}

func OrganizationUsagePeriod(at time.Time) string { return at.UTC().Format("2006-01") }
func lockOrganization(tx *gorm.DB, id int) (*Organization, error) {
	var organization Organization
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&organization, id).Error; err != nil {
		return nil, err
	}
	return &organization, nil
}
func normalizedOrganizationPeriod(period string) (string, error) {
	if period == "" {
		period = CurrentOrganizationUsagePeriod()
	}
	if _, err := time.Parse("2006-01", period); err != nil {
		return "", errors.New("invalid organization usage period")
	}
	return period, nil
}
func CurrentOrganizationUsagePeriod() string { return OrganizationUsagePeriod(time.Now()) }

func GetOrganizationById(id int) (*Organization, error) {
	if id <= 0 {
		return nil, errors.New("organization id is invalid")
	}
	var organization Organization
	if err := DB.First(&organization, id).Error; err != nil {
		return nil, err
	}
	return &organization, nil
}

func GetOrganizationMember(organizationId, userId int) (*OrganizationMember, error) {
	if organizationId <= 0 || userId <= 0 {
		return nil, errors.New("organization member id is invalid")
	}
	var member OrganizationMember
	err := DB.Where("organization_id = ? AND user_id = ?", organizationId, userId).First(&member).Error
	return &member, err
}

func OrganizationCanManage(role string) bool {
	return role == OrganizationMemberOwner || role == OrganizationMemberAdmin
}

// TryReserveOrganizationQuota atomically reserves organization balance and the
// member's current-month budget. period is captured at request submission.
func TryReserveOrganizationQuota(organizationId, userId, quota int, period string) (bool, error) {
	if quota < 0 {
		return false, errors.New("quota cannot be negative")
	}
	if quota > common.MaxWalletQuota {
		return false, errors.New("quota limit exceeded")
	}
	var err error
	period, err = normalizedOrganizationPeriod(strings.TrimSpace(period))
	if err != nil {
		return false, err
	}
	err = DB.Transaction(func(tx *gorm.DB) error {
		organization, err := lockOrganization(tx, organizationId)
		if err != nil {
			return err
		}
		var member OrganizationMember
		if err := tx.Where("organization_id = ? AND user_id = ? AND status = ?", organizationId, userId, OrganizationMemberActive).First(&member).Error; err != nil {
			return err
		}
		var usage OrganizationMemberUsage
		result := tx.Where("organization_id = ? AND user_id = ? AND period = ?", organizationId, userId, period).First(&usage)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			usage = OrganizationMemberUsage{OrganizationId: organizationId, UserId: userId, Period: period, UpdatedAt: common.GetTimestamp()}
			if err := tx.Create(&usage).Error; err != nil {
				return err
			}
		} else if result.Error != nil {
			return result.Error
		}
		if quota == 0 {
			return nil
		}
		if organization.Quota < quota || (member.MonthlyQuotaLimit > 0 && usage.UsedQuota > member.MonthlyQuotaLimit-quota) {
			return gorm.ErrRecordNotFound
		}
		result = tx.Model(&Organization{}).Where("id = ? AND quota >= ?", organizationId, quota).Updates(map[string]interface{}{"quota": gorm.Expr("quota - ?", quota), "used_quota": gorm.Expr("used_quota + ?", quota), "updated_at": common.GetTimestamp()})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}
		query := tx.Model(&usage)
		if member.MonthlyQuotaLimit > 0 {
			query = query.Where("used_quota <= ?", member.MonthlyQuotaLimit-quota)
		}
		result = query.Updates(map[string]interface{}{"used_quota": gorm.Expr("used_quota + ?", quota), "updated_at": common.GetTimestamp()})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return err == nil, err
}

// AdjustOrganizationQuota applies only the difference from a successful
// reservation. Settlement remains payable even if the organization balance is
// now insufficient; the next reservation is rejected until debt is repaid.
func AdjustOrganizationQuota(organizationId, userId, delta int, period string) error {
	if delta == 0 {
		return nil
	}
	if delta > common.MaxWalletQuota || delta < -common.MaxWalletQuota {
		return errors.New("quota adjustment limit exceeded")
	}
	var err error
	period, err = normalizedOrganizationPeriod(strings.TrimSpace(period))
	if err != nil {
		return err
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		if _, err := lockOrganization(tx, organizationId); err != nil {
			return err
		}
		var usage OrganizationMemberUsage
		result := tx.Where("organization_id = ? AND user_id = ? AND period = ?", organizationId, userId, period).First(&usage)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			usage = OrganizationMemberUsage{OrganizationId: organizationId, UserId: userId, Period: period}
			if err := tx.Create(&usage).Error; err != nil {
				return err
			}
		} else if result.Error != nil {
			return result.Error
		}
		if delta > 0 {
			result = tx.Model(&Organization{}).Where("id = ?", organizationId).Updates(map[string]interface{}{"quota": gorm.Expr("quota - ?", delta), "used_quota": gorm.Expr("used_quota + ?", delta), "updated_at": common.GetTimestamp()})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected != 1 {
				return gorm.ErrRecordNotFound
			}
			return tx.Model(&usage).Updates(map[string]interface{}{"used_quota": gorm.Expr("used_quota + ?", delta), "updated_at": common.GetTimestamp()}).Error
		}
		refund := -delta
		result = tx.Model(&Organization{}).Where("id = ? AND quota <= ?", organizationId, common.MaxWalletQuota-refund).Updates(map[string]interface{}{"quota": gorm.Expr("quota + ?", refund), "used_quota": gorm.Expr("CASE WHEN used_quota >= ? THEN used_quota - ? ELSE 0 END", refund, refund), "updated_at": common.GetTimestamp()})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return ErrWalletQuotaLimitExceeded
		}
		return tx.Model(&usage).Updates(map[string]interface{}{"used_quota": gorm.Expr("CASE WHEN used_quota >= ? THEN used_quota - ? ELSE 0 END", refund, refund), "updated_at": common.GetTimestamp()}).Error
	})
}

func AddOrganizationWalletTransaction(tx *gorm.DB, organizationId, userId, amount, balanceAfter int, kind, description string) error {
	return tx.Create(&OrganizationWalletTransaction{OrganizationId: organizationId, UserId: userId, Amount: amount, BalanceAfter: balanceAfter, Type: kind, Description: description, CreatedAt: common.GetTimestamp()}).Error
}

func TransferOrganizationOwnership(organizationId, actorUserId, targetUserId int) error {
	if actorUserId <= 0 || targetUserId <= 0 || actorUserId == targetUserId {
		return errors.New("invalid new organization owner")
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		organization, err := lockOrganization(tx, organizationId)
		if err != nil {
			return err
		}
		if organization.OwnerUserId != actorUserId {
			return errors.New("organization owner permission required")
		}
		var target OrganizationMember
		if err := tx.Where("organization_id = ? AND user_id = ? AND status = ?", organizationId, targetUserId, OrganizationMemberActive).First(&target).Error; err != nil {
			return err
		}
		if err := tx.Model(&OrganizationMember{}).Where("organization_id = ? AND user_id = ?", organizationId, actorUserId).Updates(map[string]interface{}{"role": OrganizationMemberAdmin, "updated_at": common.GetTimestamp()}).Error; err != nil {
			return err
		}
		if err := tx.Model(&target).Updates(map[string]interface{}{"role": OrganizationMemberOwner, "updated_at": common.GetTimestamp()}).Error; err != nil {
			return err
		}
		return tx.Model(organization).Updates(map[string]interface{}{"owner_user_id": targetUserId, "updated_at": common.GetTimestamp()}).Error
	})
}
