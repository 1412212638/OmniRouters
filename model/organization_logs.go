package model

import (
	"errors"

	"github.com/QuantumNous/new-api/common"
)

// GetOrganizationLogs returns organization-scoped usage with the same public
// field visibility as personal usage logs. userId=0 is reserved for managers.
func GetOrganizationLogs(organizationId, userId, logType int, startTimestamp, endTimestamp int64, modelName, tokenName string, startIdx, num int) (logs []*Log, total int64, err error) {
	if organizationId <= 0 {
		return nil, 0, errors.New("organization id is required")
	}
	if startIdx < 0 {
		startIdx = 0
	}
	if num <= 0 || num > 100 {
		num = 20
	}
	tx := LOG_DB.Model(&Log{}).Where("organization_id = ?", organizationId)
	if userId > 0 {
		tx = tx.Where("user_id = ?", userId)
	}
	if logType != LogTypeUnknown {
		tx = tx.Where("type = ?", logType)
	}
	if startTimestamp > 0 {
		tx = tx.Where("created_at >= ?", startTimestamp)
	}
	if endTimestamp > 0 {
		tx = tx.Where("created_at <= ?", endTimestamp)
	}
	if tx, err = applyExplicitLogTextFilter(tx, "model_name", modelName); err != nil {
		return nil, 0, err
	}
	if tokenName != "" {
		tx = tx.Where("token_name = ?", tokenName)
	}
	if err = tx.Limit(logSearchCountLimit).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	order := "id desc"
	if common.UsingLogDatabase(common.DatabaseTypeClickHouse) {
		order = clickHouseLogOrder("")
	}
	if err = tx.Order(order).Offset(startIdx).Limit(num).Find(&logs).Error; err != nil {
		return nil, 0, err
	}
	formatUserLogs(logs, startIdx)
	return logs, total, nil
}
