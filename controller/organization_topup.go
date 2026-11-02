package controller

import (
	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

func validateTopUpWallet(c *gin.Context, userId, quota int) error {
	if id := common.GetContextKeyInt(c, constant.ContextKeyOrganizationId); id > 0 {
		return model.ValidateOrganizationTopUp(id, userId, quota)
	}
	return model.ValidateTopUpQuotaCapacity(userId, quota)
}

func topUpGroup(c *gin.Context, userId int) (string, error) {
	if id := common.GetContextKeyInt(c, constant.ContextKeyOrganizationId); id > 0 {
		organization, err := model.GetOrganizationById(id)
		if err != nil {
			return "", err
		}
		return organization.Group, nil
	}
	return model.GetUserGroup(userId, true)
}
