package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"gorm.io/gorm"
)

type organizationInvitationReminderHandler struct{}

func init() { RegisterSystemTaskHandler(organizationInvitationReminderHandler{}) }

func (organizationInvitationReminderHandler) Type() string {
	return model.SystemTaskTypeOrganizationInvitationReminder
}
func (organizationInvitationReminderHandler) Enabled() bool {
	return common.OrganizationEnabled && organizationEmailConfigured()
}
func (organizationInvitationReminderHandler) Interval() time.Duration { return time.Hour }
func (organizationInvitationReminderHandler) NewPayload() any         { return nil }
func (organizationInvitationReminderHandler) Run(ctx context.Context, task *model.SystemTask, runnerID string) {
	sent, err := runOrganizationInvitationReminders(ctx, common.GetTimestamp(), common.SendEmail)
	status, message := model.SystemTaskStatusSucceeded, ""
	if err != nil {
		status, message = model.SystemTaskStatusFailed, err.Error()
	}
	if finishErr := model.FinishSystemTask(task.TaskID, runnerID, status, map[string]any{"sent": sent}, message); finishErr != nil {
		common.SysLog(fmt.Sprintf("organization invitation reminder task failed to persist result: %v", finishErr))
	}
}

// The scheduled task is protected by the existing cross-node DB lease. A
// per-invitation claim and sent marker also survive restarts and failed runs.
func runOrganizationInvitationReminders(ctx context.Context, now int64, send func(string, string, string) error) (int, error) {
	if !common.OrganizationEnabled || !organizationEmailConfigured() {
		return 0, nil
	}
	invitations, err := model.GetOrganizationInvitationsDueForReminder(ctx, now, 1000)
	if err != nil {
		return 0, err
	}
	sent := 0
	var firstError error
	for _, invitation := range invitations {
		if err := ctx.Err(); err != nil {
			return sent, err
		}
		claimed, err := model.ClaimOrganizationInvitationReminder(ctx, invitation.Id, now)
		if err != nil {
			return sent, err
		}
		if !claimed {
			continue
		}
		organization, organizationErr := model.GetOrganizationById(invitation.OrganizationId)
		inviter, inviterErr := model.GetOrganizationMember(invitation.OrganizationId, invitation.InvitedBy)
		var lookupErr error
		if organizationErr != nil && !errors.Is(organizationErr, gorm.ErrRecordNotFound) {
			lookupErr = organizationErr
		} else if inviterErr != nil && !errors.Is(inviterErr, gorm.ErrRecordNotFound) {
			lookupErr = inviterErr
		}
		if lookupErr != nil {
			// A temporary database failure says nothing about invitation
			// validity. Release the claim so the next sweep can retry.
			finishErr := model.FinishOrganizationInvitationReminder(ctx, invitation.Id, now, false)
			if firstError == nil {
				firstError = errors.Join(lookupErr, finishErr)
			}
			continue
		}
		if organizationErr != nil || inviterErr != nil || inviter.Status != model.OrganizationMemberActive || !model.OrganizationCanManage(inviter.Role) || (invitation.Role == model.OrganizationMemberAdmin && inviter.Role != model.OrganizationMemberOwner) {
			// The invitation cannot currently be accepted; do not advertise it.
			if finishErr := model.FinishOrganizationInvitationReminder(ctx, invitation.Id, now, true); finishErr != nil && firstError == nil {
				firstError = finishErr
			}
			continue
		}
		err = sendOrganizationNotificationMail(organizationNotificationMail{Kind: OrganizationNotificationInvitationExpiring, OrganizationName: organization.Name, Recipient: invitation.Email, Role: invitation.Role, ExpiresAt: invitation.ExpiresAt, WorkspaceURL: organizationWorkspaceURL()}, send)
		if finishErr := model.FinishOrganizationInvitationReminder(ctx, invitation.Id, now, err == nil); finishErr != nil && firstError == nil {
			firstError = finishErr
		}
		if err != nil {
			if firstError == nil {
				firstError = err
			}
			continue
		}
		sent++
	}
	return sent, firstError
}
