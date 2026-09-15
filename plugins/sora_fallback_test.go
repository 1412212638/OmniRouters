package plugins_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/pkg/jsplugin"
	builtinplugins "github.com/QuantumNous/new-api/plugins"
	taskplugin "github.com/QuantumNous/new-api/relay/channel/task/jsplugin"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSoraFallbackPreservesProviderSpecificJSON(t *testing.T) {
	source, err := builtinplugins.Source("sora")
	require.NoError(t, err)
	plugin, err := jsplugin.NewRegistry().RegisterFactory(source, jsplugin.Options{Key: "sora"})
	require.NoError(t, err)

	request := map[string]any{
		"model":         "Vidu-q3-turbo",
		"resolution":    "720p",
		"seconds":       "5",
		"prompt":        "a girl dancing with a puppy",
		"images":        []any{"https://cdn.example/first.jpg", "https://cdn.example/reference.png"},
		"aspect_ratio":  "16:9",
		"input_region":  "Mainland",
		"audio_generation": true,
		"provider_option":  map[string]any{"mode": "turbo"},
	}
	payload, err := common.Marshal(request)
	require.NoError(t, err)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/video/generations", bytes.NewReader(payload))
	c.Request.Header.Set("Content-Type", "application/json")
	defer common.CleanupBodyStorage(c)
	info := &relaycommon.RelayInfo{
		OriginModelName: "Vidu-q3-turbo",
		ChannelMeta: &relaycommon.ChannelMeta{
			ChannelType:       constant.ChannelTypeSora,
			ChannelBaseUrl:    "https://video.example",
			UpstreamModelName: "Vidu-q3-turbo",
		},
		TaskRelayInfo: &relaycommon.TaskRelayInfo{},
	}

	adaptor := taskplugin.New(plugin)
	adaptor.Init(info)
	require.Nil(t, adaptor.ValidateRequestAndSetAction(c, info))
	body, err := adaptor.BuildRequestBody(c, info)
	require.NoError(t, err)
	encoded, err := io.ReadAll(body)
	require.NoError(t, err)

	var forwarded map[string]any
	require.NoError(t, common.Unmarshal(encoded, &forwarded))
	assert.Equal(t, "Vidu-q3-turbo", forwarded["model"])
	assert.Equal(t, "720p", forwarded["resolution"])
	assert.Equal(t, "5", forwarded["seconds"])
	assert.Equal(t, "16:9", forwarded["aspect_ratio"])
	assert.Equal(t, "Mainland", forwarded["input_region"])
	assert.Equal(t, true, forwarded["audio_generation"])
	assert.Equal(t, map[string]any{"mode": "turbo"}, forwarded["provider_option"])
	assert.Equal(t, []any{"https://cdn.example/first.jpg", "https://cdn.example/reference.png"}, forwarded["images"])
}
