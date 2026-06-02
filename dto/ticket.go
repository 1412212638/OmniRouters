package dto

type TicketAttachment struct {
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
	Size int64  `json:"size,omitempty"`
	Type string `json:"type,omitempty"`
}

type TicketCreateRequest struct {
	Title       string             `json:"title"`
	Category    string             `json:"category"`
	Priority    string             `json:"priority"`
	Content     string             `json:"content"`
	RelatedType string             `json:"related_type,omitempty"`
	RelatedID   int                `json:"related_id,omitempty"`
	Attachments []TicketAttachment `json:"attachments,omitempty"`
}

type TicketMessageRequest struct {
	Content     string             `json:"content"`
	Internal    bool               `json:"internal,omitempty"`
	Attachments []TicketAttachment `json:"attachments,omitempty"`
}

type TicketStatusRequest struct {
	Status string `json:"status"`
}

type TicketAssignRequest struct {
	AssignedAdminID int `json:"assigned_admin_id"`
}

type TicketResponse struct {
	ID                int    `json:"id"`
	UserID            int    `json:"user_id"`
	Username          string `json:"username,omitempty"`
	UserDisplayName   string `json:"user_display_name,omitempty"`
	Title             string `json:"title"`
	Category          string `json:"category"`
	Priority          string `json:"priority"`
	Status            string `json:"status"`
	ClosedAt          int64  `json:"closed_at"`
	ClosedByID        int    `json:"closed_by_id"`
	ClosedByRole      string `json:"closed_by_role,omitempty"`
	CloseReason       string `json:"close_reason,omitempty"`
	ReopenUntil       int64  `json:"reopen_until"`
	AssignedAdminID   int    `json:"assigned_admin_id"`
	AssignedAdminName string `json:"assigned_admin_name,omitempty"`
	RelatedType       string `json:"related_type,omitempty"`
	RelatedID         int    `json:"related_id,omitempty"`
	LastReplyAt       int64  `json:"last_reply_at"`
	LastUserReplyAt   int64  `json:"last_user_reply_at"`
	LastAdminReplyAt  int64  `json:"last_admin_reply_at"`
	UserUnreadCount   int    `json:"user_unread_count"`
	AdminUnreadCount  int    `json:"admin_unread_count"`
	CreatedAt         int64  `json:"created_at"`
	UpdatedAt         int64  `json:"updated_at"`
}

type TicketMessageResponse struct {
	ID          int                `json:"id"`
	TicketID    int                `json:"ticket_id"`
	SenderID    int                `json:"sender_id"`
	SenderRole  string             `json:"sender_role"`
	SenderName  string             `json:"sender_name,omitempty"`
	Content     string             `json:"content"`
	Internal    bool               `json:"internal"`
	Attachments []TicketAttachment `json:"attachments,omitempty"`
	CreatedAt   int64              `json:"created_at"`
}

type TicketListResponse struct {
	Items    []TicketResponse `json:"items"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}

type TicketDetailResponse struct {
	Ticket   TicketResponse          `json:"ticket"`
	Messages []TicketMessageResponse `json:"messages"`
}
