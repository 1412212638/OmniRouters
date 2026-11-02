package model

import (
	"errors"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupOrganizationTest(t *testing.T) (*Organization, *OrganizationMember) {
	t.Helper()
	require.NoError(t, DB.AutoMigrate(&Organization{}, &OrganizationMember{}, &OrganizationMemberUsage{}, &OrganizationInvitation{}, &OrganizationWalletTransaction{}))
	for _, table := range []string{"organization_wallet_transactions", "organization_member_usages", "organization_invitations", "organization_members", "organizations"} {
		require.NoError(t, DB.Exec("DELETE FROM "+table).Error)
	}
	organization := &Organization{Name: "Team", OwnerUserId: 91001, Group: "default", Quota: 1000}
	require.NoError(t, DB.Create(organization).Error)
	member := &OrganizationMember{OrganizationId: organization.Id, UserId: 91001, Role: OrganizationMemberMember, Status: OrganizationMemberActive, MonthlyQuotaLimit: 100}
	require.NoError(t, DB.Create(member).Error)
	return organization, member
}

func TestOrganizationReservationAndMonthScopedRefund(t *testing.T) {
	organization, member := setupOrganizationTest(t)
	reserved, err := TryReserveOrganizationQuota(organization.Id, member.UserId, 60, "2026-11")
	require.NoError(t, err)
	require.True(t, reserved)
	reserved, err = TryReserveOrganizationQuota(organization.Id, member.UserId, 41, "2026-11")
	require.NoError(t, err)
	require.False(t, reserved)
	reserved, err = TryReserveOrganizationQuota(organization.Id, member.UserId, 80, "2026-12")
	require.NoError(t, err)
	require.True(t, reserved)
	require.NoError(t, AdjustOrganizationQuota(organization.Id, member.UserId, -20, "2026-11"))
	var november, december OrganizationMemberUsage
	require.NoError(t, DB.Where("organization_id = ? AND period = ?", organization.Id, "2026-11").First(&november).Error)
	require.NoError(t, DB.Where("organization_id = ? AND period = ?", organization.Id, "2026-12").First(&december).Error)
	require.Equal(t, 40, november.UsedQuota)
	require.Equal(t, 80, december.UsedQuota)
	stored, err := GetOrganizationById(organization.Id)
	require.NoError(t, err)
	require.Equal(t, 880, stored.Quota)
	require.Equal(t, 120, stored.UsedQuota)
	require.Zero(t, stored.RequestCount, "reservations must not count retries as requests")
}

func TestOrganizationSettlementDoesNotChargePersonalWallet(t *testing.T) {
	organization, member := setupOrganizationTest(t)
	require.NoError(t, DB.Model(organization).Update("quota", 50).Error)
	reserved, err := TryReserveOrganizationQuota(organization.Id, member.UserId, 40, "2026-11")
	require.NoError(t, err)
	require.True(t, reserved)
	require.NoError(t, AdjustOrganizationQuota(organization.Id, member.UserId, 30, "2026-11"))
	stored, err := GetOrganizationById(organization.Id)
	require.NoError(t, err)
	require.Equal(t, -20, stored.Quota)
	reserved, err = TryReserveOrganizationQuota(organization.Id, member.UserId, 1, "2026-11")
	require.NoError(t, err)
	require.False(t, reserved)
}

func TestRemovedOrganizationMemberCannotUseOrExposeOrganizationKeysAsPersonal(t *testing.T) {
	organization, member := setupOrganizationTest(t)
	token := Token{UserId: member.UserId, OrganizationId: organization.Id, Name: "org", Key: "organization-test-key", Status: common.TokenStatusEnabled, UnlimitedQuota: true, ExpiredTime: -1}
	require.NoError(t, DB.Create(&token).Error)
	t.Cleanup(func() { DB.Unscoped().Delete(&token) })
	_, err := ValidateUserToken(token.Key)
	require.NoError(t, err)
	tokens, err := GetAllUserTokens(member.UserId, 0, 10)
	require.NoError(t, err)
	require.Empty(t, tokens)
	require.NoError(t, RemoveOrganizationMembership(organization.Id, member.UserId, member.UserId))
	_, err = ValidateUserToken(token.Key)
	require.ErrorIs(t, err, ErrTokenInvalid)
	reserved, err := TryReserveOrganizationQuota(organization.Id, member.UserId, 1, "2026-11")
	require.NoError(t, err)
	require.False(t, reserved)
}

func TestOrganizationOwnershipTransferKeepsOneOwner(t *testing.T) {
	organization, member := setupOrganizationTest(t)
	require.NoError(t, DB.Model(member).Update("role", OrganizationMemberOwner).Error)
	target := OrganizationMember{OrganizationId: organization.Id, UserId: 91002, Role: OrganizationMemberMember, Status: OrganizationMemberActive}
	require.NoError(t, DB.Create(&target).Error)
	require.Error(t, TransferOrganizationOwnership(organization.Id, target.UserId, member.UserId))
	require.NoError(t, TransferOrganizationOwnership(organization.Id, member.UserId, target.UserId))
	stored, err := GetOrganizationById(organization.Id)
	require.NoError(t, err)
	require.Equal(t, target.UserId, stored.OwnerUserId)
	var count int64
	require.NoError(t, DB.Model(&OrganizationMember{}).Where("organization_id = ? AND role = ?", organization.Id, OrganizationMemberOwner).Count(&count).Error)
	require.EqualValues(t, 1, count)
	previous, err := GetOrganizationMember(organization.Id, member.UserId)
	require.NoError(t, err)
	require.Equal(t, OrganizationMemberAdmin, previous.Role)
	require.Error(t, UpdateOrganizationMember(organization.Id, member.UserId, target.UserId, OrganizationMemberMember, nil))
	require.NoError(t, UpdateOrganizationMember(organization.Id, target.UserId, member.UserId, OrganizationMemberMember, nil))
	storedOwner, err := GetOrganizationMember(organization.Id, target.UserId)
	require.NoError(t, err)
	require.Equal(t, OrganizationMemberOwner, storedOwner.Role)
}

func TestRemovedOrganizationMemberKeyCannotBeReenabled(t *testing.T) {
	organization, member := setupOrganizationTest(t)
	token := Token{UserId: member.UserId, OrganizationId: organization.Id, Name: "org", Key: "removed-member-organization-key", Status: common.TokenStatusEnabled, UnlimitedQuota: true, ExpiredTime: -1}
	require.NoError(t, DB.Create(&token).Error)
	t.Cleanup(func() { DB.Unscoped().Delete(&token) })
	require.NoError(t, RemoveOrganizationMembership(organization.Id, member.UserId, member.UserId))
	enabled := common.TokenStatusEnabled
	_, err := UpdateOrganizationToken(organization.Id, token.Id, nil, &enabled)
	require.ErrorContains(t, err, "no longer an active organization member")
	stored, err := GetOrganizationToken(organization.Id, token.Id)
	require.NoError(t, err)
	require.Equal(t, common.TokenStatusDisabled, stored.Status)
}

func TestOrganizationWalletTransferConservesBalanceAndRecordsLedger(t *testing.T) {
	organization, member := setupOrganizationTest(t)
	user := User{Id: member.UserId, Username: "organization-transfer-owner", Quota: 500}
	require.NoError(t, DB.Create(&user).Error)
	t.Cleanup(func() { DB.Unscoped().Delete(&user) })
	require.Error(t, TransferUserQuotaToOrganization(organization.Id, member.UserId, 200), "ordinary members cannot fund the organization")
	require.NoError(t, DB.Model(member).Update("role", OrganizationMemberOwner).Error)
	require.NoError(t, TransferUserQuotaToOrganization(organization.Id, member.UserId, 200))
	storedUser, err := GetUserById(member.UserId, false)
	require.NoError(t, err)
	storedOrganization, err := GetOrganizationById(organization.Id)
	require.NoError(t, err)
	require.Equal(t, 300, storedUser.Quota)
	require.Equal(t, 1200, storedOrganization.Quota)
	require.Equal(t, 1500, storedUser.Quota+storedOrganization.Quota)
	var transactions []OrganizationWalletTransaction
	require.NoError(t, DB.Where("organization_id = ?", organization.Id).Find(&transactions).Error)
	require.Len(t, transactions, 1)
	require.Equal(t, 200, transactions[0].Amount)
	require.Equal(t, 1200, transactions[0].BalanceAfter)
	require.Equal(t, member.UserId, transactions[0].UserId)

	require.Error(t, TransferUserQuotaToOrganization(organization.Id, member.UserId, 301))
	storedUser, err = GetUserById(member.UserId, false)
	require.NoError(t, err)
	storedOrganization, err = GetOrganizationById(organization.Id)
	require.NoError(t, err)
	require.Equal(t, 1500, storedUser.Quota+storedOrganization.Quota)
	var count int64
	require.NoError(t, DB.Model(&OrganizationWalletTransaction{}).Where("organization_id = ?", organization.Id).Count(&count).Error)
	require.EqualValues(t, 1, count)
}

func TestOrganizationWalletTransferRollsBackDebitCreditAndLedgerTogether(t *testing.T) {
	organization, member := setupOrganizationTest(t)
	user := User{Id: member.UserId, Username: "organization-transfer-rollback", Quota: 500}
	require.NoError(t, DB.Create(&user).Error)
	t.Cleanup(func() { DB.Unscoped().Delete(&user) })
	require.NoError(t, DB.Model(member).Update("role", OrganizationMemberOwner).Error)
	const callback = "test:organization-transfer-ledger-failure"
	require.NoError(t, DB.Callback().Create().Before("gorm:create").Register(callback, func(tx *gorm.DB) {
		if tx.Statement.Schema != nil && tx.Statement.Schema.Name == "OrganizationWalletTransaction" {
			tx.AddError(errors.New("ledger unavailable"))
		}
	}))
	t.Cleanup(func() { DB.Callback().Create().Remove(callback) })
	require.ErrorContains(t, TransferUserQuotaToOrganization(organization.Id, member.UserId, 200), "ledger unavailable")
	require.Equal(t, 500, getUserQuotaFromDB(t, user.Id), "a ledger write failure must roll back the personal debit")
	storedOrganization, err := GetOrganizationById(organization.Id)
	require.NoError(t, err)
	require.Equal(t, 1000, storedOrganization.Quota, "a ledger write failure must also roll back the organization credit")
	var count int64
	require.NoError(t, DB.Model(&OrganizationWalletTransaction{}).Where("organization_id = ?", organization.Id).Count(&count).Error)
	require.Zero(t, count)
}

func TestOrganizationWalletTransferOverflowDoesNotDebitPersonalWallet(t *testing.T) {
	organization, member := setupOrganizationTest(t)
	user := User{Id: member.UserId, Username: "organization-transfer-overflow", Quota: 500}
	require.NoError(t, DB.Create(&user).Error)
	t.Cleanup(func() { DB.Unscoped().Delete(&user) })
	require.NoError(t, DB.Model(member).Update("role", OrganizationMemberOwner).Error)
	require.NoError(t, DB.Model(organization).Update("quota", common.MaxWalletQuota-100).Error)
	require.ErrorContains(t, TransferUserQuotaToOrganization(organization.Id, member.UserId, 101), "balance limit")
	require.Equal(t, 500, getUserQuotaFromDB(t, user.Id))
	storedOrganization, err := GetOrganizationById(organization.Id)
	require.NoError(t, err)
	require.Equal(t, common.MaxWalletQuota-100, storedOrganization.Quota)
	var count int64
	require.NoError(t, DB.Model(&OrganizationWalletTransaction{}).Where("organization_id = ?", organization.Id).Count(&count).Error)
	require.Zero(t, count)
}
