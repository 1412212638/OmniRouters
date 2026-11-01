package model

import (
	"errors"
	"time"
)

type RealNameVerification struct {
	ID                         int    `json:"id" gorm:"primaryKey"`
	UserID                     int    `json:"user_id" gorm:"uniqueIndex;not null"`
	Type                       string `json:"type" gorm:"type:varchar(16);not null"`
	Status                     string `json:"status" gorm:"type:varchar(16);not null"`
	PersonalNameCipher         string `json:"-" gorm:"type:text"`
	PersonalIDCardCipher       string `json:"-" gorm:"type:text"`
	EnterpriseNameCipher       string `json:"-" gorm:"type:text"`
	RegistrationNumberCipher   string `json:"-" gorm:"type:text"`
	RepresentativeNameCipher   string `json:"-" gorm:"type:text"`
	RepresentativeIDCardCipher string `json:"-" gorm:"type:text"`
	DisplayName                string `json:"display_name" gorm:"type:varchar(64)"`
	IDCardSuffix               string `json:"id_card_suffix" gorm:"type:varchar(8)"`
	CompanyNameMasked          string `json:"company_name_masked" gorm:"type:varchar(128)"`
	RegistrationSuffix         string `json:"registration_suffix" gorm:"type:varchar(8)"`
	Provider                   string `json:"provider" gorm:"type:varchar(32)"`
	UpstreamReference          string `json:"upstream_reference" gorm:"type:varchar(128)"`
	VerifiedAt                 int64  `json:"verified_at"`
	CreatedAt                  int64  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt                  int64  `json:"updated_at" gorm:"autoUpdateTime"`
}

func (RealNameVerification) TableName() string { return "real_name_verifications" }

func GetRealNameVerification(userID int) (*RealNameVerification, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user id")
	}
	var verification RealNameVerification
	err := DB.Where("user_id = ?", userID).First(&verification).Error
	if err != nil {
		return nil, err
	}
	return &verification, nil
}

func HasVerifiedRealName(userID int) (bool, error) {
	var count int64
	err := DB.Model(&RealNameVerification{}).Where("user_id = ? AND status = ?", userID, "verified").Count(&count).Error
	return count > 0, err
}

func SaveRealNameVerification(verification *RealNameVerification) error {
	if verification == nil || verification.UserID <= 0 || verification.Type == "" || verification.Status == "" {
		return errors.New("invalid real name verification")
	}
	if verification.CreatedAt == 0 {
		verification.CreatedAt = time.Now().Unix()
	}
	verification.UpdatedAt = time.Now().Unix()
	return DB.Where("user_id = ?", verification.UserID).Assign(verification).FirstOrCreate(verification).Error
}

func ResetRealNameVerification(userID int) error {
	if userID <= 0 {
		return errors.New("invalid user id")
	}
	return DB.Where("user_id = ?", userID).Delete(&RealNameVerification{}).Error
}
