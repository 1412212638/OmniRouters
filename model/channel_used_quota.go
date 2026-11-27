package model

import (
	"errors"
	"fmt"
	"sync"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

type channelUsedQuotaBatchKey struct {
	id      int
	version int64
}

var channelUsedQuotaBatch = struct {
	sync.Mutex
	deltas map[channelUsedQuotaBatchKey]int
}{deltas: make(map[channelUsedQuotaBatchKey]int)}

func UpdateChannelUsedQuota(id int, quota int) {
	if quota == 0 {
		return
	}
	query := DB.Model(&Channel{}).Where("id = ?", id)
	if !common.BatchUpdateEnabled {
		writeChannelUsedQuotaDelta(query, id, quota)
		return
	}

	// Read the version from the shared database, rather than the routing cache.
	// Each delta belongs to the version current when it entered accounting.
	var channel Channel
	if err := query.Select("used_quota_version").Take(&channel).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			common.SysError(fmt.Sprintf("failed to read channel used quota version: channel_id=%d, error=%v", id, err))
		}
		return
	}
	key := channelUsedQuotaBatchKey{id: id, version: channel.UsedQuotaVersion}
	channelUsedQuotaBatch.Lock()
	addBatchDelta(channelUsedQuotaBatch.deltas, key, quota, "channel used quota")
	channelUsedQuotaBatch.Unlock()
}

func drainChannelUsedQuotaBatch() map[channelUsedQuotaBatchKey]int {
	channelUsedQuotaBatch.Lock()
	defer channelUsedQuotaBatch.Unlock()
	if len(channelUsedQuotaBatch.deltas) == 0 {
		return nil
	}
	deltas := channelUsedQuotaBatch.deltas
	channelUsedQuotaBatch.deltas = make(map[channelUsedQuotaBatchKey]int)
	return deltas
}

func flushChannelUsedQuotaBatch(deltas map[channelUsedQuotaBatchKey]int) {
	for key, quota := range deltas {
		// A reset also invalidates batches already taken by this or another node.
		query := DB.Model(&Channel{}).Where("id = ? AND used_quota_version = ?", key.id, key.version)
		writeChannelUsedQuotaDelta(query, key.id, quota)
	}
}

func writeChannelUsedQuotaDelta(query *gorm.DB, id int, quota int) {
	if err := query.UpdateColumn("used_quota", gorm.Expr("used_quota + ?", quota)).Error; err != nil {
		common.SysError(fmt.Sprintf("failed to update channel used quota: channel_id=%d, delta_quota=%d, error=%v", id, quota, err))
	}
}

func ResetChannelUsedQuota(id int) error {
	result := DB.Model(&Channel{}).Where("id = ?", id).UpdateColumns(map[string]any{
		"used_quota":         0,
		"used_quota_version": gorm.Expr("used_quota_version + 1"),
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
