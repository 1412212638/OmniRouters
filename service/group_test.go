package service

import (
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetUserUsableGroupsWithExtras(t *testing.T) {
	groups := GetUserUsableGroupsWithExtras(
		"default",
		[]string{"svip", " svip ", "auto", "missing-group", "default"},
	)

	assert.Contains(t, groups, "default")
	assert.Contains(t, groups, "svip")
	assert.NotContains(t, groups, "auto")
	assert.NotContains(t, groups, "missing-group")
}

func TestGetUserUsableGroupsForContextIncludesExtras(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	common.SetContextKey(ctx, constant.ContextKeyUserGroup, "default")
	common.SetContextKey(ctx, constant.ContextKeyUserExtraGroups, []string{"svip"})

	assert.True(t, GroupInUserUsableGroupsForContext(ctx, "svip"))
}
