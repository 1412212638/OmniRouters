package model

import (
	"context"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/stretchr/testify/require"
)

func TestWorkspaceUsageQueriesKeepPersonalAndOrganizationDataSeparate(t *testing.T) {
	truncateTables(t)
	for _, log := range []*Log{
		{UserId: 1, OrganizationId: 0, Type: LogTypeConsume, ModelName: "personal", Quota: 10},
		{UserId: 1, OrganizationId: 7, Type: LogTypeConsume, ModelName: "team", Quota: 20},
		{UserId: 2, OrganizationId: 7, Type: LogTypeConsume, ModelName: "teammate", Quota: 30},
		{UserId: 1, OrganizationId: 8, Type: LogTypeConsume, ModelName: "other-team", Quota: 40},
	} {
		require.NoError(t, createLog(log))
	}
	personal, total, err := GetUserLogs(1, LogTypeConsume, 0, 0, "", "", 0, 20, "", "", "")
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, personal, 1)
	require.Equal(t, "personal", personal[0].ModelName)
	workspaceLogs, total, err := GetUserLogs(1, LogTypeConsume, 0, 0, "", "", 0, 20, "", "", "", 7)
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, workspaceLogs, 1)
	require.Equal(t, "team", workspaceLogs[0].ModelName)
	team, total, err := GetOrganizationLogs(7, 0, LogTypeConsume, 0, 0, "", "", 0, 20)
	require.NoError(t, err)
	require.EqualValues(t, 2, total)
	require.Len(t, team, 2)
	member, total, err := GetOrganizationLogs(7, 1, LogTypeConsume, 0, 0, "", "", 0, 20)
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, member, 1)
	require.Equal(t, "team", member[0].ModelName)

	for _, row := range []QuotaData{
		{UserID: 1, OrganizationID: 0, ModelName: "personal", UseGroup: "default", CreatedAt: 1000, Quota: 10, Count: 1},
		{UserID: 1, OrganizationID: 7, ModelName: "team", UseGroup: "default", CreatedAt: 1000, Quota: 20, Count: 1},
	} {
		require.NoError(t, DB.Create(&row).Error)
	}
	dashboard, err := GetQuotaDataByUserId(1, 0, 2000)
	require.NoError(t, err)
	require.Len(t, dashboard, 1)
	require.Equal(t, "personal", dashboard[0].ModelName)
	flow, err := GetFlowQuotaData(0, 2000, "", 1, common.RoleCommonUser)
	require.NoError(t, err)
	require.Len(t, flow, 1)
	require.Equal(t, "personal", flow[0].ModelName)
	dashboard, err = GetQuotaDataByUserId(1, 0, 2000, 7)
	require.NoError(t, err)
	require.Len(t, dashboard, 1)
	require.Equal(t, "team", dashboard[0].ModelName)
	flow, err = GetFlowQuotaData(0, 2000, "", 1, common.RoleCommonUser, 7)
	require.NoError(t, err)
	require.Len(t, flow, 1)
	require.Equal(t, "team", flow[0].ModelName)
}

func TestTaskQueriesRequireTheSameWorkspace(t *testing.T) {
	truncateTables(t)
	const userID = 1101
	for _, task := range []*Task{
		{TaskID: "personal-task", UserId: userID, OrganizationId: 0, Platform: constant.TaskPlatform("test")},
		{TaskID: "team-task", UserId: userID, OrganizationId: 7, Platform: constant.TaskPlatform("test")},
	} {
		require.NoError(t, DB.Create(task).Error)
	}
	_, exists, err := GetByTaskId(userID, "team-task")
	require.NoError(t, err)
	require.False(t, exists)
	_, exists, err = GetByTaskId(userID, "personal-task", 7)
	require.NoError(t, err)
	require.False(t, exists)
	team, exists, err := GetByTaskId(userID, "team-task", 7)
	require.NoError(t, err)
	require.True(t, exists)
	require.Equal(t, 7, team.OrganizationId)
	rows, err := GetByTaskIdsForPlatforms(userID, []constant.TaskPlatform{"test"}, []string{"personal-task", "team-task"}, 7)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Equal(t, "team-task", rows[0].TaskID)
	_, exists, err = GetTaskForProtocolObservation(context.Background(), userID, "test", "team-task")
	require.NoError(t, err)
	require.False(t, exists)
	ctx := context.WithValue(context.Background(), constant.ContextKeyOrganizationId, 7)
	_, exists, err = GetTaskForProtocolObservation(ctx, userID, "test", "team-task")
	require.NoError(t, err)
	require.True(t, exists)
	personal := TaskGetAllUserTask(userID, 0, 20, SyncTaskQueryParams{})
	require.Len(t, personal, 1)
	require.Equal(t, "personal-task", personal[0].TaskID)
	require.EqualValues(t, 1, TaskCountAllUserTask(userID, SyncTaskQueryParams{}))
	organization := TaskGetAllUserTask(userID, 0, 20, SyncTaskQueryParams{OrganizationID: 7})
	require.Len(t, organization, 1)
	require.Equal(t, "team-task", organization[0].TaskID)
	require.EqualValues(t, 1, TaskCountAllUserTask(userID, SyncTaskQueryParams{OrganizationID: 7}))
}

func TestStandardTokenQueriesRequireOwnerAndWorkspace(t *testing.T) {
	truncateTables(t)
	for _, token := range []*Token{
		{UserId: 1, Name: "personal", Key: "personal-scope-key", OrganizationId: 0},
		{UserId: 1, Name: "team", Key: "team-scope-key", OrganizationId: 7},
		{UserId: 2, Name: "teammate", Key: "teammate-scope-key", OrganizationId: 7},
		{UserId: 1, Name: "another-team", Key: "other-team-scope-key", OrganizationId: 8},
	} {
		require.NoError(t, DB.Create(token).Error)
	}
	rows, err := GetAllUserTokens(1, 0, 20, 7)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Equal(t, "team", rows[0].Name)
	teamTokenID := rows[0].Id
	_, err = GetTokenByIds(teamTokenID, 1)
	require.Error(t, err)
	_, err = GetTokenByIds(teamTokenID, 2, 7)
	require.Error(t, err)
	search, total, err := SearchUserTokens(1, "team", "", 0, 20, 7)
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, search, 1)
	require.Equal(t, teamTokenID, search[0].Id)
	_, err = BatchDeleteTokens([]int{teamTokenID}, 1)
	require.NoError(t, err)
	_, err = GetTokenByIds(teamTokenID, 1, 7)
	require.NoError(t, err, "personal batch actions cannot mutate organization keys")
}
