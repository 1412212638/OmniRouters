package controller

import (
	"fmt"
	"strconv"
	"strings"
	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func filterAuditQuery(c *gin.Context, query *gorm.DB) (*gorm.DB, error) {
	if c.FullPath() == "/api/audit/self" { query = query.Where("user_id = ?", c.GetInt("id")) }
	switch c.DefaultQuery("view", "all") {
	case "important": query = query.Where("category = ? OR (category = ? AND action <> ?)", "operation", "access_token", "request")
	case "security": query = query.Where("category = ?", "security")
	case "access": query = query.Where("category = ? OR (category = ? AND action = ?)", "access", "access_token", "request")
	case "all":
	default: return nil, fmt.Errorf("invalid audit view")
	}
	if c.Query("start") != "" && c.Query("end") != "" {
		start, e1 := strconv.ParseInt(c.Query("start"), 10, 64)
		end, e2 := strconv.ParseInt(c.Query("end"), 10, 64)
		if e1 != nil || e2 != nil || end < start { return nil, fmt.Errorf("invalid time range") }
	}
	for _, key := range []string{"user_id", "start", "end"} {
		if raw := c.Query(key); raw != "" {
			value, err := strconv.ParseInt(raw, 10, 64)
			if err != nil || value < 0 { return nil, fmt.Errorf("invalid %s", key) }
			switch key {
			case "user_id": query = query.Where("user_id = ?", value)
			case "start": query = query.Where("created_at >= ?", value)
			case "end": query = query.Where("created_at <= ?", value)
			}
		}
	}
	for _, key := range []string{"username", "action", "ip", "request_id"} {
		if value := strings.TrimSpace(c.Query(key)); value != "" {
			if len(value) > 256 { return nil, fmt.Errorf("invalid %s", key) }
			query = query.Where(map[string]interface{}{key: value})
		}
	}
	if value := c.Query("outcome"); value != "" {
		switch value {
		case "success", "failed", "unknown", "pending": query = query.Where("outcome = ?", value)
		case "legacy": query = query.Where("outcome IS NULL OR outcome = ?", "")
		default: return nil, fmt.Errorf("invalid outcome")
		}
	}
	return query, nil
}

func GetSecurityAuditLogs(c *gin.Context) {
	page := common.GetPageQuery(c)
	if page.Page < 1 || page.Page > 1000000 || page.PageSize < 1 { c.JSON(400, gin.H{"success": false, "message": "invalid pagination"}); return }
	query, err := filterAuditQuery(c, model.DB.Model(&model.AuditLog{}))
	if err != nil { c.JSON(400, gin.H{"success": false, "message": err.Error()}); return }
	var total int64
	if err := query.Count(&total).Error; err != nil { common.ApiError(c, err); return }
	var items []*model.AuditLog
	if err := query.Order("created_at DESC").Order("id DESC").Offset(page.GetStartIdx()).Limit(page.GetPageSize()).Find(&items).Error; err != nil { common.ApiError(c, err); return }
	page.SetTotal(int(total)); page.SetItems(items)
	common.ApiSuccess(c, page)
}
