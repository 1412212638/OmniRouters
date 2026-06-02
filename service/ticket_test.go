package service

import (
	"errors"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func seedTicketUser(t *testing.T, id int, username string, role int) {
	t.Helper()
	user := &model.User{
		Id:       id,
		Username: username,
		AffCode:  username + "_aff",
		Role:     role,
		Status:   common.UserStatusEnabled,
	}
	require.NoError(t, model.DB.Create(user).Error)
}

func TestTicketLifecycle(t *testing.T) {
	truncate(t)
	seedTicketUser(t, 1, "ticket_user", common.RoleCommonUser)
	seedTicketUser(t, 2, "ticket_admin", common.RoleAdminUser)

	created, err := CreateTicket(1, dto.TicketCreateRequest{
		Title:    "Channel error",
		Category: "channel",
		Priority: "high",
		Content:  "The upstream channel returns 500.",
	})
	require.NoError(t, err)
	require.Equal(t, model.TicketStatusOpen, created.Ticket.Status)
	assert.Equal(t, 1, created.Ticket.AdminUnreadCount)
	assert.Len(t, created.Messages, 1)

	adminList, err := ListTickets(2, true, nil, "", "", "", "", 0)
	require.NoError(t, err)
	require.Len(t, adminList.Items, 1)
	assert.Equal(t, int64(1), adminList.Total)
	assert.Equal(t, 1, adminList.Items[0].AdminUnreadCount)

	adminDetail, err := GetTicketDetail(2, true, created.Ticket.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, adminDetail.Ticket.AdminUnreadCount)

	replied, err := AddTicketMessage(2, true, created.Ticket.ID, dto.TicketMessageRequest{
		Content: "Please retry now.",
	})
	require.NoError(t, err)
	assert.Equal(t, model.TicketStatusAnswered, replied.Ticket.Status)
	assert.Equal(t, 1, replied.Ticket.UserUnreadCount)
	assert.Equal(t, 2, replied.Ticket.AssignedAdminID)
	assert.Greater(t, replied.Ticket.LastAdminReplyAt, int64(0))
	assert.Len(t, replied.Messages, 2)

	userDetail, err := GetTicketDetail(1, false, created.Ticket.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, userDetail.Ticket.UserUnreadCount)
}

func TestTicketAccessIsScopedToOwner(t *testing.T) {
	truncate(t)
	seedTicketUser(t, 1, "owner_user", common.RoleCommonUser)
	seedTicketUser(t, 3, "other_user", common.RoleCommonUser)

	created, err := CreateTicket(1, dto.TicketCreateRequest{
		Title:   "Quota question",
		Content: "Why did my quota change?",
	})
	require.NoError(t, err)

	_, err = GetTicketDetail(3, false, created.Ticket.ID)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrTicketForbidden))
}

func TestTicketCreateOpenLimit(t *testing.T) {
	truncate(t)
	seedTicketUser(t, 1, "limited_user", common.RoleCommonUser)

	oldLimit := common.TicketMaxOpenPerUser
	common.TicketMaxOpenPerUser = 1
	t.Cleanup(func() {
		common.TicketMaxOpenPerUser = oldLimit
	})

	_, err := CreateTicket(1, dto.TicketCreateRequest{
		Title:   "First issue",
		Content: "Please help.",
	})
	require.NoError(t, err)

	_, err = CreateTicket(1, dto.TicketCreateRequest{
		Title:   "Second issue",
		Content: "Please help again.",
	})
	require.Error(t, err)
	var ticketErr *TicketI18nError
	require.True(t, errors.As(err, &ticketErr))
	assert.Equal(t, i18n.MsgTicketOpenLimitReached, ticketErr.Key)
}

func TestTicketAdminCloseReopenWindow(t *testing.T) {
	truncate(t)
	seedTicketUser(t, 1, "window_user", common.RoleCommonUser)
	seedTicketUser(t, 2, "window_admin", common.RoleAdminUser)

	oldWindow := common.TicketReopenWindowHours
	common.TicketReopenWindowHours = 24
	t.Cleanup(func() {
		common.TicketReopenWindowHours = oldWindow
	})

	created, err := CreateTicket(1, dto.TicketCreateRequest{
		Title:   "Close window",
		Content: "Please close later.",
	})
	require.NoError(t, err)

	closed, err := CloseTicket(2, true, created.Ticket.ID)
	require.NoError(t, err)
	require.Equal(t, model.TicketStatusClosed, closed.Ticket.Status)
	assert.Equal(t, model.TicketSenderAdmin, closed.Ticket.ClosedByRole)
	assert.Greater(t, closed.Ticket.ReopenUntil, closed.Ticket.ClosedAt)

	require.NoError(t, model.DB.Model(&model.Ticket{}).
		Where("id = ?", created.Ticket.ID).
		Update("reopen_until", time.Now().Unix()-1).Error)

	_, err = ReopenTicket(1, created.Ticket.ID)
	require.Error(t, err)
	var ticketErr *TicketI18nError
	require.True(t, errors.As(err, &ticketErr))
	assert.Equal(t, i18n.MsgTicketReopenWindowExpired, ticketErr.Key)

	require.NoError(t, model.DB.Model(&model.Ticket{}).
		Where("id = ?", created.Ticket.ID).
		Update("reopen_until", time.Now().Unix()+3600).Error)

	reopened, err := ReopenTicket(1, created.Ticket.ID)
	require.NoError(t, err)
	assert.Equal(t, model.TicketStatusOpen, reopened.Ticket.Status)
	assert.Equal(t, int64(0), reopened.Ticket.ReopenUntil)
	assert.Equal(t, "", reopened.Ticket.ClosedByRole)
}

func TestCloseIdleAnsweredTickets(t *testing.T) {
	truncate(t)
	seedTicketUser(t, 1, "idle_user", common.RoleCommonUser)
	seedTicketUser(t, 2, "idle_admin", common.RoleAdminUser)

	created, err := CreateTicket(1, dto.TicketCreateRequest{
		Title:   "Idle ticket",
		Content: "Waiting for support.",
	})
	require.NoError(t, err)
	_, err = AddTicketMessage(2, true, created.Ticket.ID, dto.TicketMessageRequest{
		Content: "Please confirm.",
	})
	require.NoError(t, err)

	now := time.Now().Unix()
	oldReplyAt := now - 25*3600
	require.NoError(t, model.DB.Model(&model.Ticket{}).
		Where("id = ?", created.Ticket.ID).
		Updates(map[string]interface{}{
			"last_reply_at":       oldReplyAt,
			"last_admin_reply_at": oldReplyAt,
		}).Error)

	closedCount, err := model.CloseIdleAnsweredTickets(now-24*3600, now, now+24*3600, 100)
	require.NoError(t, err)
	assert.Equal(t, 1, closedCount)

	ticket, err := model.GetTicketByID(created.Ticket.ID)
	require.NoError(t, err)
	assert.Equal(t, model.TicketStatusClosed, ticket.Status)
	assert.Equal(t, model.TicketSenderSystem, ticket.ClosedByRole)
	assert.Equal(t, model.TicketCloseReasonUserIdleTimeout, ticket.CloseReason)
	assert.Greater(t, ticket.ReopenUntil, now)
}
