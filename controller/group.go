package controller

import (
	"net/http"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/ratio_setting"

	"github.com/gin-gonic/gin"
)

func GetGroups(c *gin.Context) {
	groupNames := make([]string, 0)
	for groupName := range ratio_setting.GetGroupRatioCopy() {
		groupNames = append(groupNames, groupName)
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    groupNames,
	})
}

func GetUserGroups(c *gin.Context) {
	usableGroups := make(map[string]map[string]interface{})
	if organizationId := common.GetContextKeyInt(c, constant.ContextKeyOrganizationId); organizationId > 0 {
		organization, err := model.GetOrganizationById(organizationId)
		if err != nil {
			common.ApiError(c, err)
			return
		}
		usableGroups[organization.Group] = map[string]interface{}{
			"ratio": service.GetUserGroupRatio(organization.Group, organization.Group),
			"desc":  setting.GetUsableGroupDescription(organization.Group),
		}
		common.ApiSuccess(c, usableGroups)
		return
	}
	userId := c.GetInt("id")
	user, err := model.GetUserCache(userId)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	userGroup := user.Group
	userUsableGroups := service.GetUserUsableGroupsWithExtras(userGroup, user.GetExtraGroups())
	for groupName, _ := range ratio_setting.GetGroupRatioCopy() {
		// UserUsableGroups contains the groups that the user can use
		if desc, ok := userUsableGroups[groupName]; ok {
			usableGroups[groupName] = map[string]interface{}{
				"ratio": service.GetUserGroupRatio(userGroup, groupName),
				"desc":  desc,
			}
		}
	}
	if _, ok := userUsableGroups["auto"]; ok {
		usableGroups["auto"] = map[string]interface{}{
			"ratio": "自动",
			"desc":  setting.GetUsableGroupDescription("auto"),
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    usableGroups,
	})
}
