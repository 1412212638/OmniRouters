package model

import (
	"gorm.io/gorm"
)

// IncrementOrganizationRequestCount records one completed request without
// changing the organization's monetary usage. Billing reservations and
// settlements own UsedQuota; this counter is deliberately separate so retries
// or pre-consume operations cannot inflate request counts.
func IncrementOrganizationRequestCount(organizationId int) error {
	if organizationId <= 0 {
		return nil
	}
	return DB.Model(&Organization{}).Where("id = ?", organizationId).UpdateColumn("request_count", gorm.Expr("request_count + ?", 1)).Error
}
