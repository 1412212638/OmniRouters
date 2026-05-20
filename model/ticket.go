package model

import (
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
)

const (
	TicketStatusOpen     = "open"
	TicketStatusPending  = "pending"
	TicketStatusAnswered = "answered"
	TicketStatusClosed   = "closed"

	TicketPriorityLow    = "low"
	TicketPriorityNormal = "normal"
	TicketPriorityHigh   = "high"
	TicketPriorityUrgent = "urgent"

	TicketSenderUser   = "user"
	TicketSenderAdmin  = "admin"
	TicketSenderSystem = "system"
)

type Ticket struct {
	ID               int    `json:"id"`
	UserID           int    `json:"user_id" gorm:"index;not null"`
	Title            string `json:"title" gorm:"type:varchar(255);index;not null"`
	Category         string `json:"category" gorm:"type:varchar(64);index;default:'general'"`
	Priority         string `json:"priority" gorm:"type:varchar(32);index;default:'normal'"`
	Status           string `json:"status" gorm:"type:varchar(32);index;default:'open'"`
	AssignedAdminID  int    `json:"assigned_admin_id" gorm:"index;default:0"`
	RelatedType      string `json:"related_type" gorm:"type:varchar(64);index"`
	RelatedID        int    `json:"related_id" gorm:"index;default:0"`
	LastReplyAt      int64  `json:"last_reply_at" gorm:"index;default:0"`
	UserUnreadCount  int    `json:"user_unread_count" gorm:"default:0"`
	AdminUnreadCount int    `json:"admin_unread_count" gorm:"default:0"`
	CreatedAt        int64  `json:"created_at" gorm:"autoCreateTime;index"`
	UpdatedAt        int64  `json:"updated_at" gorm:"autoUpdateTime"`
}

type TicketMessage struct {
	ID          int    `json:"id"`
	TicketID    int    `json:"ticket_id" gorm:"index;not null"`
	SenderID    int    `json:"sender_id" gorm:"index;default:0"`
	SenderRole  string `json:"sender_role" gorm:"type:varchar(32);index;not null"`
	Content     string `json:"content" gorm:"type:text;not null"`
	Internal    bool   `json:"internal" gorm:"index;default:false"`
	Attachments string `json:"attachments" gorm:"type:text"`
	CreatedAt   int64  `json:"created_at" gorm:"autoCreateTime;index"`
}

type TicketListQuery struct {
	UserID          int
	Admin           bool
	Status          string
	Category        string
	Priority        string
	Keyword         string
	AssignedAdminID int
	PageInfo        *common.PageInfo
}

func ListTickets(query TicketListQuery) ([]Ticket, int64, error) {
	tickets := make([]Ticket, 0)
	var total int64

	db := DB.Model(&Ticket{})
	if !query.Admin {
		db = db.Where("user_id = ?", query.UserID)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.Category != "" {
		db = db.Where("category = ?", query.Category)
	}
	if query.Priority != "" {
		db = db.Where("priority = ?", query.Priority)
	}
	if query.AssignedAdminID > 0 {
		db = db.Where("assigned_admin_id = ?", query.AssignedAdminID)
	}
	if keyword := strings.TrimSpace(query.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		if id, err := strconv.Atoi(keyword); err == nil {
			db = db.Where("id = ? OR title LIKE ?", id, like)
		} else {
			db = db.Where("title LIKE ?", like)
		}
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	pageInfo := query.PageInfo
	if pageInfo == nil {
		pageInfo = &common.PageInfo{Page: 1, PageSize: common.ItemsPerPage}
	}
	err := db.Order("last_reply_at desc").Order("id desc").
		Limit(pageInfo.GetPageSize()).
		Offset(pageInfo.GetStartIdx()).
		Find(&tickets).Error
	return tickets, total, err
}

func GetTicketByID(id int) (*Ticket, error) {
	ticket := &Ticket{}
	err := DB.First(ticket, "id = ?", id).Error
	return ticket, err
}

func GetTicketMessages(ticketID int, includeInternal bool) ([]TicketMessage, error) {
	messages := make([]TicketMessage, 0)
	query := DB.Where("ticket_id = ?", ticketID)
	if !includeInternal {
		query = query.Where("internal = ?", false)
	}
	err := query.Order("id asc").Find(&messages).Error
	return messages, err
}
