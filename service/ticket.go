package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/model"

	"gorm.io/gorm"
)

var (
	ErrTicketNotFound  = errors.New("ticket not found")
	ErrTicketForbidden = errors.New("you do not have permission to access this ticket")
)

var validTicketStatuses = map[string]bool{
	model.TicketStatusOpen:     true,
	model.TicketStatusPending:  true,
	model.TicketStatusAnswered: true,
	model.TicketStatusClosed:   true,
}

var validTicketPriorities = map[string]bool{
	model.TicketPriorityLow:    true,
	model.TicketPriorityNormal: true,
	model.TicketPriorityHigh:   true,
	model.TicketPriorityUrgent: true,
}

func normalizeTicketCategory(category string) string {
	category = strings.TrimSpace(strings.ToLower(category))
	if category == "" {
		return "general"
	}
	if len(category) > 64 {
		return category[:64]
	}
	return category
}

func normalizeTicketPriority(priority string) string {
	priority = strings.TrimSpace(strings.ToLower(priority))
	if !validTicketPriorities[priority] {
		return model.TicketPriorityNormal
	}
	return priority
}

func normalizeTicketStatus(status string) (string, error) {
	status = strings.TrimSpace(strings.ToLower(status))
	if !validTicketStatuses[status] {
		return "", fmt.Errorf("invalid ticket status: %s", status)
	}
	return status, nil
}

func encodeTicketAttachments(attachments []dto.TicketAttachment) (string, error) {
	if len(attachments) == 0 {
		return "", nil
	}
	data, err := common.Marshal(attachments)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func decodeTicketAttachments(raw string) []dto.TicketAttachment {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	attachments := make([]dto.TicketAttachment, 0)
	if err := common.UnmarshalJsonStr(raw, &attachments); err != nil {
		common.SysLog("failed to decode ticket attachments: " + err.Error())
		return nil
	}
	return attachments
}

func ensureTicketAccess(ticket *model.Ticket, userID int, admin bool) error {
	if ticket == nil || ticket.ID == 0 {
		return ErrTicketNotFound
	}
	if !admin && ticket.UserID != userID {
		return ErrTicketForbidden
	}
	return nil
}

func CreateTicket(userID int, req dto.TicketCreateRequest) (*dto.TicketDetailResponse, error) {
	title := strings.TrimSpace(req.Title)
	content := strings.TrimSpace(req.Content)
	if title == "" {
		return nil, errors.New("ticket title is required")
	}
	if len(title) > 255 {
		return nil, errors.New("ticket title is too long")
	}
	if content == "" {
		return nil, errors.New("ticket content is required")
	}

	attachments, err := encodeTicketAttachments(req.Attachments)
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	ticket := &model.Ticket{
		UserID:           userID,
		Title:            title,
		Category:         normalizeTicketCategory(req.Category),
		Priority:         normalizeTicketPriority(req.Priority),
		Status:           model.TicketStatusOpen,
		RelatedType:      strings.TrimSpace(req.RelatedType),
		RelatedID:        req.RelatedID,
		LastReplyAt:      now,
		AdminUnreadCount: 1,
	}
	message := &model.TicketMessage{
		SenderID:    userID,
		SenderRole:  model.TicketSenderUser,
		Content:     content,
		Attachments: attachments,
	}

	if err := model.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(ticket).Error; err != nil {
			return err
		}
		message.TicketID = ticket.ID
		return tx.Create(message).Error
	}); err != nil {
		return nil, err
	}
	return GetTicketDetail(userID, false, ticket.ID)
}

func ListTickets(userID int, admin bool, pageInfo *common.PageInfo, status string, category string, priority string, keyword string, assignedAdminID int) (*dto.TicketListResponse, error) {
	if pageInfo == nil {
		pageInfo = &common.PageInfo{Page: 1, PageSize: common.ItemsPerPage}
	}
	status = strings.TrimSpace(strings.ToLower(status))
	if status == "all" {
		status = ""
	}
	if status != "" && !validTicketStatuses[status] {
		return nil, fmt.Errorf("invalid ticket status: %s", status)
	}
	priority = strings.TrimSpace(strings.ToLower(priority))
	if priority == "all" {
		priority = ""
	}
	if priority != "" && !validTicketPriorities[priority] {
		return nil, fmt.Errorf("invalid ticket priority: %s", priority)
	}

	tickets, total, err := model.ListTickets(model.TicketListQuery{
		UserID:          userID,
		Admin:           admin,
		Status:          status,
		Category:        normalizeOptionalFilter(category),
		Priority:        priority,
		Keyword:         keyword,
		AssignedAdminID: assignedAdminID,
		PageInfo:        pageInfo,
	})
	if err != nil {
		return nil, err
	}
	userMap := getTicketUserMap(tickets, nil)
	items := make([]dto.TicketResponse, 0, len(tickets))
	for _, ticket := range tickets {
		items = append(items, ticketToResponse(ticket, userMap))
	}
	return &dto.TicketListResponse{
		Items:    items,
		Total:    total,
		Page:     pageInfo.GetPage(),
		PageSize: pageInfo.GetPageSize(),
	}, nil
}

func normalizeOptionalFilter(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "all" {
		return ""
	}
	return value
}

func GetTicketDetail(userID int, admin bool, ticketID int) (*dto.TicketDetailResponse, error) {
	ticket, err := model.GetTicketByID(ticketID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTicketNotFound
		}
		return nil, err
	}
	if err := ensureTicketAccess(ticket, userID, admin); err != nil {
		return nil, err
	}

	if admin {
		_ = model.DB.Model(ticket).Update("admin_unread_count", 0).Error
		ticket.AdminUnreadCount = 0
	} else {
		_ = model.DB.Model(ticket).Update("user_unread_count", 0).Error
		ticket.UserUnreadCount = 0
	}

	messages, err := model.GetTicketMessages(ticketID, admin)
	if err != nil {
		return nil, err
	}
	userMap := getTicketUserMap([]model.Ticket{*ticket}, messages)
	return &dto.TicketDetailResponse{
		Ticket:   ticketToResponse(*ticket, userMap),
		Messages: messagesToResponse(messages, userMap),
	}, nil
}

func AddTicketMessage(userID int, admin bool, ticketID int, req dto.TicketMessageRequest) (*dto.TicketDetailResponse, error) {
	content := strings.TrimSpace(req.Content)
	if content == "" {
		return nil, errors.New("message content is required")
	}
	attachments, err := encodeTicketAttachments(req.Attachments)
	if err != nil {
		return nil, err
	}

	ticket, err := model.GetTicketByID(ticketID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTicketNotFound
		}
		return nil, err
	}
	if err := ensureTicketAccess(ticket, userID, admin); err != nil {
		return nil, err
	}
	if ticket.Status == model.TicketStatusClosed {
		return nil, errors.New("ticket is closed")
	}

	internal := admin && req.Internal
	senderRole := model.TicketSenderUser
	if admin {
		senderRole = model.TicketSenderAdmin
	}
	now := time.Now().Unix()
	message := &model.TicketMessage{
		TicketID:    ticketID,
		SenderID:    userID,
		SenderRole:  senderRole,
		Content:     content,
		Internal:    internal,
		Attachments: attachments,
	}

	if err := model.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(message).Error; err != nil {
			return err
		}
		updates := map[string]interface{}{
			"last_reply_at": now,
		}
		if admin {
			updates["admin_unread_count"] = 0
			if !internal {
				updates["status"] = model.TicketStatusAnswered
				updates["user_unread_count"] = gorm.Expr("user_unread_count + ?", 1)
			}
			if ticket.AssignedAdminID == 0 {
				updates["assigned_admin_id"] = userID
			}
		} else {
			updates["status"] = model.TicketStatusOpen
			updates["user_unread_count"] = 0
			updates["admin_unread_count"] = gorm.Expr("admin_unread_count + ?", 1)
		}
		return tx.Model(&model.Ticket{}).Where("id = ?", ticketID).Updates(updates).Error
	}); err != nil {
		return nil, err
	}
	return GetTicketDetail(userID, admin, ticketID)
}

func UpdateTicketStatus(adminID int, ticketID int, status string) (*dto.TicketDetailResponse, error) {
	normalized, err := normalizeTicketStatus(status)
	if err != nil {
		return nil, err
	}
	ticket, err := model.GetTicketByID(ticketID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTicketNotFound
		}
		return nil, err
	}
	updates := map[string]interface{}{
		"status": normalized,
	}
	if ticket.AssignedAdminID == 0 {
		updates["assigned_admin_id"] = adminID
	}
	if err := model.DB.Model(&model.Ticket{}).Where("id = ?", ticketID).Updates(updates).Error; err != nil {
		return nil, err
	}
	return GetTicketDetail(adminID, true, ticketID)
}

func AssignTicket(adminID int, ticketID int, assignedAdminID int) (*dto.TicketDetailResponse, error) {
	if assignedAdminID < 0 {
		return nil, errors.New("assigned admin id is invalid")
	}
	if _, err := model.GetTicketByID(ticketID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTicketNotFound
		}
		return nil, err
	}
	if err := model.DB.Model(&model.Ticket{}).Where("id = ?", ticketID).Update("assigned_admin_id", assignedAdminID).Error; err != nil {
		return nil, err
	}
	return GetTicketDetail(adminID, true, ticketID)
}

func CloseTicket(userID int, admin bool, ticketID int) (*dto.TicketDetailResponse, error) {
	ticket, err := model.GetTicketByID(ticketID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTicketNotFound
		}
		return nil, err
	}
	if err := ensureTicketAccess(ticket, userID, admin); err != nil {
		return nil, err
	}
	if err := model.DB.Model(&model.Ticket{}).Where("id = ?", ticketID).Update("status", model.TicketStatusClosed).Error; err != nil {
		return nil, err
	}
	return GetTicketDetail(userID, admin, ticketID)
}

func ReopenTicket(userID int, ticketID int) (*dto.TicketDetailResponse, error) {
	ticket, err := model.GetTicketByID(ticketID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTicketNotFound
		}
		return nil, err
	}
	if err := ensureTicketAccess(ticket, userID, false); err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	if err := model.DB.Model(&model.Ticket{}).Where("id = ?", ticketID).Updates(map[string]interface{}{
		"status":             model.TicketStatusOpen,
		"last_reply_at":      now,
		"admin_unread_count": gorm.Expr("admin_unread_count + ?", 1),
	}).Error; err != nil {
		return nil, err
	}
	return GetTicketDetail(userID, false, ticketID)
}

func getTicketUserMap(tickets []model.Ticket, messages []model.TicketMessage) map[int]model.User {
	idSet := map[int]bool{}
	for _, ticket := range tickets {
		if ticket.UserID > 0 {
			idSet[ticket.UserID] = true
		}
		if ticket.AssignedAdminID > 0 {
			idSet[ticket.AssignedAdminID] = true
		}
	}
	for _, message := range messages {
		if message.SenderID > 0 {
			idSet[message.SenderID] = true
		}
	}
	ids := make([]int, 0, len(idSet))
	for id := range idSet {
		ids = append(ids, id)
	}
	users, err := model.GetUsersByIds(ids)
	if err != nil {
		common.SysLog("failed to load ticket users: " + err.Error())
		return map[int]model.User{}
	}
	userMap := make(map[int]model.User, len(users))
	for _, user := range users {
		userMap[user.Id] = user
	}
	return userMap
}

func ticketToResponse(ticket model.Ticket, userMap map[int]model.User) dto.TicketResponse {
	response := dto.TicketResponse{
		ID:               ticket.ID,
		UserID:           ticket.UserID,
		Title:            ticket.Title,
		Category:         ticket.Category,
		Priority:         ticket.Priority,
		Status:           ticket.Status,
		AssignedAdminID:  ticket.AssignedAdminID,
		RelatedType:      ticket.RelatedType,
		RelatedID:        ticket.RelatedID,
		LastReplyAt:      ticket.LastReplyAt,
		UserUnreadCount:  ticket.UserUnreadCount,
		AdminUnreadCount: ticket.AdminUnreadCount,
		CreatedAt:        ticket.CreatedAt,
		UpdatedAt:        ticket.UpdatedAt,
	}
	if user, ok := userMap[ticket.UserID]; ok {
		response.Username = user.Username
		response.UserDisplayName = user.DisplayName
	}
	if admin, ok := userMap[ticket.AssignedAdminID]; ok {
		response.AssignedAdminName = admin.Username
	}
	return response
}

func messagesToResponse(messages []model.TicketMessage, userMap map[int]model.User) []dto.TicketMessageResponse {
	responses := make([]dto.TicketMessageResponse, 0, len(messages))
	for _, message := range messages {
		response := dto.TicketMessageResponse{
			ID:          message.ID,
			TicketID:    message.TicketID,
			SenderID:    message.SenderID,
			SenderRole:  message.SenderRole,
			Content:     message.Content,
			Internal:    message.Internal,
			Attachments: decodeTicketAttachments(message.Attachments),
			CreatedAt:   message.CreatedAt,
		}
		if user, ok := userMap[message.SenderID]; ok {
			if strings.TrimSpace(user.DisplayName) != "" {
				response.SenderName = user.DisplayName
			} else {
				response.SenderName = user.Username
			}
		}
		if response.SenderName == "" && message.SenderRole == model.TicketSenderSystem {
			response.SenderName = "System"
		}
		responses = append(responses, response)
	}
	return responses
}
