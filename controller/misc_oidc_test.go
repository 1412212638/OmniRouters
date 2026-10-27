package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetStatusReturnsEffectiveOIDCDisplayName(t *testing.T) {
	settings := system_setting.GetOIDCSettings()
	originalDisplayName := settings.DisplayName
	originalOptionMap := common.OptionMap
	t.Cleanup(func() {
		settings.DisplayName = originalDisplayName
		common.OptionMap = originalOptionMap
	})
	common.OptionMap = map[string]string{}

	tests := []struct {
		name        string
		displayName string
		want        string
	}{
		{
			name:        "custom name is trimmed",
			displayName: "  Acme SSO  ",
			want:        "Acme SSO",
		},
		{
			name:        "whitespace-only name falls back",
			displayName: "   ",
			want:        "OIDC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings.DisplayName = tt.displayName
			response := httptest.NewRecorder()
			context, _ := gin.CreateTestContext(response)
			context.Request = httptest.NewRequest(http.MethodGet, "/api/status", nil)

			GetStatus(context)

			var payload struct {
				Success bool           `json:"success"`
				Data    map[string]any `json:"data"`
			}
			require.NoError(t, common.Unmarshal(response.Body.Bytes(), &payload))
			require.True(t, payload.Success)
			assert.Equal(t, tt.want, payload.Data["oidc_display_name"])
		})
	}
}

func TestGetStatusOmitsSensitiveSettingsForCommonUser(t *testing.T) {
	common.OptionMapRWMutex.Lock()
	previous := common.OptionMap
	common.OptionMap = map[string]string{
		"SidebarModulesAdmin": "{\"admin\":{\"enabled\":true}}",
		"HeaderNavModules":    "{\"pricing\":{\"enabled\":true}}",
	}
	common.OptionMapRWMutex.Unlock()
	t.Cleanup(func() {
		common.OptionMapRWMutex.Lock()
		common.OptionMap = previous
		common.OptionMapRWMutex.Unlock()
	})

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(sessions.Sessions("session", cookie.NewStore([]byte("status-test"))))
	router.GET("/login", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Set("id", 1)
		session.Set("role", common.RoleCommonUser)
		session.Set("status", common.UserStatusEnabled)
		require.NoError(t, session.Save())
		c.Status(http.StatusNoContent)
	})
	router.GET("/api/status", middleware.TryUserAuth(), GetStatus)

	loginResponse := httptest.NewRecorder()
	router.ServeHTTP(loginResponse, httptest.NewRequest(http.MethodGet, "/login", nil))
	require.Equal(t, http.StatusNoContent, loginResponse.Code)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	for _, cookie := range loginResponse.Result().Cookies() {
		request.AddCookie(cookie)
	}
	router.ServeHTTP(response, request)

	var payload struct {
		Success bool           `json:"success"`
		Data    map[string]any `json:"data"`
	}
	require.NoError(t, common.Unmarshal(response.Body.Bytes(), &payload))
	require.True(t, payload.Success)
	assert.NotContains(t, payload.Data, "SidebarModulesAdmin")
	assert.NotContains(t, payload.Data, "version")
	assert.Equal(t, "{\"pricing\":{\"enabled\":true}}", payload.Data["HeaderNavModules"])
}

func TestGetStatusOmitsSensitiveAnonymousSettings(t *testing.T) {
	common.OptionMapRWMutex.Lock()
	previous := common.OptionMap
	common.OptionMap = map[string]string{
		"SidebarModulesAdmin": "{\"admin\":{\"enabled\":true}}",
		"HeaderNavModules":    "{\"pricing\":{\"enabled\":true}}",
	}
	common.OptionMapRWMutex.Unlock()
	t.Cleanup(func() {
		common.OptionMapRWMutex.Lock()
		common.OptionMap = previous
		common.OptionMapRWMutex.Unlock()
	})

	response := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(response)
	context.Request = httptest.NewRequest(http.MethodGet, "/api/status", nil)

	GetStatus(context)

	var payload struct {
		Success bool           `json:"success"`
		Data    map[string]any `json:"data"`
	}
	require.NoError(t, common.Unmarshal(response.Body.Bytes(), &payload))
	require.True(t, payload.Success)
	assert.NotContains(t, payload.Data, "SidebarModulesAdmin")
	assert.NotContains(t, payload.Data, "version")
	assert.NotContains(t, payload.Data, "start_time")
	assert.NotContains(t, payload.Data, "server_address")
	assert.NotContains(t, payload.Data, "stripe_unit_price")
	assert.NotContains(t, payload.Data, "passkey_origins")
	assert.Equal(t, "{\"pricing\":{\"enabled\":true}}", payload.Data["HeaderNavModules"])
}
