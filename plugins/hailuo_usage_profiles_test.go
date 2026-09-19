package plugins_test

import (
	"maps"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/QuantumNous/new-api/pkg/jsplugin"
	builtinplugins "github.com/QuantumNous/new-api/plugins"
	taskplugin "github.com/QuantumNous/new-api/relay/channel/task/jsplugin"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Every model selects only the dimensions it can actually bill, while
// submission still reports the union so prices saved against the previous
// schema keep evaluating.
func TestHailuoUsageProfilesPerModel(t *testing.T) {
	source, err := builtinplugins.Source("hailuo")
	require.NoError(t, err)
	registry := jsplugin.NewRegistry()
	plugin, err := registry.RegisterFactory(source, jsplugin.Options{Key: "hailuo"})
	require.NoError(t, err)
	families := []struct {
		models      []string
		fields      []string
		resolutions []string
	}{
		{
			models:      []string{"MiniMax-H3"},
			fields:      []string{"input_images", "input_video_seconds", "resolution", "seconds"},
			resolutions: []string{"768P", "2K"},
		},
		{
			models:      []string{"MiniMax-Hailuo-2.3", "MiniMax-Hailuo-2.3-Fast"},
			fields:      []string{"resolution", "seconds"},
			resolutions: []string{"768P", "1080P"},
		},
		{
			models:      []string{"MiniMax-Hailuo-02"},
			fields:      []string{"resolution", "seconds"},
			resolutions: []string{"512P", "768P", "1080P"},
		},
		{
			models: []string{"T2V-01-Director", "T2V-01", "I2V-01-Director", "I2V-01-live", "I2V-01", "S2V-01"},
			fields: []string{"seconds"},
		},
	}
	profiled := make([]string, 0, len(plugin.Meta.Models))
	for _, family := range families {
		for _, name := range family.models {
			profiled = append(profiled, name)
			t.Run(name, func(t *testing.T) {
				schema, examples := plugin.Meta.UsageForModel(name)
				assert.Equal(t, family.fields, slices.Sorted(maps.Keys(schema)))
				assert.Equal(t, family.resolutions, schema["resolution"].Enum)
				assert.NotEmpty(t, examples)
			})
		}
	}
	assert.ElementsMatch(t, plugin.Meta.Models, profiled, "every declared model must select a usage profile")

	t.Run("submission keeps facts outside the selected profile billable", func(t *testing.T) {
		info := &relaycommon.RelayInfo{
			ChannelMeta:     &relaycommon.ChannelMeta{ApiKey: "test-ak", ChannelBaseUrl: "https://api.minimax.example", UpstreamModelName: "MiniMax-Hailuo-2.3"},
			OriginModelName: "MiniMax-Hailuo-2.3",
			TaskRelayInfo:   &relaycommon.TaskRelayInfo{PublicTaskID: "task_public"},
		}
		adaptor := taskplugin.New(plugin)
		adaptor.Init(info)
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodPost, "/v1/videos", nil)
		c.Set("task_request", map[string]any{"model": "MiniMax-Hailuo-2.3", "prompt": "p", "duration": 10, "resolution": "768P"})
		require.Nil(t, adaptor.ValidateRequestAndSetAction(c, info))

		facts, err := adaptor.ExtractUsageFactsValidated(c, info)
		require.NoError(t, err)
		assert.Equal(t, map[string]any{
			"seconds": float64(10), "resolution": "768P", "input_images": float64(0), "input_video_seconds": float64(0),
		}, facts)
	})
}
