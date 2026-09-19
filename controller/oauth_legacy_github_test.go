package controller

import (
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/oauth"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestGitHubLegacyLoginRequiresIndependentAuthentication(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	previousDB, previousRegistration := model.DB, common.RegisterEnabled
	model.DB, common.RegisterEnabled = db, false
	t.Cleanup(func() {
		model.DB, common.RegisterEnabled = previousDB, previousRegistration
		raw, _ := db.DB()
		_ = raw.Close()
	})
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.AuditLog{}))
	legacy := model.User{Username: "legacy-owner", GitHubId: "reused-login"}
	permanent := model.User{Username: "numeric-owner", GitHubId: "12345"}
	require.NoError(t, db.Create(&legacy).Error)
	require.NoError(t, db.Create(&permanent).Error)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/api/oauth/github", nil)
	provider := &oauth.GitHubProvider{}

	user, err := findOrCreateOAuthUser(c, provider, &oauth.OAuthUser{
		ProviderUserID: "99999", Extra: map[string]any{"legacy_id": "reused-login"},
	}, nil)
	require.Nil(t, user)
	var rejected *OAuthLegacyBindingNotConfirmedError
	require.ErrorAs(t, err, &rejected)
	var stored model.User
	require.NoError(t, db.First(&stored, legacy.Id).Error)
	require.Equal(t, "reused-login", stored.GitHubId, "username match must not rewrite ownership")
	var audit model.AuditLog
	require.NoError(t, db.Where("user_id = ?", legacy.Id).First(&audit).Error)
	require.Equal(t, "github_legacy_binding_declined", audit.Action)
	require.False(t, audit.Success)

	user, err = findOrCreateOAuthUser(c, provider, &oauth.OAuthUser{
		ProviderUserID: "99999", Extra: map[string]any{"legacy_id": "12345"},
	}, nil)
	require.Nil(t, user)
	var disabled *OAuthRegistrationDisabledError
	require.ErrorAs(t, err, &disabled, "numeric login must not select a different numeric account ID")

	user, err = findOrCreateOAuthUser(c, provider, &oauth.OAuthUser{
		ProviderUserID: "12345", Extra: map[string]any{"legacy_id": "reused-login"},
	}, nil)
	require.NoError(t, err)
	require.Equal(t, permanent.Id, user.Id, "permanent ID must take precedence")
}
