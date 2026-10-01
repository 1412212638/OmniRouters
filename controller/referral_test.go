package controller

import (
	"testing"

	"github.com/QuantumNous/new-api/model"
	"github.com/stretchr/testify/require"
)

func TestMaskReferralName(t *testing.T) {
	for input, expected := range map[string]string{
		"":       "",
		"a":      "***",
		"ab":     "***",
		"alice":  "a***e",
		"张三":     "***",
		"张三丰":   "张***丰",
	} {
		require.Equal(t, expected, maskReferralName(input))
	}
}

func TestBuildReferralFriendResponseMasksPersonalData(t *testing.T) {
	response := buildReferralFriendResponse(model.ReferralFriend{User: model.User{
		Id: 7, Username: "alice", DisplayName: "Alice", Email: "alice@example.com",
	}})
	require.Equal(t, "a***e", response.Username)
	require.Equal(t, "A***e", response.DisplayName)
	require.Equal(t, "***@example.com", response.Email)
	require.Equal(t, "pending", response.FirstTopUpStatus)
}
