package service

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 限流窗口必须由 NOTIFICATION_LIMIT_DURATION_MINUTE 决定，不能在整点被重置
func TestCheckNotificationLimitWindowSpansClockHours(t *testing.T) {
	originalRedis := common.RedisEnabled
	originalCount := constant.NotifyLimitCount
	originalMinute := constant.NotificationLimitDurationMinute
	t.Cleanup(func() {
		common.RedisEnabled = originalRedis
		constant.NotifyLimitCount = originalCount
		constant.NotificationLimitDurationMinute = originalMinute
	})
	common.RedisEnabled = false
	constant.NotifyLimitCount = 2
	constant.NotificationLimitDurationMinute = 720

	tests := []struct {
		name        string
		windowStart time.Duration
		expected    bool
	}{
		{name: "limit reached two hours ago still blocks", windowStart: 2 * time.Hour, expected: false},
		{name: "limit reached after window expired allows", windowStart: 721 * time.Minute, expected: true},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userId := 900001 + i
			key := fmt.Sprintf("%d:%s", userId, dto.NotifyTypeChannelUpdate)
			notifyLimitMu.Lock()
			notifyLimitStore[key] = limitCount{Count: 2, ExpiresAt: time.Now().Add(720*time.Minute - tt.windowStart)}
			notifyLimitMu.Unlock()
			t.Cleanup(func() {
				notifyLimitMu.Lock()
				delete(notifyLimitStore, key)
				notifyLimitMu.Unlock()
			})

			canSend, err := CheckNotificationLimit(userId, dto.NotifyTypeChannelUpdate)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, canSend)
		})
	}
}

func TestCheckNotificationLimitRedisWindowSpansClockHours(t *testing.T) {
	originalRedis := common.RedisEnabled
	originalRDB := common.RDB
	originalCount := constant.NotifyLimitCount
	originalMinute := constant.NotificationLimitDurationMinute
	server := miniredis.RunT(t)
	common.RDB = redis.NewClient(&redis.Options{Addr: server.Addr(), MaxRetries: -1})
	t.Cleanup(func() {
		_ = common.RDB.Close()
		common.RDB = originalRDB
		common.RedisEnabled = originalRedis
		constant.NotifyLimitCount = originalCount
		constant.NotificationLimitDurationMinute = originalMinute
	})
	common.RedisEnabled = true
	constant.NotifyLimitCount = 2
	constant.NotificationLimitDurationMinute = 720

	// 窗口在 2 小时前开始并已发满，剩余 10 小时
	userId := 900010
	key := fmt.Sprintf("notify_limit:%d:%s", userId, dto.NotifyTypeChannelUpdate)
	require.NoError(t, server.Set(key, "2"))
	server.SetTTL(key, 10*time.Hour)

	steps := []struct {
		name     string
		elapse   time.Duration
		expected bool
	}{
		{name: "limit reached two hours ago still blocks", expected: false},
		{name: "new window opens after the old one ends", elapse: 10*time.Hour + time.Minute, expected: true},
		{name: "second notification in the new window", expected: true},
		{name: "third notification in the new window is blocked", expected: false},
		{name: "new window ends after the configured duration", elapse: 720 * time.Minute, expected: true},
	}
	for _, step := range steps {
		server.FastForward(step.elapse)
		canSend, err := CheckNotificationLimit(userId, dto.NotifyTypeChannelUpdate)
		require.NoError(t, err, step.name)
		assert.Equal(t, step.expected, canSend, step.name)
	}
}

func TestCheckQuotaNotificationLimitAllowsOneAttemptPerDay(t *testing.T) {
	oldRedisEnabled, oldRedis := common.RedisEnabled, common.RDB
	oldCount, oldMinutes := constant.QuotaNotifyLimitCount, constant.QuotaNotifyLimitDurationMinute
	t.Cleanup(func() {
		common.RedisEnabled, common.RDB = oldRedisEnabled, oldRedis
		constant.QuotaNotifyLimitCount, constant.QuotaNotifyLimitDurationMinute = oldCount, oldMinutes
	})
	constant.QuotaNotifyLimitCount, constant.QuotaNotifyLimitDurationMinute = 1, 24*60
	const userID = 900020
	key := fmt.Sprintf("%d:%s", userID, dto.NotifyTypeQuotaExceed)

	verify := func(t *testing.T, forward func(time.Duration)) {
		t.Helper()
		const attempts = 16
		start := make(chan struct{})
		allowed := make(chan bool, attempts)
		var callers sync.WaitGroup
		for range attempts {
			callers.Add(1)
			go func() {
				defer callers.Done()
				<-start
				canSend, err := CheckNotificationLimit(userID, dto.NotifyTypeQuotaExceed)
				assert.NoError(t, err)
				allowed <- canSend
			}()
		}
		close(start)
		callers.Wait()
		close(allowed)
		sent := 0
		for canSend := range allowed {
			if canSend {
				sent++
			}
		}
		require.Equal(t, 1, sent)

		forward(24*time.Hour - time.Minute)
		canSend, err := CheckNotificationLimit(userID, dto.NotifyTypeQuotaExceed)
		require.NoError(t, err)
		require.False(t, canSend)

		forward(time.Minute)
		canSend, err = CheckNotificationLimit(userID, dto.NotifyTypeQuotaExceed)
		require.NoError(t, err)
		require.True(t, canSend)
	}

	t.Run("in memory", func(t *testing.T) {
		common.RedisEnabled = false
		t.Cleanup(func() {
			notifyLimitMu.Lock()
			delete(notifyLimitStore, key)
			notifyLimitMu.Unlock()
		})
		verify(t, func(duration time.Duration) {
			notifyLimitMu.Lock()
			defer notifyLimitMu.Unlock()
			window := notifyLimitStore[key]
			window.ExpiresAt = window.ExpiresAt.Add(-duration)
			notifyLimitStore[key] = window
		})
	})
	t.Run("Redis", func(t *testing.T) {
		server := miniredis.RunT(t)
		client := redis.NewClient(&redis.Options{Addr: server.Addr(), MaxRetries: -1})
		t.Cleanup(func() { require.NoError(t, client.Close()) })
		common.RedisEnabled, common.RDB = true, client
		verify(t, server.FastForward)
	})
}
