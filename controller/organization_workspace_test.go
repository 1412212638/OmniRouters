package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestWorkspaceModelSuggestionsUseOnlyOrganizationGroup(t *testing.T) {
	db := setupModelListControllerTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.Organization{}))
	require.NoError(t, db.Create(&model.User{Id: 1018, Username: "workspace-model-user", Group: "default", Status: common.UserStatusEnabled}).Error)
	require.NoError(t, db.Create(&model.Organization{Id: 7, Name: "Team", Group: "workspace-group", OwnerUserId: 1018}).Error)
	require.NoError(t, db.Create(&[]model.Ability{
		{Group: "default", Model: "personal-model", ChannelId: 1, Enabled: true},
		{Group: "workspace-group", Model: "organization-model", ChannelId: 2, Enabled: true},
	}).Error)

	for _, test := range []struct {
		name         string
		organization int
		query        string
		models       []string
	}{
		{name: "personal", models: []string{"personal-model"}},
		{name: "organization", organization: 7, models: []string{"organization-model"}},
		{name: "organization rejects personal group", organization: 7, query: "?group=default"},
	} {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest(http.MethodGet, "/api/user/models"+test.query, nil)
			ctx.Set("id", 1018)
			common.SetContextKey(ctx, constant.ContextKeyOrganizationId, test.organization)
			GetUserModels(ctx)
			require.ElementsMatch(t, test.models, decodeUserModelsResponse(t, recorder))
		})
	}

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/user/self/groups", nil)
	ctx.Set("id", 1018)
	common.SetContextKey(ctx, constant.ContextKeyOrganizationId, 7)
	GetUserGroups(ctx)
	var groups struct {
		Success bool                              `json:"success"`
		Data    map[string]map[string]interface{} `json:"data"`
	}
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &groups))
	require.True(t, groups.Success)
	require.Len(t, groups.Data, 1)
	require.Contains(t, groups.Data, "workspace-group")
}
