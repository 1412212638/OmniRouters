package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestAuthRateLimitScopesDoNotSpendEachOthersBudget(t *testing.T) {
	previousRedis, previousEnabled := common.RedisEnabled, common.CriticalRateLimitEnable
	previousCount, previousDuration := common.CriticalRateLimitNum, common.CriticalRateLimitDuration
	common.RedisEnabled, common.CriticalRateLimitEnable = false, true
	common.CriticalRateLimitNum, common.CriticalRateLimitDuration = 1, 1200
	t.Cleanup(func() {
		common.RedisEnabled, common.CriticalRateLimitEnable = previousRedis, previousEnabled
		common.CriticalRateLimitNum, common.CriticalRateLimitDuration = previousCount, previousDuration
	})
	t.Setenv("SESSION_REFRESH_RATE_LIMIT", "1")
	t.Setenv("SESSION_REFRESH_RATE_LIMIT_DURATION", "60")
	router := gin.New()
	require.NoError(t, router.SetTrustedProxies(nil))
	ok := func(c *gin.Context) { c.Status(http.StatusNoContent) }
	router.POST("/sensitive", CriticalRateLimit(), ok)
	router.POST("/refresh", SessionRefreshRateLimit(), ok)
	router.POST("/login", ScopedCriticalRateLimit("login"), ok)
	router.POST("/logout", ScopedCriticalRateLimit("logout"), ok)
	request := func(path, peer string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, path, nil)
		req.RemoteAddr = peer
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)
		return res
	}
	for _, path := range []string{"/sensitive", "/refresh", "/login", "/logout"} {
		require.Equal(t, http.StatusNoContent, request(path, "192.0.2.217:1234").Code, path)
		blocked := request(path, "192.0.2.217:1234")
		require.Equal(t, http.StatusTooManyRequests, blocked.Code, path)
		require.NotEmpty(t, blocked.Header().Get("Retry-After"))
		require.NotEmpty(t, blocked.Header().Get("X-RateLimit-Scope"))
		require.Contains(t, blocked.Body.String(), "RATE_LIMITED")
		require.Equal(t, http.StatusNoContent, request(path, "192.0.2.218:1234").Code, "another client must remain unaffected")
	}
}
