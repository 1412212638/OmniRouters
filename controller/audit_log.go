package controller

import (
	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

func GetSecurityAuditLogs(c *gin.Context) {
	page := common.GetPageQuery(c)
	query := model.DB.Model(&model.AuditLog{}).Order("created_at DESC").Order("id DESC")
	if c.FullPath() == "/api/audit/self" { query = query.Where("user_id = ?", c.GetInt("id")) }
	var total int64
	if err := query.Count(&total).Error; err != nil { common.ApiError(c, err); return }
	var items []*model.AuditLog
	if err := query.Offset(page.GetStartIdx()).Limit(page.GetPageSize()).Find(&items).Error; err != nil { common.ApiError(c, err); return }
	page.SetTotal(int(total)); page.SetItems(items)
	common.ApiSuccess(c, page)
}
