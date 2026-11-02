package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func workspaceContext(path, organization string) *gin.Context {
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = httptest.NewRequest(http.MethodGet, path, nil)
	context.Request.Header.Set("New-Api-Organization", organization)
	context.Set("id", 1)
	return context
}

func TestDashboardWorkspaceHeaderIgnoresAccountAndAdminRoutes(t *testing.T) {
	for _, path := range []string{"/api/user/self", "/api/user/auth/refresh", "/api/option", "/api/log/", "/api/data/"} {
		context := workspaceContext(path, "not-a-valid-id")
		require.NoError(t, setupDashboardOrganizationContext(context), path)
		require.Zero(t, common.GetContextKeyInt(context, constant.ContextKeyOrganizationId))
	}
}

func TestDashboardWorkspaceHeaderRejectsMalformedOrDisabledOrganization(t *testing.T) {
	for _, id := range []string{"0", "-1", "nope", "1,2"} {
		require.ErrorContains(t, setupDashboardOrganizationContext(workspaceContext("/api/token/", id)), "invalid organization workspace")
	}
	for _, path := range []string{"/api/user/models", "/api/user/self/groups"} {
		require.ErrorContains(t, setupDashboardOrganizationContext(workspaceContext(path, "invalid")), "invalid organization workspace")
	}
	previous := common.OrganizationEnabled
	common.OrganizationEnabled = false
	t.Cleanup(func() { common.OrganizationEnabled = previous })
	require.ErrorContains(t, setupDashboardOrganizationContext(workspaceContext("/api/log/self", "7")), "organization feature is disabled")
	require.NoError(t, setupDashboardOrganizationContext(workspaceContext("/api/log/self", "")))
}

func TestDashboardWorkspaceHeaderRequiresActiveMembership(t *testing.T) {
	previousDB, previousEnabled := model.DB, common.OrganizationEnabled
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	model.DB, common.OrganizationEnabled = db, true
	t.Cleanup(func() { model.DB, common.OrganizationEnabled = previousDB, previousEnabled })
	require.NoError(t, db.AutoMigrate(&model.Organization{}, &model.OrganizationMember{}))
	require.NoError(t, db.Create(&model.Organization{Id: 7, Name: "Team", OwnerUserId: 1}).Error)
	require.ErrorContains(t, setupDashboardOrganizationContext(workspaceContext("/api/data/self", "7")), "membership is unavailable")
	member := &model.OrganizationMember{OrganizationId: 7, UserId: 1, Status: model.OrganizationMemberActive, Role: model.OrganizationMemberMember}
	require.NoError(t, db.Create(member).Error)
	context := workspaceContext("/api/data/self", "7")
	require.NoError(t, setupDashboardOrganizationContext(context))
	require.Equal(t, 7, common.GetContextKeyInt(context, constant.ContextKeyOrganizationId))
	require.NoError(t, db.Model(member).Update("status", model.OrganizationMemberRemoved).Error)
	require.ErrorContains(t, setupDashboardOrganizationContext(workspaceContext("/api/data/self", "7")), "membership is unavailable")
}
