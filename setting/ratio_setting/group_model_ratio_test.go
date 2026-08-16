package ratio_setting

import (
	"testing"

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

func TestCheckGroupModelRatios(t *testing.T) {
	require.NoError(t, CheckGroupModelRatio(`{"default":{"model":0}}`))
	require.Error(t, CheckGroupModelRatio(`{"default":{"model":-0.1}}`))
	require.Error(t, CheckGroupModelRatio(`{"":{"model":0.5}}`))
	require.Error(t, CheckGroupModelRatio(`{"default":{"":0.5}}`))
	require.NoError(t, CheckGroupModelUserRatio(`{"default":{"model":{"42":0.3}}}`))
	require.Error(t, CheckGroupModelUserRatio(`{"default":{"model":{"0":0.3}}}`))
	require.Error(t, CheckGroupModelUserRatio(`{"":{"model":{"42":0.3}}}`))
	require.Error(t, CheckGroupModelUserRatio(`{"default":{"":{"42":0.3}}}`))
}
