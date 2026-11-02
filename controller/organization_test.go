package controller

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func setupOrganizationControllerTest(t *testing.T) {
	t.Helper()
	db := setupTokenControllerTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.Organization{}, &model.OrganizationMember{}, &model.OrganizationMemberUsage{}, &model.OrganizationInvitation{}, &model.OrganizationWalletTransaction{}))
	require.NoError(t, db.Create(&model.User{Id: 1, Username: "owner", DisplayName: "Owner"}).Error)
	require.NoError(t, db.Create(&model.User{Id: 2, Username: "member", DisplayName: "Member"}).Error)
	require.NoError(t, db.Create(&model.User{Id: 3, Username: "admin", DisplayName: "Administrator"}).Error)
	require.NoError(t, db.Create(&model.Organization{Id: 1, Name: "Team", OwnerUserId: 1, Quota: 1000}).Error)
	require.NoError(t, db.Create(&model.OrganizationMember{OrganizationId: 1, UserId: 1, Role: model.OrganizationMemberOwner, Status: model.OrganizationMemberActive}).Error)
	require.NoError(t, db.Create(&model.OrganizationMember{OrganizationId: 1, UserId: 2, Role: model.OrganizationMemberMember, Status: model.OrganizationMemberActive}).Error)
	require.NoError(t, db.Create(&model.OrganizationMember{OrganizationId: 1, UserId: 3, Role: model.OrganizationMemberAdmin, Status: model.OrganizationMemberActive}).Error)
}

func organizationTestContext(userID int, body string) (*gin.Context, *httptest.ResponseRecorder) {
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set("id", userID)
	ctx.Params = gin.Params{{Key: "organization_id", Value: "1"}}
	ctx.Request = httptest.NewRequest(http.MethodPatch, "/api/organizations/1", strings.NewReader(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	return ctx, recorder
}

func TestOrganizationMemberListAndWalletVisibility(t *testing.T) {
	setupOrganizationControllerTest(t)
	ctx, recorder := organizationTestContext(2, "")
	GetOrganizationMembers(ctx)
	var response struct {
		Success bool `json:"success"`
	}
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &response))
	require.False(t, response.Success)
	ctx, recorder = organizationTestContext(2, "")
	GetOrganization(ctx)
	var overview struct {
		Success bool                           `json:"success"`
		Data    organizationMembershipResponse `json:"data"`
	}
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &overview))
	require.True(t, overview.Success)
	require.Zero(t, overview.Data.Organization.Quota)
}

func TestOrganizationLegacyBillingUsageRequiresManager(t *testing.T) {
	setupOrganizationControllerTest(t)
	for _, userID := range []int{1, 2, 3} {
		ctx, recorder := organizationTestContext(userID, "")
		common.SetContextKey(ctx, constant.ContextKeyOrganizationId, 1)
		GetUsage(ctx)
		if userID == 2 {
			require.Equal(t, http.StatusForbidden, recorder.Code, "ordinary members must not see the shared wallet's lifetime usage through the legacy billing endpoint")
		} else {
			require.Equal(t, http.StatusOK, recorder.Code)
		}
	}
}

func TestOrganizationMembersOnlySeeOwnMaskedTokens(t *testing.T) {
	setupOrganizationControllerTest(t)
	require.NoError(t, model.DB.Create(&model.Token{UserId: 2, OrganizationId: 1, Name: "own", Key: "organization-own-secret", UnlimitedQuota: true}).Error)
	require.NoError(t, model.DB.Create(&model.Token{UserId: 3, OrganizationId: 1, Name: "other", Key: "organization-other-secret", UnlimitedQuota: true}).Error)
	ctx, recorder := organizationTestContext(2, "")
	GetOrganizationTokens(ctx)
	var response struct {
		Success bool          `json:"success"`
		Data    []model.Token `json:"data"`
	}
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &response))
	require.True(t, response.Success)
	require.Len(t, response.Data, 1)
	require.Equal(t, 2, response.Data[0].UserId)
	require.NotEqual(t, "organization-own-secret", response.Data[0].Key)
	var metadata struct {
		Data []struct {
			CreatorName string `json:"creator_name"`
		} `json:"data"`
	}
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &metadata))
	require.Equal(t, "Member", metadata.Data[0].CreatorName)
}

func TestOrganizationMemberCannotRevealAnotherMembersKey(t *testing.T) {
	setupOrganizationControllerTest(t)
	token := &model.Token{UserId: 3, OrganizationId: 1, Name: "other", Key: "other-member-secret", Status: common.TokenStatusEnabled, UnlimitedQuota: true}
	require.NoError(t, model.DB.Create(token).Error)
	ctx, recorder := organizationTestContext(2, "")
	ctx.Params = append(ctx.Params, gin.Param{Key: "token_id", Value: strconv.Itoa(token.Id)})
	GetOrganizationTokenKey(ctx)
	var response struct {
		Success bool `json:"success"`
	}
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &response))
	require.False(t, response.Success)
	require.NotContains(t, recorder.Body.String(), token.Key)
}

func TestOrganizationManagerWalletLedgerIncludesActor(t *testing.T) {
	setupOrganizationControllerTest(t)
	require.NoError(t, model.DB.Create(&model.OrganizationWalletTransaction{OrganizationId: 1, UserId: 1, Amount: 500, BalanceAfter: 1500, Type: "transfer_in", Description: "Personal wallet transfer"}).Error)
	ctx, recorder := organizationTestContext(2, "")
	GetOrganizationWalletTransactions(ctx)
	var denied struct {
		Success bool `json:"success"`
	}
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &denied))
	require.False(t, denied.Success)
	ctx, recorder = organizationTestContext(3, "")
	GetOrganizationWalletTransactions(ctx)
	var response struct {
		Success bool `json:"success"`
		Data    []struct {
			UserName string `json:"user_name"`
			Amount   int    `json:"amount"`
			Type     string `json:"type"`
		} `json:"data"`
	}
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &response))
	require.True(t, response.Success)
	require.Len(t, response.Data, 1)
	require.Equal(t, "Owner", response.Data[0].UserName)
	require.Equal(t, 500, response.Data[0].Amount)
	require.Equal(t, "transfer_in", response.Data[0].Type)
}

func TestOrganizationAdministratorCannotChangeAnotherAdministrator(t *testing.T) {
	setupOrganizationControllerTest(t)
	require.NoError(t, model.DB.Create(&model.OrganizationMember{OrganizationId: 1, UserId: 4, Role: model.OrganizationMemberAdmin, Status: model.OrganizationMemberActive}).Error)
	ctx, recorder := organizationTestContext(3, `{"role":"member"}`)
	ctx.Params = append(ctx.Params, gin.Param{Key: "user_id", Value: "4"})
	UpdateOrganizationMember(ctx)
	var response struct {
		Success bool `json:"success"`
	}
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &response))
	require.False(t, response.Success)
	member, err := model.GetOrganizationMember(1, 4)
	require.NoError(t, err)
	require.Equal(t, model.OrganizationMemberAdmin, member.Role)
}

func TestOrganizationInvitationRevocationIsPendingAndWorkspaceScoped(t *testing.T) {
	setupOrganizationControllerTest(t)
	pending := &model.OrganizationInvitation{OrganizationId: 1, Email: "pending@example.com", Role: model.OrganizationMemberMember, InvitedBy: 1, Status: model.OrganizationInvitationPending, ExpiresAt: common.GetTimestamp() + 3600}
	accepted := &model.OrganizationInvitation{OrganizationId: 1, Email: "accepted@example.com", Role: model.OrganizationMemberMember, InvitedBy: 1, Status: model.OrganizationInvitationAccepted, ExpiresAt: common.GetTimestamp() + 3600}
	require.NoError(t, model.DB.Create(pending).Error)
	require.NoError(t, model.DB.Create(accepted).Error)

	ctx, recorder := organizationTestContext(1, "")
	ctx.Params = append(ctx.Params, gin.Param{Key: "invitation_id", Value: strconv.Itoa(pending.Id)})
	RevokeOrganizationInvitation(ctx)
	var response struct {
		Success bool `json:"success"`
	}
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &response))
	require.True(t, response.Success)
	var stored model.OrganizationInvitation
	require.NoError(t, model.DB.First(&stored, pending.Id).Error)
	require.Equal(t, model.OrganizationInvitationRevoked, stored.Status)

	ctx, recorder = organizationTestContext(1, "")
	ctx.Params = append(ctx.Params, gin.Param{Key: "invitation_id", Value: strconv.Itoa(accepted.Id)})
	RevokeOrganizationInvitation(ctx)
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &response))
	require.False(t, response.Success)
	require.NoError(t, model.DB.First(&stored, accepted.Id).Error)
	require.Equal(t, model.OrganizationInvitationAccepted, stored.Status)
}
