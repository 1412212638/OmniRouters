package perf_metrics_setting

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestShouldExcludeStatusCode(t *testing.T) {
	orig := perfMetricsSetting
	t.Cleanup(func() { perfMetricsSetting = orig })

	perfMetricsSetting.ExcludedStatusCodes = "400,401,403,404"

	require.True(t, ShouldExcludeStatusCode(400))
	require.True(t, ShouldExcludeStatusCode(404))
	require.False(t, ShouldExcludeStatusCode(429))
	require.False(t, ShouldExcludeStatusCode(500))
}

func TestShouldExcludeStatusCodeRange(t *testing.T) {
	orig := perfMetricsSetting
	t.Cleanup(func() { perfMetricsSetting = orig })

	perfMetricsSetting.ExcludedStatusCodes = "400-404"

	require.True(t, ShouldExcludeStatusCode(402))
	require.False(t, ShouldExcludeStatusCode(405))
}

func TestShouldExcludeStatusCodeInvalidConfigFailsClosed(t *testing.T) {
	orig := perfMetricsSetting
	t.Cleanup(func() { perfMetricsSetting = orig })

	perfMetricsSetting.ExcludedStatusCodes = "oops"

	require.False(t, ShouldExcludeStatusCode(400))
}
