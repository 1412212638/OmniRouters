package model

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const ExternalIdentityProviderTelegram = "telegram"

var ErrExternalIdentityAlreadyClaimed = errors.New("external identity is already claimed")

// ExternalIdentityClaim provides a durable, provider-independent ownership
// record while legacy provider columns remain readable during migration.
type ExternalIdentityClaim struct {
	Id        int64     `json:"id" gorm:"primaryKey"`
	Provider  string    `json:"provider" gorm:"type:varchar(32);not null;uniqueIndex:idx_external_identity_subject,priority:1;uniqueIndex:idx_external_identity_user,priority:1"`
	Subject   string    `json:"subject" gorm:"type:varchar(128);not null;uniqueIndex:idx_external_identity_subject,priority:2"`
	UserId    int       `json:"user_id" gorm:"not null;index;uniqueIndex:idx_external_identity_user,priority:2"`
	CreatedAt time.Time `json:"created_at"`
}

func (ExternalIdentityClaim) TableName() string { return "external_identity_claims" }

func ClaimExternalIdentityWithTx(tx *gorm.DB, provider, subject string, userId int) error {
	provider, subject = strings.TrimSpace(provider), strings.TrimSpace(subject)
	if tx == nil || provider == "" || subject == "" || userId <= 0 {
		return errors.New("external identity claim is invalid")
	}
	if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&ExternalIdentityClaim{Provider: provider, Subject: subject, UserId: userId}).Error; err != nil {
		return err
	}
	var owner ExternalIdentityClaim
	if err := tx.Where("provider = ? AND subject = ?", provider, subject).First(&owner).Error; err != nil {
		return err
	}
	if owner.UserId != userId {
		return ErrExternalIdentityAlreadyClaimed
	}
	return nil
}

// BindExternalIdentityWithTx claims a provider subject and updates the legacy
// Telegram column in one transaction. The legacy check remains necessary for
// installations upgraded before external_identity_claims was introduced.
func BindExternalIdentityWithTx(tx *gorm.DB, provider, subject string, userId int) error {
	provider, subject = strings.TrimSpace(provider), strings.TrimSpace(subject)
	if tx == nil || userId <= 0 || provider == "" || subject == "" {
		return errors.New("external identity binding is invalid")
	}
	if provider != ExternalIdentityProviderTelegram {
		return errors.New("unsupported external identity provider")
	}
	var legacy User
	if err := tx.Where("telegram_id = ?", subject).First(&legacy).Error; err == nil && legacy.Id != userId {
		return ErrExternalIdentityAlreadyClaimed
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if err := ClaimExternalIdentityWithTx(tx, provider, subject, userId); err != nil {
		return err
	}
	return tx.Model(&User{}).Where("id = ?", userId).Update("telegram_id", subject).Error
}

// InitializeExternalIdentityClaims backfills legacy Telegram bindings before
// any future OAuth flow starts using the claim table.
func InitializeExternalIdentityClaims() error {
	var users []User
	if err := DB.Select("id", "telegram_id").Where("telegram_id <> ?", "").Find(&users).Error; err != nil {
		return err
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		for _, user := range users {
			if err := ClaimExternalIdentityWithTx(tx, ExternalIdentityProviderTelegram, user.TelegramId, user.Id); err != nil {
				return err
			}
		}
		return nil
	})
}
