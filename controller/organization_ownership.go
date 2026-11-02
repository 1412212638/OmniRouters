package controller

import (
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
)

func TransferOrganizationOwnership(c *gin.Context) {
	organizationId, err := strconv.Atoi(c.Param("organization_id"))
	if err != nil || organizationId <= 0 {
		common.ApiErrorMsg(c, "invalid organization id")
		return
	}
	var request struct {
		UserId int `json:"user_id"`
	}
	if err := c.ShouldBindJSON(&request); err != nil || request.UserId <= 0 {
		common.ApiErrorMsg(c, "invalid user id")
		return
	}
	if err := model.TransferOrganizationOwnership(organizationId, c.GetInt("id"), request.UserId); err != nil {
		common.ApiErrorMsg(c, err.Error())
		return
	}
	service.TryNotifyOrganization(service.OrganizationNotification{Kind: service.OrganizationNotificationOwnershipTransferred, OrganizationId: organizationId, UserId: request.UserId, Role: model.OrganizationMemberOwner})
	service.TryNotifyOrganization(service.OrganizationNotification{Kind: service.OrganizationNotificationOwnershipTransferred, OrganizationId: organizationId, UserId: c.GetInt("id"), Role: model.OrganizationMemberAdmin})
	common.ApiSuccess(c, nil)
}
