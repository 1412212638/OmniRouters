package model

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubscriptionGroupListNormalizeValueAndScan(t *testing.T) {
	groups := NormalizeSubscriptionGroups([]string{" economy ", "official", "economy", "", "auto"})
	assert.Equal(t, SubscriptionGroupList{"economy", "official"}, groups)
	assert.True(t, groups.Allows("economy"))
	assert.False(t, groups.Allows("other"))
	assert.True(t, SubscriptionGroupList{}.Allows("other"))

	value, err := groups.Value()
	require.NoError(t, err)
	assert.JSONEq(t, `["economy","official"]`, value.(string))

	var scanned SubscriptionGroupList
	require.NoError(t, scanned.Scan(value))
	assert.Equal(t, groups, scanned)
}

func seedSubscriptionGroupPlan(t *testing.T, id int, groups SubscriptionGroupList) *SubscriptionPlan {
	t.Helper()
	plan := &SubscriptionPlan{
		Id:               id,
		Title:            "Group plan",
		PriceAmount:      10,
		DurationUnit:     SubscriptionDurationMonth,
		DurationValue:    1,
		TotalAmount:      1000,
		QuotaResetPeriod: SubscriptionResetNever,
		AllowedGroups:    groups,
		Enabled:          true,
	}
	require.NoError(t, DB.Create(plan).Error)
	return plan
}

func seedSubscriptionGroupSubscription(t *testing.T, id, userId, planId int, groups SubscriptionGroupList, used int64) {
	t.Helper()
	now := GetDBTimestamp()
	require.NoError(t, DB.Create(&UserSubscription{
		Id:            id,
		UserId:        userId,
		PlanId:        planId,
		AmountTotal:   1000,
		AmountUsed:    used,
		StartTime:     now - 60,
		EndTime:       now + 3600,
		Status:        "active",
		AllowedGroups: groups,
	}).Error)
}

func TestSubscriptionAllowedGroupsAreSnapshottedAtPurchase(t *testing.T) {
	truncateTables(t)
	plan := seedSubscriptionGroupPlan(t, 9711, SubscriptionGroupList{"economy", "official"})

	sub, err := CreateUserSubscriptionFromPlanTx(DB, 9712, plan, "admin")
	require.NoError(t, err)
	assert.Equal(t, SubscriptionGroupList{"economy", "official"}, sub.AllowedGroups)

	plan.AllowedGroups = SubscriptionGroupList{"official"}
	require.NoError(t, DB.Save(plan).Error)
	var stored UserSubscription
	require.NoError(t, DB.First(&stored, sub.Id).Error)
	assert.Equal(t, SubscriptionGroupList{"economy", "official"}, stored.AllowedGroups)
}

func TestPreConsumeSelectsSubscriptionMatchingActualGroup(t *testing.T) {
	truncateTables(t)
	seedSubscriptionGroupPlan(t, 9721, SubscriptionGroupList{"economy"})
	seedSubscriptionGroupPlan(t, 9722, SubscriptionGroupList{"official"})
	seedSubscriptionGroupSubscription(t, 9723, 9720, 9721, SubscriptionGroupList{"economy"}, 0)
	seedSubscriptionGroupSubscription(t, 9724, 9720, 9722, SubscriptionGroupList{"official"}, 0)

	result, err := PreConsumeUserSubscription("group-official", 9720, "official", "gpt-test", 0, 100)
	require.NoError(t, err)
	assert.Equal(t, 9724, result.UserSubscriptionId)
	assert.Equal(t, SubscriptionGroupList{"official"}, result.AllowedGroups)

	var economy, official UserSubscription
	require.NoError(t, DB.First(&economy, 9723).Error)
	require.NoError(t, DB.First(&official, 9724).Error)
	assert.Zero(t, economy.AmountUsed)
	assert.EqualValues(t, 100, official.AmountUsed)
}

func TestPreConsumeSupportsSmartPlanAndReportsRestrictionErrors(t *testing.T) {
	truncateTables(t)
	seedSubscriptionGroupPlan(t, 9731, SubscriptionGroupList{"economy", "official"})
	seedSubscriptionGroupSubscription(t, 9732, 9730, 9731, SubscriptionGroupList{"economy", "official"}, 0)

	_, err := PreConsumeUserSubscription("smart-economy", 9730, "economy", "gpt-test", 0, 100)
	require.NoError(t, err)
	_, err = PreConsumeUserSubscription("smart-official", 9730, "official", "gpt-test", 0, 100)
	require.NoError(t, err)

	_, err = PreConsumeUserSubscription("smart-other", 9730, "other", "gpt-test", 0, 100)
	assert.ErrorIs(t, err, ErrSubscriptionGroupNotAllowed)
	_, err = PreConsumeUserSubscription("smart-insufficient", 9730, "economy", "gpt-test", 0, 900)
	assert.ErrorIs(t, err, ErrSubscriptionQuotaInsufficient)
	assert.False(t, errors.Is(err, ErrSubscriptionGroupNotAllowed))
}
