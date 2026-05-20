package service

import (
	"errors"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
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
