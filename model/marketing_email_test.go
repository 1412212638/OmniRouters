package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/require"
)

func TestMarketingEmailAudienceFilters(t *testing.T) {
	truncateTables(t)
	now := common.GetTimestamp()
	users := []User{
		{Id: 1, Username: "paid-active", AffCode: "marketing-1", Status: common.UserStatusEnabled, LastLoginAt: now - 10, Group: "default"},
		{Id: 2, Username: "unpaid-active", AffCode: "marketing-2", Status: common.UserStatusEnabled, LastLoginAt: now - 20, Group: "premium"},
		{Id: 3, Username: "paid-old", AffCode: "marketing-3", Status: common.UserStatusEnabled, LastLoginAt: now - 31*24*60*60, Group: "default"},
		{Id: 4, Username: "disabled", AffCode: "marketing-4", Status: common.UserStatusDisabled, LastLoginAt: now - 5, Group: "default"},
	}
	for i := range users {
		require.NoError(t, DB.Create(&users[i]).Error)
	}
	require.NoError(t, DB.Create(&TopUp{UserId: 1, TradeNo: "marketing-topup-1", Status: common.TopUpStatusSuccess}).Error)
	require.NoError(t, DB.Create(&TopUp{UserId: 3, TradeNo: "marketing-topup-3", Status: common.TopUpStatusSuccess}).Error)

	all, err := GetMarketingEmailAudienceUserIDs("all", "", nil, 100)
	require.NoError(t, err)
	require.Equal(t, []int{1, 2, 3}, all)

	paid, err := GetMarketingEmailAudienceUserIDs("topup", "", nil, 100)
	require.NoError(t, err)
	require.Equal(t, []int{1, 3}, paid)

	active, err := GetMarketingEmailAudienceUserIDs("active", "", nil, 100)
	require.NoError(t, err)
	require.Equal(t, []int{1, 2, 4}, active)

	group, err := GetMarketingEmailAudienceUserIDs("group", "premium", nil, 100)
	require.NoError(t, err)
	require.Equal(t, []int{2}, group)

	manual, err := GetMarketingEmailAudienceUserIDs("manual", "", []int{3, 1, 3}, 100)
	require.NoError(t, err)
	require.Equal(t, []int{1, 3}, manual)
}
