package service

import (
	"context"
	"errors"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func withOrganizationEmailSettings(t *testing.T, configured bool) {
	t.Helper()
	common.OptionMapRWMutex.Lock()
	previous := common.OptionMap
	common.OptionMap = map[string]string{"EmailProvider": "resend"}
	if configured {
		common.OptionMap["ResendApiKey"] = "test-key"
		common.OptionMap["ResendFrom"] = "service@example.com"
	}
	common.OptionMapRWMutex.Unlock()
	previousLanguage, previousSystemName := common.EmailLanguage, common.SystemName
	t.Cleanup(func() {
		common.OptionMapRWMutex.Lock()
		common.OptionMap = previous
		common.OptionMapRWMutex.Unlock()
		common.EmailLanguage, common.SystemName = previousLanguage, previousSystemName
	})
}

func TestOrganizationNotificationMailUsesConfiguredSenderAndEscapesNames(t *testing.T) {
	withOrganizationEmailSettings(t, true)
	common.EmailLanguage = "en"
	common.SystemName = "Gateway\r\nInjected"
	called := 0
	err := sendOrganizationNotificationMail(organizationNotificationMail{
		Kind: OrganizationNotificationInvitation, OrganizationName: `<img src=x onerror=alert(1)>`,
		Recipient: "member@example.com", Role: model.OrganizationMemberMember, ExpiresAt: 1794000000,
		WorkspaceURL: "https://example.com/organizations?a=1&b=2",
	}, func(subject, recipient, content string) error {
		called++
		assert.Equal(t, "member@example.com", recipient)
		assert.NotContains(t, subject, "\r")
		assert.NotContains(t, subject, "\n")
		assert.Contains(t, content, "&lt;img src=x onerror=alert(1)&gt;")
		assert.NotContains(t, content, "<img")
		assert.Contains(t, content, "a=1&amp;b=2")
		assert.Contains(t, content, "UTC")
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, 1, called)
}

func TestOrganizationNotificationMailSkipsUnconfiguredEmail(t *testing.T) {
	withOrganizationEmailSettings(t, false)
	require.NoError(t, sendOrganizationNotificationMail(organizationNotificationMail{
		Kind: OrganizationNotificationRemoved, Recipient: "member@example.com",
	}, func(string, string, string) error {
		t.Fatal("unconfigured email must not send")
		return nil
	}))
}

func TestOrganizationNotificationMailPreservesDeliveryFailure(t *testing.T) {
	withOrganizationEmailSettings(t, true)
	want := errors.New("mail unavailable")
	err := sendOrganizationNotificationMail(organizationNotificationMail{
		Kind: OrganizationNotificationRemoved, Recipient: "member@example.com",
	}, func(string, string, string) error { return want })
	require.ErrorIs(t, err, want)
}

func TestOrganizationNotificationMailSupportsEveryEventInBothLanguages(t *testing.T) {
	withOrganizationEmailSettings(t, true)
	for _, language := range []string{"en", "zh"} {
		common.EmailLanguage = language
		for _, kind := range []string{OrganizationNotificationInvitation, OrganizationNotificationInvitationExpiring, OrganizationNotificationRoleChanged, OrganizationNotificationRemoved, OrganizationNotificationOwnershipTransferred} {
			subject, content := buildOrganizationNotificationMail(organizationNotificationMail{
				Kind: kind, OrganizationName: "Team <A>", Role: model.OrganizationMemberAdmin,
			})
			assert.NotEmpty(t, subject)
			assert.Contains(t, content, "Team &lt;A&gt;")
			if language == "en" {
				assert.Contains(t, subject, "Organization")
			} else {
				assert.Contains(t, subject, "组织")
			}
		}
	}
}

func TestOrganizationInvitationReminderIsSentOnceAndOnlyBeforeExpiry(t *testing.T) {
	withOrganizationEmailSettings(t, true)
	previousEnabled := common.OrganizationEnabled
	common.OrganizationEnabled = true
	t.Cleanup(func() { common.OrganizationEnabled = previousEnabled })
	require.NoError(t, model.DB.AutoMigrate(&model.Organization{}, &model.OrganizationMember{}, &model.OrganizationInvitation{}))
	const organizationID = 9820
	require.NoError(t, model.DB.Create(&model.Organization{Id: organizationID, Name: "Reminder team", OwnerUserId: 9820}).Error)
	require.NoError(t, model.DB.Create(&model.OrganizationMember{OrganizationId: organizationID, UserId: 9820, Role: model.OrganizationMemberOwner, Status: model.OrganizationMemberActive}).Error)
	t.Cleanup(func() {
		model.DB.Where("organization_id = ?", organizationID).Delete(&model.OrganizationInvitation{})
		model.DB.Where("organization_id = ?", organizationID).Delete(&model.OrganizationMember{})
		model.DB.Delete(&model.Organization{}, organizationID)
	})
	const now int64 = 1794000000
	for _, invitation := range []model.OrganizationInvitation{
		{OrganizationId: organizationID, Email: "due@example.com", InvitedBy: 9820, Role: model.OrganizationMemberMember, Status: model.OrganizationInvitationPending, ExpiresAt: now + 3600},
		{OrganizationId: organizationID, Email: "later@example.com", InvitedBy: 9820, Role: model.OrganizationMemberMember, Status: model.OrganizationInvitationPending, ExpiresAt: now + 86401},
		{OrganizationId: organizationID, Email: "expired@example.com", InvitedBy: 9820, Role: model.OrganizationMemberMember, Status: model.OrganizationInvitationPending, ExpiresAt: now - 1},
		{OrganizationId: organizationID, Email: "accepted@example.com", InvitedBy: 9820, Role: model.OrganizationMemberMember, Status: model.OrganizationInvitationAccepted, ExpiresAt: now + 3600},
		{OrganizationId: organizationID, Email: "revoked@example.com", InvitedBy: 9820, Role: model.OrganizationMemberMember, Status: model.OrganizationInvitationRevoked, ExpiresAt: now + 3600},
	} {
		require.NoError(t, model.DB.Create(&invitation).Error)
	}
	calls := 0
	send := func(_ string, recipient string, _ string) error {
		calls++
		assert.Equal(t, "due@example.com", recipient)
		return nil
	}
	sent, err := runOrganizationInvitationReminders(context.Background(), now, send)
	require.NoError(t, err)
	assert.Equal(t, 1, sent)
	sent, err = runOrganizationInvitationReminders(context.Background(), now, send)
	require.NoError(t, err)
	assert.Zero(t, sent)
	assert.Equal(t, 1, calls)
}

func TestOrganizationInvitationReminderReleasesClaimAfterDeliveryFailure(t *testing.T) {
	withOrganizationEmailSettings(t, true)
	previousEnabled := common.OrganizationEnabled
	common.OrganizationEnabled = true
	t.Cleanup(func() { common.OrganizationEnabled = previousEnabled })
	require.NoError(t, model.DB.AutoMigrate(&model.Organization{}, &model.OrganizationMember{}, &model.OrganizationInvitation{}))
	const organizationID = 9821
	const now int64 = 1794000000
	require.NoError(t, model.DB.Create(&model.Organization{Id: organizationID, Name: "Retry team", OwnerUserId: 9821}).Error)
	require.NoError(t, model.DB.Create(&model.OrganizationMember{OrganizationId: organizationID, UserId: 9821, Role: model.OrganizationMemberOwner, Status: model.OrganizationMemberActive}).Error)
	invitation := model.OrganizationInvitation{OrganizationId: organizationID, Email: "retry@example.com", InvitedBy: 9821, Role: model.OrganizationMemberMember, Status: model.OrganizationInvitationPending, ExpiresAt: now + 3600}
	require.NoError(t, model.DB.Create(&invitation).Error)
	t.Cleanup(func() {
		model.DB.Where("organization_id = ?", organizationID).Delete(&model.OrganizationInvitation{})
		model.DB.Where("organization_id = ?", organizationID).Delete(&model.OrganizationMember{})
		model.DB.Delete(&model.Organization{}, organizationID)
	})
	_, err := runOrganizationInvitationReminders(context.Background(), now, func(string, string, string) error { return errors.New("temporary delivery failure") })
	require.ErrorContains(t, err, "temporary delivery failure")
	sent, err := runOrganizationInvitationReminders(context.Background(), now, func(string, string, string) error { return nil })
	require.NoError(t, err)
	assert.Equal(t, 1, sent)
}

func TestOrganizationInvitationReminderRetriesTemporaryLookupFailure(t *testing.T) {
	for index, table := range []string{"organizations", "organization_members"} {
		t.Run(table, func(t *testing.T) {
			withOrganizationEmailSettings(t, true)
			previousEnabled := common.OrganizationEnabled
			common.OrganizationEnabled = true
			t.Cleanup(func() { common.OrganizationEnabled = previousEnabled })
			require.NoError(t, model.DB.AutoMigrate(&model.Organization{}, &model.OrganizationMember{}, &model.OrganizationInvitation{}))
			organizationID := 9822 + index
			const now int64 = 1794000000
			require.NoError(t, model.DB.Create(&model.Organization{Id: organizationID, Name: "Database retry team", OwnerUserId: organizationID}).Error)
			require.NoError(t, model.DB.Create(&model.OrganizationMember{OrganizationId: organizationID, UserId: organizationID, Role: model.OrganizationMemberOwner, Status: model.OrganizationMemberActive}).Error)
			invitation := model.OrganizationInvitation{OrganizationId: organizationID, Email: "database-retry@example.com", InvitedBy: organizationID, Role: model.OrganizationMemberMember, Status: model.OrganizationInvitationPending, ExpiresAt: now + 3600}
			require.NoError(t, model.DB.Create(&invitation).Error)
			t.Cleanup(func() {
				model.DB.Where("organization_id = ?", organizationID).Delete(&model.OrganizationInvitation{})
				model.DB.Where("organization_id = ?", organizationID).Delete(&model.OrganizationMember{})
				model.DB.Delete(&model.Organization{}, organizationID)
			})

			lookupFailure := errors.New("temporary organization database lookup failure")
			injected := false
			callbackName := "test_organization_reminder_lookup_failure_" + table
			require.NoError(t, model.DB.Callback().Query().Before("gorm:query").Register(callbackName, func(tx *gorm.DB) {
				if tx.Statement.Table == table && !injected {
					injected = true
					tx.AddError(lookupFailure)
				}
			}))
			t.Cleanup(func() { _ = model.DB.Callback().Query().Remove(callbackName) })
			calls := 0
			send := func(string, string, string) error { calls++; return nil }
			sent, err := runOrganizationInvitationReminders(context.Background(), now, send)
			require.ErrorIs(t, err, lookupFailure)
			assert.True(t, injected)
			assert.Zero(t, sent)
			assert.Zero(t, calls)
			var saved model.OrganizationInvitation
			require.NoError(t, model.DB.First(&saved, invitation.Id).Error)
			assert.Zero(t, saved.ReminderSentAt)
			assert.Zero(t, saved.ReminderClaimedAt)

			sent, err = runOrganizationInvitationReminders(context.Background(), now, send)
			require.NoError(t, err)
			assert.Equal(t, 1, sent)
			assert.Equal(t, 1, calls)
			require.NoError(t, model.DB.First(&saved, invitation.Id).Error)
			assert.Equal(t, now, saved.ReminderSentAt)
			assert.Zero(t, saved.ReminderClaimedAt)
		})
	}
}
