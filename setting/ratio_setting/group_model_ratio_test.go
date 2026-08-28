package ratio_setting

import (
	"fmt"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/setting/config"
	"github.com/stretchr/testify/require"
)

func TestResolveGroupModelRatioPriority(t *testing.T) {
	saved := map[string]string{}
	require.NoError(t, config.GlobalConfig.SaveToDB(func(key, value string) error {
		saved[key] = value
		return nil
	}))
	t.Cleanup(func() { require.NoError(t, config.GlobalConfig.LoadFromDB(saved)) })
	require.NoError(t, config.GlobalConfig.LoadFromDB(map[string]string{
		"group_ratio_setting.group_model_ratio":      `{"default":{"gemini-test":0.5,"free-test":0}}`,
		"group_ratio_setting.group_model_user_ratio": `{"default":{"gemini-test":{"42":0.3}}}`,
	}))

	ratio, matched, userSpecific := ResolveGroupModelRatio("default", "gemini-test", 1)
	require.Equal(t, 0.5, ratio)
	require.True(t, matched)
	require.False(t, userSpecific)

	ratio, matched, userSpecific = ResolveGroupModelRatio("default", "gemini-test", 42)
	require.Equal(t, 0.3, ratio)
	require.True(t, matched)
	require.True(t, userSpecific)

	ratio, matched, _ = ResolveGroupModelRatio("default", "free-test", 1)
	require.Zero(t, ratio)
	require.True(t, matched)

	ratio, matched, _ = ResolveGroupModelRatio("default", "missing", 42)
	require.Equal(t, 1.0, ratio)
	require.False(t, matched)
}

func TestResolveGroupModelRatioExpiry(t *testing.T) {
	saved := map[string]string{}
	require.NoError(t, config.GlobalConfig.SaveToDB(func(key, value string) error {
		saved[key] = value
		return nil
	}))
	t.Cleanup(func() { require.NoError(t, config.GlobalConfig.LoadFromDB(saved)) })
	now := time.Now().Unix()
	require.NoError(t, config.GlobalConfig.LoadFromDB(map[string]string{
		"group_ratio_setting.group_model_ratio":        `{"default":{"active":0.5,"expired":0.4,"permanent":0.3}}`,
		"group_ratio_setting.group_model_user_ratio":   `{"default":{"expired":{"42":0.2}}}`,
		"group_ratio_setting.group_model_ratio_expiry": fmt.Sprintf(`{"default":{"active":%d,"expired":%d}}`, now+3600, now-1),
	}))

	ratio, matched, _ := ResolveGroupModelRatio("default", "active", 1)
	require.Equal(t, 0.5, ratio)
	require.True(t, matched)

	ratio, matched, _ = ResolveGroupModelRatio("default", "expired", 42)
	require.Equal(t, 1.0, ratio)
	require.False(t, matched)

	ratio, matched, _ = ResolveGroupModelRatio("default", "permanent", 1)
	require.Equal(t, 0.3, ratio)
	require.True(t, matched)

	visible := GetGroupModelRatioForUser(42)
	require.NotContains(t, visible["default"], "expired")
}

func TestCheckGroupModelRatios(t *testing.T) {
	require.NoError(t, CheckGroupModelRatio(`{"default":{"model":0}}`))
	require.Error(t, CheckGroupModelRatio(`{"default":{"model":-0.1}}`))
	require.Error(t, CheckGroupModelRatio(`{"":{"model":0.5}}`))
	require.Error(t, CheckGroupModelRatio(`{"default":{"":0.5}}`))
	require.NoError(t, CheckGroupModelUserRatio(`{"default":{"model":{"42":0.3}}}`))
	require.Error(t, CheckGroupModelUserRatio(`{"default":{"model":{"0":0.3}}}`))
	require.Error(t, CheckGroupModelUserRatio(`{"":{"model":{"42":0.3}}}`))
	require.Error(t, CheckGroupModelUserRatio(`{"default":{"":{"42":0.3}}}`))
	require.NoError(t, CheckGroupModelRatioExpiry(`{"default":{"model":1893456000}}`))
	require.NoError(t, CheckGroupModelRatioExpiry(`{"default":{"model":0}}`))
	require.Error(t, CheckGroupModelRatioExpiry(`{"default":{"model":-1}}`))
	require.Error(t, CheckGroupModelRatioExpiry(`{"":{"model":1}}`))
	require.Error(t, CheckGroupModelRatioExpiry(`{"default":{"":1}}`))
}
