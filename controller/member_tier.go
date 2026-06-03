package controller

import (
	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/service"

	"github.com/gin-gonic/gin"
)

func GetMemberTierSelf(c *gin.Context) {
	userId := c.GetInt("id")
	evaluation, err := service.EvaluateMemberTierForUser(userId, false)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	common.ApiSuccess(c, evaluation)
}

func AdminRecalculateMemberTiers(c *gin.Context) {
	result, err := service.RecalculateAllMemberTiers()
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, result)
}
