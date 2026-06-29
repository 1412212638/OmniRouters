package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserBaseExtraGroupsRoundTrip(t *testing.T) {
	user := User{
		ExtraGroups: StringList{"seed", " seed ", "", "vip"},
	}

	cached := user.ToBaseUser()

	assert.JSONEq(t, `["seed","vip"]`, cached.ExtraGroups)
	assert.Equal(t, []string{"seed", "vip"}, cached.GetExtraGroups())
}
