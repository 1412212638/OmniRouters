package limiter

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"
)

func TestAllowFallsBackToEvalAfterScriptFlush(t *testing.T) {
	ctx := context.Background()
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { require.NoError(t, client.Close()) })

	sha, err := client.ScriptLoad(ctx, rateLimitScript).Result()
	require.NoError(t, err)
	limiter := &RedisLimiter{client: client, limitScriptSHA: sha}

	require.NoError(t, client.ScriptFlush(ctx).Err())
	allowed, err := limiter.Allow(ctx, "rate-limit:after-flush", WithCapacity(60), WithRate(60), WithRequested(1))

	require.NoError(t, err)
	require.True(t, allowed)
	exists, err := client.ScriptExists(ctx, sha).Result()
	require.NoError(t, err)
	require.Equal(t, []bool{true}, exists)
}
