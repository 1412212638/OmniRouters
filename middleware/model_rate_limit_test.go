package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"
)

var modelRateLimitTestUsers atomic.Int64

func modelLimitRequest(router http.Handler, path string) int {
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
	return response.Code
}

func modelLimitBackend(t *testing.T, backend string, total, success int) (int, gin.HandlerFunc) {
	t.Helper()
	user := 8200000 + int(modelRateLimitTestUsers.Add(1))
	if backend == "memory" {
		return user, memoryRateLimitHandler(60, total, success)
	}
	addr := os.Getenv("TEST_REDIS_ADDR")
	if addr == "" {
		t.Skip("TEST_REDIS_ADDR is required for Redis integration tests")
	}
	client := redis.NewClient(&redis.Options{Addr: addr, DB: 15})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	require.NoError(t, client.Ping(ctx).Err())
	previous := common.RDB
	common.RDB = client
	t.Cleanup(func() { common.RDB = previous; _ = client.Close() })
	key := "rateLimit:" + ModelRequestRateLimitSuccessCountMark + ":" + strconv.Itoa(user)
	totalKey := "rateLimit:" + strconv.Itoa(user)
	require.NoError(t, client.Del(ctx, key, totalKey).Err())
	t.Cleanup(func() { _ = client.Del(context.Background(), key, totalKey).Err() })
	return user, redisRateLimitHandler(60, total, success)
}

func TestModelRateLimitProtocolFailureReleasesSuccessLimit(t *testing.T) {
	for _, backend := range []string{"memory", "redis"} {
		t.Run(backend, func(t *testing.T) {
			user, limit := modelLimitBackend(t, backend, 0, 1)
			router := gin.New()
			router.GET("/:outcome", func(c *gin.Context) { c.Set("id", user) }, limit, func(c *gin.Context) {
				status := relaycommon.NewStreamStatus()
				switch c.Param("outcome") {
				case "failed":
					status.MarkFailed("upstream", "server_error", 0)
				case "cancelled":
					status.MarkCancelled()
				case "cut":
					status.RequireTerminal()
					status.SetEndReason(relaycommon.StreamEndReasonEOF, nil)
				default:
					status.MarkCompleted()
				}
				common.SetContextKey(c, constant.ContextKeyResponseStreamStatus, status)
				c.Status(http.StatusOK)
			})
			for _, path := range []string{"/failed", "/cancelled", "/cut", "/completed"} {
				require.Equal(t, http.StatusOK, modelLimitRequest(router, path), path)
			}
			require.Equal(t, http.StatusTooManyRequests, modelLimitRequest(router, "/completed"))
		})
	}
}

func TestModelMemoryRateLimitConcurrentReservationAndFailureRelease(t *testing.T) {
	user, limit := modelLimitBackend(t, "memory", 0, 1)
	entered, release := make(chan struct{}), make(chan struct{})
	finished := make(chan int, 1)
	router := gin.New()
	router.GET("/:outcome", func(c *gin.Context) { c.Set("id", user) }, limit, func(c *gin.Context) {
		if c.Param("outcome") == "slow" {
			close(entered)
			<-release
			c.Status(http.StatusBadGateway)
		}
	})
	go func() { finished <- modelLimitRequest(router, "/slow") }()
	select {
	case <-entered:
	case <-time.After(5 * time.Second):
		t.Fatal("request did not enter handler")
	}
	blocked := modelLimitRequest(router, "/completed")
	close(release)
	require.Equal(t, http.StatusTooManyRequests, blocked)
	require.Equal(t, http.StatusBadGateway, <-finished)
	require.Equal(t, http.StatusOK, modelLimitRequest(router, "/completed"))
	require.Equal(t, http.StatusTooManyRequests, modelLimitRequest(router, "/completed"))
}

func TestModelMemoryRateLimitPanicCleanupAndDisabledSuccessLimit(t *testing.T) {
	for _, success := range []int{0, 1} {
		user, limit := modelLimitBackend(t, "memory", 0, success)
		router := gin.New()
		router.Use(gin.Recovery())
		router.GET("/:outcome", func(c *gin.Context) { c.Set("id", user) }, limit, func(c *gin.Context) {
			if c.Param("outcome") == "panic" {
				panic("test")
			}
		})
		require.Equal(t, http.StatusInternalServerError, modelLimitRequest(router, "/panic"))
		require.Equal(t, http.StatusOK, modelLimitRequest(router, "/completed"))
		if success == 0 {
			require.Equal(t, http.StatusOK, modelLimitRequest(router, "/completed"))
		}
	}
}

func TestModelMemoryRateLimitFailuresStillCountTowardTotal(t *testing.T) {
	user, limit := modelLimitBackend(t, "memory", 2, 1)
	router := gin.New()
	router.GET("/", func(c *gin.Context) { c.Set("id", user) }, limit, func(c *gin.Context) { c.Status(http.StatusBadGateway) })
	require.Equal(t, http.StatusBadGateway, modelLimitRequest(router, "/"))
	require.Equal(t, http.StatusBadGateway, modelLimitRequest(router, "/"))
	require.Equal(t, http.StatusTooManyRequests, modelLimitRequest(router, "/"))
}

func TestModelRedisRateLimitFailuresStillCountTowardTotal(t *testing.T) {
	user, limit := modelLimitBackend(t, "redis", 1, 0)
	var calls int
	router := gin.New()
	router.GET("/", func(c *gin.Context) { c.Set("id", user) }, limit, func(c *gin.Context) {
		calls++
		c.Status(http.StatusBadGateway)
	})
	require.Equal(t, http.StatusBadGateway, modelLimitRequest(router, "/"))
	require.Equal(t, http.StatusTooManyRequests, modelLimitRequest(router, "/"))
	require.Equal(t, 1, calls)
}

func TestModelRateLimitCancelledRequestDoesNotCountAsSuccess(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx, cancel := context.WithCancel(context.Background())
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
	status := relaycommon.NewStreamStatus()
	status.MarkCompleted()
	common.SetContextKey(c, constant.ContextKeyResponseStreamStatus, status)
	cancel()
	require.False(t, modelRequestSucceeded(c))
}
