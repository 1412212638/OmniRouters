package middleware

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/common/limiter"
	"github.com/QuantumNous/new-api/constant"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/setting"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

const (
	ModelRequestRateLimitCountMark        = "MRRL"
	ModelRequestRateLimitSuccessCountMark = "MRRLS"
)

var modelRateLimitCounters = struct {
	sync.Mutex
	values map[string]int64
}{values: make(map[string]int64)}

func newRateLimitConfigured(group string) bool {
	if _, found := setting.GetModelRateLimit(group); found {
		return true
	}
	global := setting.EffectiveGlobalRateLimit()
	if global.RPM > 0 || global.TPM > 0 || global.RPD > 0 {
		return true
	}
	// A saved v2 configuration with all three values set to zero is an
	// explicit unlimited configuration. Fresh legacy installations still use
	// the old success-count default until the new settings are saved.
	return setting.ModelRequestRateLimitCount == 0 && setting.ModelRequestRateLimitSuccessCount == 0
}

func checkFixedWindowMemory(key string, limit int, window time.Duration, amount int64) bool {
	if limit <= 0 {
		return true
	}
	bucket := time.Now().Unix() / int64(window.Seconds())
	key += ":" + strconv.FormatInt(bucket, 10)
	modelRateLimitCounters.Lock()
	defer modelRateLimitCounters.Unlock()
	current := modelRateLimitCounters.values[key]
	if amount == 0 {
		return current < int64(limit)
	}
	if current+amount > int64(limit) {
		return false
	}
	modelRateLimitCounters.values[key] = current + amount
	return true
}

func checkFixedWindowRedis(ctx context.Context, rdb *redis.Client, key string, limit int, window time.Duration, amount int64) (bool, error) {
	if limit <= 0 {
		return true, nil
	}
	bucket := time.Now().Unix() / int64(window.Seconds())
	key += ":" + strconv.FormatInt(bucket, 10)
	if amount == 0 {
		value, err := rdb.Get(ctx, key).Int64()
		if err == redis.Nil {
			return true, nil
		}
		return value < int64(limit), err
	}
	value, err := rdb.IncrBy(ctx, key, amount).Result()
	if err != nil {
		return false, err
	}
	if value == amount {
		if err := rdb.Expire(ctx, key, window+time.Minute).Err(); err != nil {
			return false, err
		}
	}
	if value <= int64(limit) {
		return true, nil
	}
	_ = rdb.DecrBy(ctx, key, amount).Err()
	return false, nil
}

func recordFixedWindowMemory(key string, window time.Duration, amount int64) {
	if amount <= 0 {
		return
	}
	bucket := time.Now().Unix() / int64(window.Seconds())
	key += ":" + strconv.FormatInt(bucket, 10)
	modelRateLimitCounters.Lock()
	modelRateLimitCounters.values[key] += amount
	modelRateLimitCounters.Unlock()
}

func recordFixedWindowRedis(ctx context.Context, rdb *redis.Client, key string, window time.Duration, amount int64) error {
	if amount <= 0 {
		return nil
	}
	bucket := time.Now().Unix() / int64(window.Seconds())
	key += ":" + strconv.FormatInt(bucket, 10)
	value, err := rdb.IncrBy(ctx, key, amount).Result()
	if err != nil {
		return err
	}
	if value == amount {
		return rdb.Expire(ctx, key, window+time.Minute).Err()
	}
	return nil
}

func newRateLimitHandler(c *gin.Context, subject string, limits setting.ModelRateLimit) {
	ctx := context.Background()
	checks := []struct {
		name   string
		limit  int
		window time.Duration
	}{
		{"rpm", limits.RPM, time.Minute},
		{"tpm", limits.TPM, time.Minute},
		{"rpd", limits.RPD, 24 * time.Hour},
	}
	for _, check := range checks {
		key := "rateLimit:v2:" + check.name + ":" + subject
		amount := int64(1)
		if check.name == "tpm" {
			// Token usage is known after the upstream response. Check the current
			// window before dispatch, then record actual usage below.
			amount = 0
		}
		allowed := true
		var err error
		if common.RedisEnabled {
			allowed, err = checkFixedWindowRedis(ctx, common.RDB, key, check.limit, check.window, check.amount)
		} else {
			allowed = checkFixedWindowMemory(key, check.limit, check.window, check.amount)
		}
		if err != nil {
			abortWithOpenAiMessage(c, http.StatusInternalServerError, "rate_limit_check_failed")
			return
		}
		if !allowed {
			abortWithOpenAiMessage(c, http.StatusTooManyRequests, fmt.Sprintf("您已达到%s限制", strings.ToUpper(check.name)))
			return
		}
	}
	c.Next()
	if limits.TPM > 0 {
		tokens := int64(common.GetContextKeyInt(c, constant.ContextKeyRateLimitTokens))
		if tokens <= 0 {
			tokens = int64(common.GetContextKeyInt(c, constant.ContextKeyEstimatedTokens))
		}
		if tokens > int64(limits.TPM) {
			tokens = int64(limits.TPM)
		}
		if tokens > 0 {
			key := "rateLimit:v2:tpm:" + subject
			if common.RedisEnabled {
				_ = recordFixedWindowRedis(ctx, common.RDB, key, time.Minute, tokens)
			} else {
				recordFixedWindowMemory(key, time.Minute, tokens)
			}
		}
	}
}

// 检查Redis中的请求限制
func checkRedisRateLimit(ctx context.Context, rdb *redis.Client, key string, maxCount int, duration int64) (bool, error) {
	// 如果maxCount为0，表示不限制
	if maxCount == 0 {
		return true, nil
	}

	// 获取当前计数
	length, err := rdb.LLen(ctx, key).Result()
	if err != nil {
		return false, err
	}

	// 如果未达到限制，允许请求
	if length < int64(maxCount) {
		return true, nil
	}

	// 检查时间窗口
	oldTimeStr, _ := rdb.LIndex(ctx, key, -1).Result()
	oldTime, err := time.Parse(timeFormat, oldTimeStr)
	if err != nil {
		return false, err
	}

	nowTimeStr := time.Now().Format(timeFormat)
	nowTime, err := time.Parse(timeFormat, nowTimeStr)
	if err != nil {
		return false, err
	}
	// 如果在时间窗口内已达到限制，拒绝请求
	subTime := nowTime.Sub(oldTime).Seconds()
	if int64(subTime) < duration {
		rdb.Expire(ctx, key, time.Duration(setting.ModelRequestRateLimitDurationMinutes)*time.Minute)
		return false, nil
	}

	return true, nil
}

// 记录Redis请求
func recordRedisRequest(ctx context.Context, rdb *redis.Client, key string, maxCount int) {
	// 如果maxCount为0，不记录请求
	if maxCount == 0 {
		return
	}

	now := time.Now().Format(timeFormat)
	rdb.LPush(ctx, key, now)
	rdb.LTrim(ctx, key, 0, int64(maxCount-1))
	rdb.Expire(ctx, key, time.Duration(setting.ModelRequestRateLimitDurationMinutes)*time.Minute)
}

// Redis限流处理器
func redisRateLimitHandler(subject string, duration int64, totalMaxCount, successMaxCount int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		rdb := common.RDB

		// 1. 检查成功请求数限制
		successKey := fmt.Sprintf("rateLimit:%s:%s", ModelRequestRateLimitSuccessCountMark, subject)
		allowed, err := checkRedisRateLimit(ctx, rdb, successKey, successMaxCount, duration)
		if err != nil {
			fmt.Println("检查成功请求数限制失败:", err.Error())
			abortWithOpenAiMessage(c, http.StatusInternalServerError, "rate_limit_check_failed")
			return
		}
		if !allowed {
			abortWithOpenAiMessage(c, http.StatusTooManyRequests, fmt.Sprintf("您已达到请求数限制：%d分钟内最多请求%d次", setting.ModelRequestRateLimitDurationMinutes, successMaxCount))
			return
		}

		//2.检查总请求数限制并记录总请求（当totalMaxCount为0时会自动跳过，使用令牌桶限流器
		if totalMaxCount > 0 {
			totalKey := fmt.Sprintf("rateLimit:%s", subject)
			// 初始化
			tb := limiter.New(ctx, rdb)
			allowed, err = tb.Allow(
				ctx,
				totalKey,
				limiter.WithCapacity(rateLimitCapacity(totalMaxCount, duration)),
				limiter.WithRate(int64(totalMaxCount)),
				limiter.WithRequested(duration),
			)

			if err != nil {
				fmt.Println("检查总请求数限制失败:", err.Error())
				abortWithOpenAiMessage(c, http.StatusInternalServerError, "rate_limit_check_failed")
				return
			}

			if !allowed {
				abortWithOpenAiMessage(c, http.StatusTooManyRequests, fmt.Sprintf("您已达到总请求数限制：%d分钟内最多请求%d次，包括失败次数，请检查您的请求是否正确", setting.ModelRequestRateLimitDurationMinutes, totalMaxCount))
				return
			}
		}

		// 4. 处理请求
		c.Next()

		// 5. 如果请求成功，记录成功请求
		if modelRequestSucceeded(c) {
			recordRedisRequest(ctx, rdb, successKey, successMaxCount)
		}
	}
}

// 内存限流处理器
func memoryRateLimitHandler(subject string, duration int64, totalMaxCount, successMaxCount int) gin.HandlerFunc {
	inMemoryRateLimiter.Init(time.Duration(setting.ModelRequestRateLimitDurationMinutes) * time.Minute)

	return func(c *gin.Context) {
		totalKey := ModelRequestRateLimitCountMark + subject
		successKey := ModelRequestRateLimitSuccessCountMark + subject

		// 1. 检查总请求数限制（当totalMaxCount为0时跳过）
		if totalMaxCount > 0 && !inMemoryRateLimiter.Request(totalKey, totalMaxCount, duration) {
			c.Status(http.StatusTooManyRequests)
			c.Abort()
			return
		}

		var reservation *common.RateLimitReservation
		if successMaxCount > 0 {
			reservation = inMemoryRateLimiter.Reserve(successKey, successMaxCount, duration)
			if reservation == nil {
				c.AbortWithStatus(http.StatusTooManyRequests)
				return
			}
			defer reservation.Complete(false)
		}

		// 3. 处理请求
		c.Next()

		// 4. 如果请求成功，记录到实际的成功请求计数中
		reservation.Complete(modelRequestSucceeded(c))
	}
}

func modelRequestSucceeded(c *gin.Context) bool {
	status, _ := common.GetContextKeyType[*relaycommon.StreamStatus](c, constant.ContextKeyResponseStreamStatus)
	return c.Writer.Status() < 400 &&
		(c.Request == nil || c.Request.Context().Err() == nil) && status.RequestSucceeded()
}

// ModelRequestRateLimit 模型请求限流中间件
func ModelRequestRateLimit() func(c *gin.Context) {
	return func(c *gin.Context) {
		// 在每个请求时检查是否启用限流
		if !setting.ModelRequestRateLimitEnabled {
			c.Next()
			return
		}

		// 计算限流参数
		duration := rateLimitDurationSeconds(setting.ModelRequestRateLimitDurationMinutes)
		totalMaxCount := setting.ModelRequestRateLimitCount
		successMaxCount := setting.ModelRequestRateLimitSuccessCount

		// 获取分组
		group := common.GetContextKeyString(c, constant.ContextKeyTokenGroup)
		if group == "" {
			group = common.GetContextKeyString(c, constant.ContextKeyUserGroup)
		}

		//获取分组的限流配置
		groupTotalCount, groupSuccessCount, found := setting.GetGroupRateLimit(group)
		subject := strconv.Itoa(c.GetInt("id"))
		if found {
			totalMaxCount = groupTotalCount
			successMaxCount = groupSuccessCount
			subject += ":" + group
		}

		if newRateLimitConfigured(group) {
			limits := setting.EffectiveGlobalRateLimit()
			if groupLimits, ok := setting.GetModelRateLimit(group); ok {
				limits = groupLimits
			}
			newRateLimitHandler(c, subject, limits)
			return
		}

		// 根据存储类型选择并执行限流处理器
		if common.RedisEnabled {
			redisRateLimitHandler(subject, duration, totalMaxCount, successMaxCount)(c)
		} else {
			memoryRateLimitHandler(subject, duration, totalMaxCount, successMaxCount)(c)
		}
	}
}

func rateLimitDurationSeconds(durationMinutes int) int64 {
	if durationMinutes <= 0 {
		return 0
	}
	minutes := int64(durationMinutes)
	if minutes > math.MaxInt64/60 {
		return math.MaxInt64
	}
	return minutes * 60
}

func rateLimitCapacity(count int, durationSeconds int64) int64 {
	if count <= 0 || durationSeconds <= 0 {
		return 0
	}
	c := int64(count)
	if c > math.MaxInt64/durationSeconds {
		return math.MaxInt64
	}
	return c * durationSeconds
}
