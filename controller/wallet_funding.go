package controller

import (
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

// Self access returns aggregates only. Internal notes and operator information
// belong exclusively to the administrative ledger endpoint.
func GetSelfWalletFunding(c *gin.Context) {
	_, _, summary, err := model.GetWalletFunding(c.GetInt("id"), 0, 1, "")
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, summary)
}

func GetUserWalletFunding(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}
	user, err := model.GetUserById(id, false)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if !canManageTargetRole(c.GetInt("role"), user.Role) {
		common.ApiErrorI18n(c, i18n.MsgUserNoPermissionSameLevel)
		return
	}
	page := common.GetPageQuery(c)
	if page.Page < 1 {
		page.Page = 1
	}
	if page.Page > 1000000 {
		page.Page = 1000000
	}
	if page.PageSize < 1 {
		page.PageSize = 20
	}
	rows, total, summary, err := model.GetWalletFunding(id, page.GetStartIdx(), page.GetPageSize(), c.Query("kind"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{"items": rows, "total": total, "page": page.GetPage(), "page_size": page.GetPageSize(), "summary": summary})
}
