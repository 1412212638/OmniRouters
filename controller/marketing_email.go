package controller

import (
	"fmt"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"

	"github.com/gin-gonic/gin"
)

type marketingEmailSendRequest struct {
	UserIds         []int  `json:"user_ids"`
	SubjectTemplate string `json:"subject_template"`
	ContentTemplate string `json:"content_template"`
}

func SendMarketingEmail(c *gin.Context) {
	var req marketingEmailSendRequest
	if err := common.DecodeJson(c.Request.Body, &req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	if len(req.UserIds) == 0 {
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

	result, err := service.SendMarketingEmailToUsers(req.UserIds, req.SubjectTemplate, req.ContentTemplate)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	operatorId := c.GetInt("id")
	model.RecordLog(operatorId, model.LogTypeSystem, fmt.Sprintf("发送营销邮件，目标用户: %d，成功: %d，跳过: %d，失败: %d", result.Total, result.Sent, result.Skipped, result.Failed))
	common.ApiSuccess(c, result)
}
