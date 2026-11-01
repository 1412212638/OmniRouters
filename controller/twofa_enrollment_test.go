package controller

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestTwoFAEnrollmentDisabledRejectsSetupAndEnable(t *testing.T) {
	common.OptionMapRWMutex.Lock()
	previous := common.OptionMap
	common.OptionMap = map[string]string{"two_fa.enrollment_enabled": "false"}
	common.OptionMapRWMutex.Unlock()
	t.Cleanup(func() {
		common.OptionMapRWMutex.Lock()
		common.OptionMap = previous
		common.OptionMapRWMutex.Unlock()
	})

	for name, handler := range map[string]gin.HandlerFunc{"setup": Setup2FA, "enable": Enable2FA} {
		t.Run(name, func(t *testing.T) {
			response := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(response)
			ctx.Request = httptest.NewRequest(http.MethodPost, "/api/user/2fa/"+name, strings.NewReader(`{"code":"123456"}`))
			ctx.Request.Header.Set("Content-Type", "application/json")
			ctx.Set("id", 1)
			handler(ctx)

			require.Equal(t, http.StatusOK, response.Code)
			var payload struct {
				Success bool   `json:"success"`
				Message string `json:"message"`
			}
			require.NoError(t, common.Unmarshal(response.Body.Bytes(), &payload))
			require.False(t, payload.Success)
			require.NotEmpty(t, payload.Message)
		})
	}
}
