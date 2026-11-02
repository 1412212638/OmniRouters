package model

import (
	"context"
	"errors"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type organizationTopUpProviderTest struct {
	provider          string
	method            string
	creditedQuota     int
	duplicateRejected bool
	complete          func(string) error
}

func organizationTopUpProviders() []organizationTopUpProviderTest {
	return []organizationTopUpProviderTest{
		{PaymentProviderEpay, "alipay", 30, false, func(tradeNo string) error {
			_, err := RechargeEpay(tradeNo, "alipay", "198.51.100.1")
			return err
		}},
		{PaymentProviderStripe, PaymentMethodStripe, 40, true, func(tradeNo string) error {
			return Recharge(tradeNo, "cus_organization_test", "198.51.100.1")
		}},
		{PaymentProviderCreem, PaymentMethodCreem, 3, true, func(tradeNo string) error {
			return RechargeCreem(tradeNo, "checkout@example.com", "Organization payer", "198.51.100.1")
		}},
		{PaymentProviderWaffo, PaymentMethodWaffo, 30, false, func(tradeNo string) error {
			return RechargeWaffo(tradeNo, "198.51.100.1")
		}},
		{PaymentProviderWaffoPancake, PaymentMethodWaffoPancake, 30, false, func(tradeNo string) error {
			_, err := RechargeWaffoPancakeWithResult(tradeNo, "198.51.100.1")
			return err
		}},
	}
}

func setupOrganizationTopUpTest(t *testing.T, provider organizationTopUpProviderTest) (*Organization, *OrganizationMember, *TopUp) {
	t.Helper()
	truncateTables(t)
	organization, member := setupOrganizationTest(t)
	require.NoError(t, DB.Model(member).Update("role", OrganizationMemberOwner).Error)
	user := User{Id: member.UserId, Username: "organization-topup-payer", Status: common.UserStatusEnabled, Quota: 75}
	require.NoError(t, DB.Create(&user).Error)
	oldQuotaPerUnit, oldEnabled := common.QuotaPerUnit, common.OrganizationEnabled
	common.QuotaPerUnit, common.OrganizationEnabled = 10, true
	t.Cleanup(func() { common.QuotaPerUnit, common.OrganizationEnabled = oldQuotaPerUnit, oldEnabled })
	order := &TopUp{
		UserId: user.Id, OrganizationId: organization.Id, Amount: 3, Money: 4,
		TradeNo: "organization-topup-" + provider.provider, PaymentMethod: provider.method,
		PaymentProvider: provider.provider, Status: common.TopUpStatusPending,
	}
	require.NoError(t, order.Insert())
	return organization, member, order
}

// Any Redis command here would be a personal-cache mutation or lookup. The
// hook fails locally before network I/O so this test needs no Redis service.
type organizationTopUpCacheGuard struct{ calls int }

func (h *organizationTopUpCacheGuard) BeforeProcess(ctx context.Context, _ redis.Cmder) (context.Context, error) {
	h.calls++
	return ctx, errors.New("organization top-up must not access personal Redis cache")
}
func (h *organizationTopUpCacheGuard) AfterProcess(context.Context, redis.Cmder) error {
	return nil
}
func (h *organizationTopUpCacheGuard) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	h.calls += len(cmds)
	return ctx, errors.New("organization top-up must not access personal Redis cache")
}
func (h *organizationTopUpCacheGuard) AfterProcessPipeline(context.Context, []redis.Cmder) error {
	return nil
}

func TestOrganizationTopUpCallbacksCreditOnlyOriginalWalletOnce(t *testing.T) {
	for _, provider := range organizationTopUpProviders() {
		t.Run(provider.provider, func(t *testing.T) {
			organization, member, order := setupOrganizationTopUpTest(t, provider)
			// Permission and feature changes after checkout cannot redirect paid
			// funds into the payer's personal wallet or strand a valid payment.
			require.NoError(t, DB.Model(member).Update("status", OrganizationMemberRemoved).Error)
			common.OrganizationEnabled = false
			require.NoError(t, provider.complete(order.TradeNo))
			err := provider.complete(order.TradeNo)
			if provider.duplicateRejected {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			stored, err := GetOrganizationById(organization.Id)
			require.NoError(t, err)
			require.Equal(t, 1000+provider.creditedQuota, stored.Quota)
			require.Equal(t, 75, getUserQuotaFromDB(t, member.UserId))
			var payer User
			require.NoError(t, DB.First(&payer, member.UserId).Error)
			require.Empty(t, payer.Email, "organization checkout must not overwrite personal email")
			require.Empty(t, payer.StripeCustomer, "organization checkout must not overwrite personal payment identity")
			require.Equal(t, common.TopUpStatusSuccess, GetTopUpByTradeNo(order.TradeNo).Status)
			var ledger []OrganizationWalletTransaction
			require.NoError(t, DB.Where("organization_id = ?", organization.Id).Find(&ledger).Error)
			require.Len(t, ledger, 1)
			require.Equal(t, provider.creditedQuota, ledger[0].Amount)
			require.Equal(t, stored.Quota, ledger[0].BalanceAfter)
			require.Equal(t, member.UserId, ledger[0].UserId)
			require.Equal(t, "topup", ledger[0].Type)
			require.Equal(t, order.TradeNo, ledger[0].Description)
			var logs []Log
			require.NoError(t, LOG_DB.Where("user_id = ? AND type = ?", member.UserId, LogTypeTopup).Find(&logs).Error)
			require.Len(t, logs, 1, "duplicate callback must not create another top-up audit log")
			require.Equal(t, organization.Id, logs[0].OrganizationId)
		})
	}
}

func TestOrganizationTopUpCacheSyncNeverCreditsPersonalBalance(t *testing.T) {
	oldRedisEnabled, oldRDB := common.RedisEnabled, common.RDB
	guard := &organizationTopUpCacheGuard{}
	client := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"})
	client.AddHook(guard)
	common.RedisEnabled, common.RDB = true, client
	t.Cleanup(func() {
		common.RedisEnabled, common.RDB = oldRedisEnabled, oldRDB
		client.Close()
	})
	syncTopUpWalletCache(&TopUp{UserId: 91001, OrganizationId: 123}, 40, "organization callback")
	require.Zero(t, guard.calls, "organization callback dispatch must not mutate or hydrate the payer's cached quota")
}

func TestOrganizationTopUpCallbackFailureRollsBackOrderWalletAndLedger(t *testing.T) {
	for _, provider := range organizationTopUpProviders() {
		for _, failure := range []string{"wallet ceiling", "ledger failure"} {
			t.Run(provider.provider+"/"+failure, func(t *testing.T) {
				organization, member, order := setupOrganizationTopUpTest(t, provider)
				beforeQuota := 1000
				if failure == "wallet ceiling" {
					beforeQuota = common.MaxWalletQuota - provider.creditedQuota + 1
					require.NoError(t, DB.Model(organization).Update("quota", beforeQuota).Error)
				} else {
					const callback = "test:organization-topup-ledger-failure"
					require.NoError(t, DB.Callback().Create().Before("gorm:create").Register(callback, func(tx *gorm.DB) {
						if tx.Statement.Schema != nil && tx.Statement.Schema.Name == "OrganizationWalletTransaction" {
							tx.AddError(errors.New("ledger unavailable"))
						}
					}))
					t.Cleanup(func() { DB.Callback().Create().Remove(callback) })
				}
				require.Error(t, provider.complete(order.TradeNo))
				stored, err := GetOrganizationById(organization.Id)
				require.NoError(t, err)
				require.Equal(t, beforeQuota, stored.Quota)
				require.Equal(t, 75, getUserQuotaFromDB(t, member.UserId))
				require.Equal(t, common.TopUpStatusPending, GetTopUpByTradeNo(order.TradeNo).Status)
				var count int64
				require.NoError(t, DB.Model(&OrganizationWalletTransaction{}).Where("organization_id = ?", organization.Id).Count(&count).Error)
				require.Zero(t, count)
			})
		}
	}
}

func TestOrganizationTopUpCapacityRequiresEnabledActiveManager(t *testing.T) {
	organization, member, _ := setupOrganizationTopUpTest(t, organizationTopUpProviders()[0])
	require.NoError(t, ValidateOrganizationTopUp(organization.Id, member.UserId, 1))
	for _, role := range []string{OrganizationMemberOwner, OrganizationMemberAdmin} {
		require.NoError(t, DB.Model(member).Update("role", role).Error)
		require.NoError(t, ValidateOrganizationTopUp(organization.Id, member.UserId, 1))
	}
	require.NoError(t, DB.Model(member).Update("role", OrganizationMemberMember).Error)
	require.ErrorContains(t, ValidateOrganizationTopUp(organization.Id, member.UserId, 1), "manager permission")
	require.NoError(t, DB.Model(member).Updates(map[string]interface{}{"role": OrganizationMemberOwner, "status": OrganizationMemberRemoved}).Error)
	require.ErrorContains(t, ValidateOrganizationTopUp(organization.Id, member.UserId, 1), "manager permission")
	require.NoError(t, DB.Model(member).Update("status", OrganizationMemberActive).Error)
	for _, quota := range []int{0, -1, common.MaxWalletQuota + 1} {
		require.ErrorIs(t, ValidateOrganizationTopUp(organization.Id, member.UserId, quota), ErrInvalidTopUpQuota)
	}
	require.ErrorIs(t, ValidateOrganizationTopUp(organization.Id, member.UserId, common.MaxWalletQuota), ErrTopUpQuotaLimitExceeded)
	common.OrganizationEnabled = false
	require.ErrorContains(t, ValidateOrganizationTopUp(organization.Id, member.UserId, 1), "disabled")
}

func TestOrganizationTopUpsExcludedFromPersonalLifetimeTotals(t *testing.T) {
	_, member, order := setupOrganizationTopUpTest(t, organizationTopUpProviders()[0])
	order.Status = common.TopUpStatusSuccess
	require.NoError(t, order.Update())
	personal := &TopUp{UserId: member.UserId, Amount: 2, PaymentMethod: "alipay", PaymentProvider: PaymentProviderEpay, TradeNo: "personal-lifetime-total", Status: common.TopUpStatusSuccess}
	require.NoError(t, personal.Insert())
	total, err := GetUserSuccessfulTopUpQuotaTotal(member.UserId)
	require.NoError(t, err)
	require.EqualValues(t, 20, total)
	totals, err := GetSuccessfulTopUpQuotaTotals()
	require.NoError(t, err)
	require.EqualValues(t, 20, totals[member.UserId])
}

func TestManualOrganizationCreemTopUpUsesQuotaUnitsAndIsIdempotent(t *testing.T) {
	provider := organizationTopUpProviders()[2]
	organization, member, order := setupOrganizationTopUpTest(t, provider)
	require.NoError(t, ManualCompleteTopUp(order.TradeNo, "198.51.100.1"))
	require.NoError(t, ManualCompleteTopUp(order.TradeNo, "198.51.100.1"))
	stored, err := GetOrganizationById(organization.Id)
	require.NoError(t, err)
	require.Equal(t, 1003, stored.Quota, "Creem amount is already quota and must not be multiplied twice")
	require.Equal(t, 75, getUserQuotaFromDB(t, member.UserId))
	var logs []Log
	require.NoError(t, LOG_DB.Where("type = ?", LogTypeTopup).Find(&logs).Error)
	require.Len(t, logs, 1, "repeat manual completion must not write a spurious user-zero log")
	require.Equal(t, member.UserId, logs[0].UserId)
	require.Equal(t, organization.Id, logs[0].OrganizationId)
}

func TestOrganizationTopUpDoesNotConsumePersonalFirstTopUpReferral(t *testing.T) {
	provider := organizationTopUpProviders()[0]
	_, member, order := setupOrganizationTopUpTest(t, provider)
	resetReferralRewardRows(t)
	inviter := User{Id: member.UserId + 1, Username: "organization-topup-inviter", Status: common.UserStatusEnabled}
	require.NoError(t, DB.Create(&inviter).Error)
	require.NoError(t, DB.Model(&User{}).Where("id = ?", member.UserId).Update("inviter_id", inviter.Id).Error)
	oldRate := common.ReferralFirstTopUpCashbackRate
	common.ReferralFirstTopUpCashbackRate = 10
	t.Cleanup(func() { common.ReferralFirstTopUpCashbackRate = oldRate })
	require.NoError(t, provider.complete(order.TradeNo))
	var count int64
	require.NoError(t, DB.Model(&ReferralReward{}).Where("invitee_id = ?", member.UserId).Count(&count).Error)
	require.Zero(t, count, "organization payment must not claim the payer's personal referral reward")
	personal := &TopUp{UserId: member.UserId, Amount: 2, PaymentMethod: provider.method, PaymentProvider: provider.provider, TradeNo: "personal-first-topup-after-organization", Status: common.TopUpStatusPending}
	require.NoError(t, personal.Insert())
	require.NoError(t, provider.complete(personal.TradeNo))
	reward, err := GetReferralRewardByInviteeType(nil, member.UserId, ReferralRewardTypeFirstTopUpCashback)
	require.NoError(t, err)
	require.Equal(t, personal.Id, *reward.SourceTopUpId)
	require.Equal(t, 2, reward.RewardQuota)
}
