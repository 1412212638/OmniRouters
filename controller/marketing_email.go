package controller

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"

	"github.com/gin-gonic/gin"
)

type marketingEmailSendRequest struct {
	AudienceType    string `json:"audience_type"`
	AudienceGroup   string `json:"audience_group"`
	UserIds         []int  `json:"user_ids"`
	SubjectTemplate string `json:"subject_template"`
	ContentTemplate string `json:"content_template"`
}

func GetMarketingEmailSendTask(c *gin.Context) {
	task, err := model.GetSystemTaskByTaskID(c.Param("task_id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if task == nil || task.Type != model.SystemTaskTypeMarketingEmailSend {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "task not found"})
		return
	}
	response := task.ToResponse()
	response.Payload = nil
	common.ApiSuccess(c, response)
}

func SendMarketingEmail(c *gin.Context) {
	var req marketingEmailSendRequest
	if err := common.DecodeJson(c.Request.Body, &req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	if req.AudienceType == "" {
		req.AudienceType = "manual"
	}
	if req.AudienceType == "manual" && len(req.UserIds) == 0 {
		common.ApiErrorMsg(c, "请选择至少一个收件人")
		return
	}
	if strings.TrimSpace(req.SubjectTemplate) == "" && strings.TrimSpace(common.MarketingEmailSubjectTemplate) == "" {
		common.ApiErrorMsg(c, "邮件标题不能为空")
		return
	}
	if strings.TrimSpace(req.ContentTemplate) == "" && strings.TrimSpace(common.MarketingEmailContentTemplate) == "" {
		common.ApiErrorMsg(c, "邮件内容不能为空")
		return
	}

	task, err := service.StartMarketingEmailSend(req.AudienceType, req.AudienceGroup, req.UserIds, req.SubjectTemplate, req.ContentTemplate)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	operatorId := c.GetInt("id")
	model.RecordLog(operatorId, model.LogTypeSystem, fmt.Sprintf("开始营销邮件发送任务，任务 ID: %s", task.TaskID))
	common.ApiSuccess(c, gin.H{"task_id": task.TaskID, "status": task.Status})
}

func GetMarketingEmailAudience(c *gin.Context) {
	audienceType := c.Query("audience_type")
	if audienceType == "" {
		audienceType = "all"
	}
	total, err := service.GetMarketingEmailAudienceCount(audienceType, c.Query("audience_group"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{"total": total})
}
