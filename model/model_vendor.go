package model

import "gorm.io/gorm"

// ModelVendor stores the many-to-many relationship between a model and the
// suppliers that can provide it. The legacy Model.VendorID remains the
// primary/display supplier for compatibility.
type ModelVendor struct {
	ID       int    `json:"id"`
	ModelID  int    `json:"model_id" gorm:"not null;uniqueIndex:uk_model_vendor,priority:1"`
	VendorID int    `json:"vendor_id" gorm:"not null;uniqueIndex:uk_model_vendor,priority:2;index"`
	SortOrder int   `json:"sort_order" gorm:"not null;default:0;index"`
	Model    Model  `json:"-" gorm:"foreignKey:ModelID;constraint:OnDelete:CASCADE"`
	Vendor   Vendor `json:"-" gorm:"foreignKey:VendorID;constraint:OnDelete:CASCADE"`
}

func (ModelVendor) TableName() string { return "model_vendors" }

func GetModelVendorIDs(modelID int) ([]int, error) {
	var rows []ModelVendor
	if err := DB.Where("model_id = ?", modelID).Order("sort_order ASC, vendor_id ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	ids := make([]int, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.VendorID)
	}
	return ids, nil
}

func GetModelVendorIDsMap(modelIDs []int) (map[int][]int, error) {
	result := make(map[int][]int)
	if len(modelIDs) == 0 {
		return result, nil
	}
	var rows []ModelVendor
	if err := DB.Where("model_id IN ?", modelIDs).Order("model_id ASC, sort_order ASC, vendor_id ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.ModelID] = append(result[row.ModelID], row.VendorID)
	}
	return result, nil
}

// MigrateLegacyModelVendors backfills the many-to-many table from the legacy
// single vendor_id column. It is safe to run repeatedly.
func MigrateLegacyModelVendors() error {
	var models []Model
	if err := DB.Select("id", "vendor_id").Where("vendor_id > ?", 0).Find(&models).Error; err != nil {
		return err
	}
	for _, model := range models {
		var count int64
		if err := DB.Model(&ModelVendor{}).Where("model_id = ? AND vendor_id = ?", model.Id, model.VendorID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			if err := DB.Create(&ModelVendor{ModelID: model.Id, VendorID: model.VendorID}).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func SetModelVendorIDs(modelID int, vendorIDs []int) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("model_id = ?", modelID).Delete(&ModelVendor{}).Error; err != nil {
			return err
		}
		seen := make(map[int]struct{}, len(vendorIDs))
		rows := make([]ModelVendor, 0, len(vendorIDs))
		for _, vendorID := range vendorIDs {
			if vendorID <= 0 {
				continue
			}
			if _, exists := seen[vendorID]; exists {
				continue
			}
			seen[vendorID] = struct{}{}
			rows = append(rows, ModelVendor{
				ModelID:   modelID,
				VendorID:  vendorID,
				SortOrder: len(rows),
			})
		}
		if len(rows) > 0 {
			return tx.Create(&rows).Error
		}
		return nil
	})
}
