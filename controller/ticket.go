package controller

import (
	"errors"
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/service"

	"github.com/gin-gonic/gin"
)

func getTicketID(c *gin.Context) (int, bool) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		common.ApiErrorI18n(c, i18n.MsgTicketInvalidId)
		return 0, false
	}
	return id, true
}

func handleTicketError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	if errors.Is(err, service.ErrTicketNotFound) {
		common.ApiErrorI18n(c, i18n.MsgTicketNotFound)
		return
	}
	if errors.Is(err, service.ErrTicketForbidden) {
		common.ApiErrorI18n(c, i18n.MsgTicketForbidden)
		return
	}
	var ticketErr *service.TicketI18nError
	if errors.As(err, &ticketErr) {
		if ticketErr.Params != nil {
			common.ApiErrorI18n(c, ticketErr.Key, ticketErr.Params)
		} else {
			common.ApiErrorI18n(c, ticketErr.Key)
		}
		return
	}
	common.ApiError(c, err)
}

func ListTickets(c *gin.Context) {
	userID := c.GetInt("id")
	pageInfo := common.GetPageQuery(c)
	resp, err := service.ListTickets(
		userID,
		false,
		pageInfo,
		c.Query("status"),
		c.Query("category"),
		c.Query("priority"),
		c.Query("keyword"),
		0,
	)
	if err != nil {
		handleTicketError(c, err)
		return
	}
	common.ApiSuccess(c, resp)
}

func AdminListTickets(c *gin.Context) {
	adminID := c.GetInt("id")
	pageInfo := common.GetPageQuery(c)
	assignedAdminID, _ := strconv.Atoi(c.Query("assigned_admin_id"))
	resp, err := service.ListTickets(
		adminID,
		true,
		pageInfo,
		c.Query("status"),
		c.Query("category"),
		c.Query("priority"),
		c.Query("keyword"),
		assignedAdminID,
	)
	if err != nil {
		handleTicketError(c, err)
		return
	}
	common.ApiSuccess(c, resp)
}

func CreateTicket(c *gin.Context) {
	req := dto.TicketCreateRequest{}
	if err := common.UnmarshalBodyReusable(c, &req); err != nil {
		common.ApiError(c, err)
		return
	}
	resp, err := service.CreateTicket(c.GetInt("id"), req)
	if err != nil {
		handleTicketError(c, err)
		return
	}
	common.ApiSuccess(c, resp)
}

func GetTicket(c *gin.Context) {
	id, ok := getTicketID(c)
	if !ok {
		return
	}
	resp, err := service.GetTicketDetail(c.GetInt("id"), false, id)
	if err != nil {
		handleTicketError(c, err)
		return
	}
	common.ApiSuccess(c, resp)
}

func AdminGetTicket(c *gin.Context) {
	id, ok := getTicketID(c)
	if !ok {
		return
	}
	resp, err := service.GetTicketDetail(c.GetInt("id"), true, id)
	if err != nil {
		handleTicketError(c, err)
		return
	}
	common.ApiSuccess(c, resp)
}

func AddTicketMessage(c *gin.Context) {
	id, ok := getTicketID(c)
	if !ok {
		return
	}
	req := dto.TicketMessageRequest{}
	if err := common.UnmarshalBodyReusable(c, &req); err != nil {
		common.ApiError(c, err)
		return
	}
	resp, err := service.AddTicketMessage(c.GetInt("id"), false, id, req)
	if err != nil {
		handleTicketError(c, err)
		return
	}
	common.ApiSuccess(c, resp)
}

func AdminAddTicketMessage(c *gin.Context) {
	id, ok := getTicketID(c)
	if !ok {
		return
	}
	req := dto.TicketMessageRequest{}
	if err := common.UnmarshalBodyReusable(c, &req); err != nil {
		common.ApiError(c, err)
		return
	}
	resp, err := service.AddTicketMessage(c.GetInt("id"), true, id, req)
	if err != nil {
		handleTicketError(c, err)
		return
	}
	common.ApiSuccess(c, resp)
}

func AdminUpdateTicketStatus(c *gin.Context) {
	id, ok := getTicketID(c)
	if !ok {
		return
	}
	req := dto.TicketStatusRequest{}
	if err := common.UnmarshalBodyReusable(c, &req); err != nil {
		common.ApiError(c, err)
		return
	}
	resp, err := service.UpdateTicketStatus(c.GetInt("id"), id, req.Status)
	if err != nil {
		handleTicketError(c, err)
		return
	}
	common.ApiSuccess(c, resp)
}

func AdminAssignTicket(c *gin.Context) {
	id, ok := getTicketID(c)
	if !ok {
		return
	}
	req := dto.TicketAssignRequest{}
	if err := common.UnmarshalBodyReusable(c, &req); err != nil {
		common.ApiError(c, err)
		return
	}
	resp, err := service.AssignTicket(c.GetInt("id"), id, req.AssignedAdminID)
	if err != nil {
		handleTicketError(c, err)
		return
	}
	common.ApiSuccess(c, resp)
}

func CloseTicket(c *gin.Context) {
	id, ok := getTicketID(c)
	if !ok {
		return
	}
	resp, err := service.CloseTicket(c.GetInt("id"), false, id)
	if err != nil {
		handleTicketError(c, err)
		return
	}
	common.ApiSuccess(c, resp)
}

func AdminCloseTicket(c *gin.Context) {
	id, ok := getTicketID(c)
	if !ok {
		return
	}
	resp, err := service.CloseTicket(c.GetInt("id"), true, id)
	if err != nil {
		handleTicketError(c, err)
		return
	}
	common.ApiSuccess(c, resp)
}

func ReopenTicket(c *gin.Context) {
	id, ok := getTicketID(c)
	if !ok {
		return
	}
	resp, err := service.ReopenTicket(c.GetInt("id"), id)
	if err != nil {
		handleTicketError(c, err)
		return
	}
	common.ApiSuccess(c, resp)
}
