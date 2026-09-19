package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/pkg/jsplugin"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestTaskPluginNoChannelMessageDoesNotExposePluginKeys(t *testing.T) {
	require.NoError(t, i18n.Init())
	for _, language := range []string{"en", "zh-CN", "zh-TW"} {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("POST", "/v1/responses", nil)
		c.Request.Header.Set("Accept-Language", language)
		c.Set(jsplugin.ContextKeyPinnedPlugin, jsplugin.PinnedPlugin{
			Plugin: &jsplugin.LoadedPlugin{Meta: jsplugin.Meta{Key: "private-plugin-key"}},
		})
		message := noAvailableChannelMessage(c, "test-group", "test-model")
		require.NotContains(t, message, "private-plugin-key")
		require.NotContains(t, message, "disable or override")
		require.Contains(t, message, "test-group")
		require.Contains(t, message, "test-model")
	}
}
