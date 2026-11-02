package controller

import (
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

func GetOrganizationLogs(c *gin.Context) {
	organizationId, err := strconv.Atoi(c.Param("organization_id"))
	if err != nil || organizationId <= 0 {
		common.ApiErrorMsg(c, "invalid organization id")
		return
	}
	member, err := organizationMember(c, organizationId)
	if err != nil {
		common.ApiErrorMsg(c, err.Error())
		return
	}
	userId := c.GetInt("id")
	if model.OrganizationCanManage(member.Role) {
		userId, _ = strconv.Atoi(c.Query("user_id"))
	}
	page := common.GetPageQuery(c)
	logType, _ := strconv.Atoi(c.Query("type"))
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	logs, total, err := model.GetOrganizationLogs(organizationId, userId, logType, startTimestamp, endTimestamp, c.Query("model_name"), c.Query("token_name"), page.GetStartIdx(), page.GetPageSize())
	if err != nil {
		common.ApiError(c, err)
		return
	}
	page.SetTotal(int(total))
	page.SetItems(logs)
	common.ApiSuccess(c, page)
}
