package service

import (
	"context"
	"testing"

	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func seedOrganizationTaskFunding(t *testing.T, organizationID, userID int, period string) {
	t.Helper()
	require.NoError(t, model.DB.AutoMigrate(&model.Organization{}, &model.OrganizationMember{}, &model.OrganizationMemberUsage{}))
	require.NoError(t, model.DB.Create(&model.Organization{Id: organizationID, Name: "Billing team", OwnerUserId: userID, Quota: 1000}).Error)
	require.NoError(t, model.DB.Create(&model.OrganizationMember{OrganizationId: organizationID, UserId: userID, Role: model.OrganizationMemberOwner, Status: model.OrganizationMemberActive}).Error)
	t.Cleanup(func() {
		model.DB.Where("organization_id = ?", organizationID).Delete(&model.OrganizationMemberUsage{})
		model.DB.Where("organization_id = ?", organizationID).Delete(&model.OrganizationMember{})
		model.DB.Delete(&model.Organization{}, organizationID)
	})
	reserved, err := model.TryReserveOrganizationQuota(organizationID, userID, 200, period)
	require.NoError(t, err)
	require.True(t, reserved)
}

func TestOrganizationTaskSettlementKeepsFundingAndUsageInOriginalWorkspace(t *testing.T) {
	truncate(t)
	const userID, organizationID, channelID = 960, 960, 960
	const period = "2025-01"
	seedUser(t, userID, 777)
	seedChannel(t, channelID)
	seedOrganizationTaskFunding(t, organizationID, userID, period)
	require.NoError(t, model.DB.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]any{"used_quota": 42, "request_count": 3}).Error)
	require.NoError(t, model.DB.Model(&model.Channel{}).Where("id = ?", channelID).Update("used_quota", 200).Error)

	task := makeTask(userID, channelID, 200, 0, BillingSourceOrganization, 0)
	task.PrivateData.OrganizationId = organizationID
	task.PrivateData.OrganizationUsagePeriod = period
	require.NoError(t, task.Insert())
	// Removing the member or deleting their key must not reroute an already
	// accepted task to their personal wallet during later polling.
	require.NoError(t, model.DB.Model(&model.OrganizationMember{}).Where("organization_id = ?", organizationID).Update("status", model.OrganizationMemberRemoved).Error)

	RecalculateTaskQuota(context.Background(), task, 300, "actual usage")
	organization, err := model.GetOrganizationById(organizationID)
	require.NoError(t, err)
	assert.Equal(t, 700, organization.Quota)
	assert.Equal(t, 300, organization.UsedQuota)
	assert.Equal(t, 777, getUserQuota(t, userID))
	used, requests := getUserUsageAccounting(t, userID)
	assert.Equal(t, 42, used)
	assert.Equal(t, 3, requests)
	assert.Equal(t, organizationID, getLastLog(t).OrganizationId)

	var usage model.OrganizationMemberUsage
	require.NoError(t, model.DB.Where("organization_id = ? AND user_id = ? AND period = ?", organizationID, userID, period).First(&usage).Error)
	assert.Equal(t, 300, usage.UsedQuota)
	require.True(t, RefundTaskQuota(context.Background(), task, "failed"))
	organization, err = model.GetOrganizationById(organizationID)
	require.NoError(t, err)
	assert.Equal(t, 1000, organization.Quota)
	assert.Zero(t, organization.UsedQuota)
	assert.Equal(t, 777, getUserQuota(t, userID))
	assert.Equal(t, organizationID, getLastLog(t).OrganizationId)
}

func TestOrganizationMidjourneyRefundDoesNotCreditPersonalWallet(t *testing.T) {
	truncate(t)
	const userID, organizationID, channelID = 961, 961, 961
	const period = "2025-01"
	seedUser(t, userID, 777)
	seedChannel(t, channelID)
	seedOrganizationTaskFunding(t, organizationID, userID, period)
	require.NoError(t, model.DB.Model(&model.Channel{}).Where("id = ?", channelID).Update("used_quota", 200).Error)
	info := &relaycommon.RelayInfo{UserId: userID, OrganizationId: organizationID, OrganizationUsagePeriod: period}
	task := &model.Midjourney{UserId: userID, MjId: "organization-refund", ChannelId: channelID}
	prepared, err := PrepareMidjourneyTaskBilling(info, task, 200, true)
	require.NoError(t, err)
	require.True(t, prepared)
	require.NoError(t, task.Insert())

	require.True(t, RefundMidjourneyQuota(context.Background(), task, "failed"))
	organization, err := model.GetOrganizationById(organizationID)
	require.NoError(t, err)
	assert.Equal(t, 1000, organization.Quota)
	assert.Zero(t, organization.UsedQuota)
	assert.Equal(t, 777, getUserQuota(t, userID))
	assert.Equal(t, organizationID, getLastLog(t).OrganizationId)
}

func TestOrganizationTaskMissingOrganizationDoesNotFallBackToPersonalWallet(t *testing.T) {
	truncate(t)
	seedUser(t, 962, 777)
	task := makeTask(962, 0, 200, 0, BillingSourceOrganization, 0)
	require.Error(t, taskAdjustFunding(task, -200))
	assert.Equal(t, 777, getUserQuota(t, 962))
}

func TestMidjourneyQueriesRequireTheSameWorkspace(t *testing.T) {
	truncate(t)
	for _, task := range []*model.Midjourney{
		{UserId: 963, MjId: "personal-mj", OrganizationId: 0},
		{UserId: 963, MjId: "team-mj", OrganizationId: 7},
	} {
		require.NoError(t, task.Insert())
	}
	assert.Nil(t, model.GetByMJId(963, "team-mj"))
	assert.Nil(t, model.GetByMJId(963, "personal-mj", 7))
	require.NotNil(t, model.GetByMJId(963, "team-mj", 7))
	team := model.GetByMJIds(963, []string{"personal-mj", "team-mj"}, 7)
	require.Len(t, team, 1)
	assert.Equal(t, "team-mj", team[0].MjId)
	personal := model.GetAllUserTask(963, 0, 20, model.TaskQueryParams{})
	require.Len(t, personal, 1)
	assert.Equal(t, "personal-mj", personal[0].MjId)
	assert.EqualValues(t, 1, model.CountAllUserTask(963, model.TaskQueryParams{}))
}
