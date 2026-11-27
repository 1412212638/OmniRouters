package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/bytedance/gopkg/util/gopool"
)

// notifyLimitStore is used for in-memory rate limiting when Redis is disabled
var (
	notifyLimitMu    sync.Mutex
	notifyLimitStore = make(map[string]limitCount)
	cleanupOnce      sync.Once
)

type limitCount struct {
	Count     int
	ExpiresAt time.Time
}

// startCleanupTask starts a background task to clean up expired entries
func startCleanupTask() {
	gopool.Go(func() {
		for {
			time.Sleep(time.Hour)
			notifyLimitMu.Lock()
			now := time.Now()
			for key, limit := range notifyLimitStore {
				if !now.Before(limit.ExpiresAt) {
					delete(notifyLimitStore, key)
				}
			}
			notifyLimitMu.Unlock()
		}
	})
}

// CheckNotificationLimit checks if the user has exceeded their notification limit
// Returns true if the user can send notification, false if limit exceeded
func CheckNotificationLimit(userId int, notifyType string) (bool, error) {
	count := constant.NotifyLimitCount
	minutes := constant.NotificationLimitDurationMinute
	if notifyType == dto.NotifyTypeQuotaExceed {
		count = constant.QuotaNotifyLimitCount
		minutes = constant.QuotaNotifyLimitDurationMinute
	}
	duration := time.Duration(minutes) * time.Minute
	if common.RedisEnabled {
		return checkRedisLimit(userId, notifyType, count, duration)
	}
	return checkMemoryLimit(userId, notifyType, count, duration), nil
}

func checkRedisLimit(userId int, notifyType string, limit int, duration time.Duration) (bool, error) {
	ctx := context.Background()
	key := fmt.Sprintf("notify_limit:%d:%s", userId, notifyType)

	// SETNX 只在窗口开始时写入过期时间，INCR 会保留原 TTL；放在同一事务里避免产生永不过期的计数
	pipe := common.RDB.TxPipeline()
	pipe.SetNX(ctx, key, 0, duration)
	incr := pipe.Incr(ctx, key)
	if _, err := pipe.Exec(ctx); err != nil {
		return false, fmt.Errorf("failed to increment notification count: %w", err)
	}

	count := incr.Val()
	return count <= int64(limit), nil
}

func checkMemoryLimit(userId int, notifyType string, limit int, duration time.Duration) bool {
	// Ensure cleanup task is started
	cleanupOnce.Do(startCleanupTask)

	key := fmt.Sprintf("%d:%s", userId, notifyType)
	notifyLimitMu.Lock()
	defer notifyLimitMu.Unlock()
	now := time.Now()

	currentLimit := notifyLimitStore[key]
	if !now.Before(currentLimit.ExpiresAt) {
		currentLimit = limitCount{ExpiresAt: now.Add(duration)}
	}
	currentLimit.Count++
	notifyLimitStore[key] = currentLimit
	return currentLimit.Count <= limit
}
