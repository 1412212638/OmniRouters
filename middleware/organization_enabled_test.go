package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestOrganizationEnabledRejectsDisabledWorkspace(t *testing.T) {
	previous := common.OrganizationEnabled
	common.OrganizationEnabled = false
	t.Cleanup(func() { common.OrganizationEnabled = previous })

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/organizations", OrganizationEnabled(), func(c *gin.Context) {
		t.Fatal("disabled organization route reached its handler")
	})
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/organizations", nil))
	require.Equal(t, http.StatusForbidden, response.Code)
}

func TestSetupContextForOrganizationTokenRejectsDisabledWorkspace(t *testing.T) {
	previous := common.OrganizationEnabled
	common.OrganizationEnabled = false
	t.Cleanup(func() { common.OrganizationEnabled = previous })

	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	err := SetupContextForToken(context, &model.Token{UserId: 1, OrganizationId: 1})
	require.ErrorContains(t, err, "organization workspaces are disabled")
}

func TestSetupContextForPersonalTokenRemainsAvailableWhenOrganizationsDisabled(t *testing.T) {
	previous := common.OrganizationEnabled
	common.OrganizationEnabled = false
	t.Cleanup(func() { common.OrganizationEnabled = previous })

	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	require.NoError(t, SetupContextForToken(context, &model.Token{UserId: 1, UnlimitedQuota: true}))
}
