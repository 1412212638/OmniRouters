package model

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ReferralFriendsSummary is the aggregate payload used by the Friends Hub.
// Registration rewards remain backed by the legacy user counters while
// first-top-up rewards come from the referral ledger.
type ReferralFriendsSummary struct {
	InviteCount                      int64
	FirstTopUpInviteCount            int64
	InviterRegistrationRewardQuota  int
	InviteeRegistrationRewardQuota   int
	RegistrationPendingQuota         int
	RegistrationTotalQuota           int
	CashbackPendingQuota             int
	CashbackTotalQuota               int
}

type ReferralFriend struct {
	User         User
	Cashback     *ReferralReward
}

// GetReferralFriendsSummary returns all Friends Hub counters for an inviter.
func GetReferralFriendsSummary(userID int) (*ReferralFriendsSummary, *User, error) {
	if userID <= 0 {
		return nil, nil, ErrReferralRewardInvalid
	}
	var user User
	if err := DB.Select("id, aff_code, aff_quota, aff_history, quota").First(&user, userID).Error; err != nil {
		return nil, nil, err
	}
	var inviteCount int64
	if err := DB.Model(&User{}).Where("inviter_id = ?", userID).Count(&inviteCount).Error; err != nil {
		return nil, nil, err
	}
	var cashbackInviteCount int64
	if err := DB.Model(&ReferralReward{}).Where("inviter_id = ? AND reward_type = ? AND status NOT IN ?", userID, ReferralRewardTypeFirstTopUpCashback, []string{ReferralRewardStatusCancelled, ReferralRewardStatusReversed, ReferralRewardStatusFailed}).Count(&cashbackInviteCount).Error; err != nil {
		return nil, nil, err
	}
	var totals struct {
		Pending int64
		Total   int64
	}
	if err := DB.Model(&ReferralReward{}).Select(
		"COALESCE(SUM(CASE WHEN status = ? THEN reward_quota ELSE 0 END), 0) AS pending, COALESCE(SUM(CASE WHEN status IN ? THEN reward_quota ELSE 0 END), 0) AS total",
		ReferralRewardStatusPending,
		[]string{ReferralRewardStatusPending, ReferralRewardStatusCredited},
	).Where("inviter_id = ? AND reward_type = ?", userID, ReferralRewardTypeFirstTopUpCashback).Scan(&totals).Error; err != nil {
		return nil, nil, err
	}
	return &ReferralFriendsSummary{
		InviteCount:                     inviteCount,
		FirstTopUpInviteCount:           cashbackInviteCount,
		InviterRegistrationRewardQuota: common.QuotaForInviter,
		InviteeRegistrationRewardQuota: common.QuotaForInvitee,
		RegistrationPendingQuota:        user.AffQuota,
		RegistrationTotalQuota:          user.AffHistoryQuota,
		CashbackPendingQuota:             int(totals.Pending),
		CashbackTotalQuota:               int(totals.Total),
	}, &user, nil
}

// GetReferralFriends returns invitees with their first-top-up reward, if any.
// The User value is selected explicitly so credentials and OAuth identifiers
// are never serialized into the Friends Hub response.
func GetReferralFriends(userID, offset, limit int) ([]ReferralFriend, int64, error) {
	if userID <= 0 || offset < 0 || limit <= 0 {
		return nil, 0, ErrReferralRewardInvalid
	}
	query := DB.Model(&User{}).Where("inviter_id = ?", userID)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var users []User
	if err := query.Select("id, username, display_name, email, created_at").Order("id desc").Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, err
	}
	if len(users) == 0 {
		return []ReferralFriend{}, total, nil
	}
	ids := make([]int, 0, len(users))
	for _, user := range users {
		ids = append(ids, user.Id)
	}
	var rewards []ReferralReward
	if err := DB.Where("inviter_id = ? AND invitee_id IN ? AND reward_type = ?", userID, ids, ReferralRewardTypeFirstTopUpCashback).Find(&rewards).Error; err != nil {
		return nil, 0, err
	}
	byInvitee := make(map[int]*ReferralReward, len(rewards))
	for i := range rewards {
		byInvitee[rewards[i].InviteeId] = &rewards[i]
	}
	items := make([]ReferralFriend, 0, len(users))
	for i := range users {
		items = append(items, ReferralFriend{User: users[i], Cashback: byInvitee[users[i].Id]})
	}
	return items, total, nil
}

func GetReferralRewards(userID, offset, limit int) ([]ReferralReward, int64, error) {
	if userID <= 0 || offset < 0 || limit <= 0 {
		return nil, 0, ErrReferralRewardInvalid
	}
	query := DB.Where("inviter_id = ? AND reward_type = ?", userID, ReferralRewardTypeFirstTopUpCashback)
	var total int64
	if err := query.Model(&ReferralReward{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rewards []ReferralReward
	if err := query.Order("id desc").Offset(offset).Limit(limit).Find(&rewards).Error; err != nil {
		return nil, 0, err
	}
	if len(rewards) > 0 {
		ids := make([]int, 0, len(rewards))
		for _, reward := range rewards {
			ids = append(ids, reward.InviteeId)
		}
		var invitees []User
		if err := DB.Select("id", "username").Where("id IN ?", ids).Find(&invitees).Error; err != nil {
			return nil, 0, err
		}
		names := make(map[int]string, len(invitees))
		for _, invitee := range invitees {
			names[invitee.Id] = invitee.Username
		}
		for i := range rewards {
			rewards[i].InviteeName = names[rewards[i].InviteeId]
		}
	}
	return rewards, total, nil
}

// TransferPendingReferralCashback atomically credits pending cashback rows.
// An empty rewardIDs slice transfers all pending rows for the inviter.
func TransferPendingReferralCashback(userID int, rewardIDs []int) (int, int, error) {
	if userID <= 0 {
		return 0, 0, ErrReferralRewardInvalid
	}
	var transferred int
	var count int
	err := DB.Transaction(func(tx *gorm.DB) error {
		var user User
		if err := lockForUpdate(tx).First(&user, userID).Error; err != nil {
			return err
		}
		query := lockForUpdate(tx).Where("inviter_id = ? AND reward_type = ? AND status = ?", userID, ReferralRewardTypeFirstTopUpCashback, ReferralRewardStatusPending)
		if len(rewardIDs) > 0 {
			query = query.Where("id IN ?", rewardIDs)
		}
		var rewards []ReferralReward
		if err := query.Order("id").Find(&rewards).Error; err != nil {
			return err
		}
		for _, reward := range rewards {
			if reward.RewardQuota <= 0 {
				continue
			}
			if transferred > common.MaxWalletQuota-reward.RewardQuota {
				return ErrWalletQuotaLimitExceeded
			}
			transferred += reward.RewardQuota
			count++
		}
		if transferred == 0 {
			return nil
		}
		result := tx.Model(&User{}).Where("id = ? AND quota <= ?", userID, common.MaxWalletQuota-transferred).Update("quota", gorm.Expr("quota + ?", transferred))
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return ErrWalletQuotaLimitExceeded
		}
		now := common.GetTimestamp()
		ids := make([]int, 0, count)
		for _, reward := range rewards {
			if reward.RewardQuota > 0 {
				ids = append(ids, reward.Id)
			}
		}
		if err := tx.Model(&ReferralReward{}).Where("id IN ? AND status = ?", ids, ReferralRewardStatusPending).Updates(map[string]interface{}{"status": ReferralRewardStatusCredited, "credited_at": now}).Error; err != nil {
			return err
		}
		return nil
	})
	if err == nil && transferred > 0 {
		syncCreditUserQuotaCache(userID, transferred, "referral cashback transfer")
	}
	return transferred, count, err
}

// Referral reward types identify the business event that created a ledger
// row. The type is part of the idempotency key so future reward programs can
// coexist for the same invitee without sharing a row.
const (
	ReferralRewardTypeFirstTopUpCashback = "first_topup_cashback"
)

// Referral reward states are deliberately persisted rather than inferred from
// the user's wallet balance. This lets later settlement and reversal flows
// remain auditable and retryable.
const (
	ReferralRewardStatusPending   = "pending"
	ReferralRewardStatusCredited  = "credited"
	ReferralRewardStatusNoReward  = "no_reward"
	ReferralRewardStatusReversed  = "reversed"
	ReferralRewardStatusCancelled = "cancelled"
	ReferralRewardStatusFailed    = "failed"
)

var (
	ErrReferralRewardInvalid       = errors.New("referral reward is invalid")
	ErrReferralRewardNotFound      = errors.New("referral reward not found")
	ErrReferralRewardTypeInvalid   = errors.New("referral reward type is invalid")
	ErrReferralRewardStatusInvalid = errors.New("referral reward status is invalid")
)

// ReferralReward is the durable ledger for invite-related rewards. Amounts
// are stored in wallet quota units so settlement does not depend on the
// payment provider's currency or floating-point representation. BaseMoney is
// retained as the provider-reported monetary amount for display/audit only.
//
// SourceTopUpID is nullable because future reward types may not originate from
// a top-up. The composite unique index still prevents duplicate source orders
// for top-up rewards on SQLite, MySQL, and PostgreSQL while allowing multiple
// rows whose source is NULL.
type ReferralReward struct {
	Id            int     `json:"id" gorm:"primaryKey"`
	InviterId     int     `json:"inviter_id" gorm:"not null;index:idx_referral_rewards_inviter_created,priority:1"`
	InviteeId     int     `json:"invitee_id" gorm:"not null;index;uniqueIndex:ux_referral_rewards_invitee_type,priority:1"`
	RewardType    string  `json:"reward_type" gorm:"type:varchar(64);not null;uniqueIndex:ux_referral_rewards_invitee_type,priority:2;uniqueIndex:ux_referral_rewards_source_type,priority:2"`
	Status        string  `json:"status" gorm:"type:varchar(32);not null;index;default:'pending'"`
	SourceTopUpId *int    `json:"source_topup_id,omitempty" gorm:"index;uniqueIndex:ux_referral_rewards_source_type,priority:1"`
	SourceTradeNo string  `json:"source_trade_no,omitempty" gorm:"type:varchar(255);index"`
	BaseQuota     int     `json:"base_quota" gorm:"not null;default:0"`
	BaseMoney     float64 `json:"base_money" gorm:"type:decimal(20,8);not null;default:0"`
	RewardRateBps int     `json:"reward_rate_bps" gorm:"not null;default:0"`
	RewardQuota   int     `json:"reward_quota" gorm:"not null;default:0"`
	FailureReason string  `json:"failure_reason,omitempty" gorm:"type:text"`
	CreatedAt     int64   `json:"created_at" gorm:"not null;index:idx_referral_rewards_inviter_created,priority:2"`
	CreditedAt    *int64  `json:"credited_at,omitempty"`
	ReversedAt    *int64  `json:"reversed_at,omitempty"`
	InviteeName   string  `json:"invitee_name,omitempty" gorm:"-"`
}

func (ReferralReward) TableName() string { return "referral_rewards" }

func validReferralRewardType(rewardType string) bool {
	return strings.TrimSpace(rewardType) == ReferralRewardTypeFirstTopUpCashback
}

func validReferralRewardStatus(status string) bool {
	switch strings.TrimSpace(status) {
	case ReferralRewardStatusPending,
		ReferralRewardStatusCredited,
		ReferralRewardStatusNoReward,
		ReferralRewardStatusReversed,
		ReferralRewardStatusCancelled,
		ReferralRewardStatusFailed:
		return true
	default:
		return false
	}
}

func validateReferralReward(reward *ReferralReward) error {
	if reward == nil || reward.InviterId <= 0 || reward.InviteeId <= 0 || reward.InviterId == reward.InviteeId {
		return ErrReferralRewardInvalid
	}
	if !validReferralRewardType(reward.RewardType) {
		return ErrReferralRewardTypeInvalid
	}
	if reward.Status == "" {
		reward.Status = ReferralRewardStatusPending
	}
	if !validReferralRewardStatus(reward.Status) {
		return ErrReferralRewardStatusInvalid
	}
	if reward.BaseQuota < 0 || reward.RewardQuota < 0 || reward.BaseMoney < 0 || reward.RewardRateBps < 0 || reward.RewardRateBps > 10000 {
		return ErrReferralRewardInvalid
	}
	if reward.SourceTopUpId != nil && *reward.SourceTopUpId <= 0 {
		return ErrReferralRewardInvalid
	}
	return nil
}

func referralCashbackRateBps(rate float64) (int, error) {
	if math.IsNaN(rate) || math.IsInf(rate, 0) || rate < 0 || rate > 100 {
		return 0, fmt.Errorf("invalid referral first top-up cashback rate: %v", rate)
	}
	bps := decimal.NewFromFloat(rate).Mul(decimal.NewFromInt(100)).Round(0).IntPart()
	if bps < 0 || bps > 10000 {
		return 0, fmt.Errorf("invalid referral first top-up cashback rate basis points: %d", bps)
	}
	return int(bps), nil
}

func referralCashbackQuota(creditedQuota int, rateBps int) (int, error) {
	if creditedQuota <= 0 || rateBps < 0 || rateBps > 10000 {
		return 0, ErrReferralRewardInvalid
	}
	return common.WalletQuotaFromDecimalStrict(
		decimal.NewFromInt(int64(creditedQuota)).Mul(decimal.NewFromInt(int64(rateBps))).Div(decimal.NewFromInt(10000)),
	)
}

// ApplyFirstTopUpReferralCashbackTx records the first successful wallet top-up
// reward for an invitee. It must be called inside the same transaction that
// changes the top-up from pending to success and credits the invitee wallet.
// The reward remains pending until the user explicitly transfers it to the
// wallet, so registration rewards and cashback never share AffQuota.
func ApplyFirstTopUpReferralCashbackTx(tx *gorm.DB, topUp *TopUp, creditedQuota int) error {
	if tx == nil || topUp == nil || topUp.Id <= 0 || topUp.UserId <= 0 || creditedQuota <= 0 {
		return ErrReferralRewardInvalid
	}

	var invitee User
	if err := lockForUpdate(tx).First(&invitee, topUp.UserId).Error; err != nil {
		return err
	}
	if invitee.InviterId <= 0 || invitee.InviterId == invitee.Id {
		return nil
	}
	var previousWalletTopUps int64
	if err := tx.Model(&TopUp{}).
		Where("user_id = ? AND status = ? AND amount > ? AND id <> ?", invitee.Id, common.TopUpStatusSuccess, 0, topUp.Id).
		Count(&previousWalletTopUps).Error; err != nil {
		return err
	}
	if previousWalletTopUps > 0 {
		return nil
	}
	var inviter User
	if err := tx.First(&inviter, invitee.InviterId).Error; err != nil {
		// A deleted or missing inviter cannot receive a reward. The top-up itself
		// remains valid and should not fail because an old referral is stale.
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	_ = inviter

	rate := common.ReferralFirstTopUpCashbackRate
	rateBps, rateErr := referralCashbackRateBps(rate)
	if rateErr != nil {
		common.SysError(fmt.Sprintf("invalid referral first top-up cashback rate: %v; treating as disabled", rate))
		rateBps = 0
	}
	rewardQuota, err := referralCashbackQuota(creditedQuota, rateBps)
	if err != nil {
		common.SysError(fmt.Sprintf("invalid referral cashback quota for topup %d: %v; treating as zero", topUp.Id, err))
		rewardQuota = 0
	}
	sourceID := topUp.Id
	status := ReferralRewardStatusPending
	failureReason := ""
	if rewardQuota == 0 {
		status = ReferralRewardStatusNoReward
		failureReason = "cashback disabled or below one quota unit"
	}
	reward := &ReferralReward{
		InviterId:     inviter.Id,
		InviteeId:     invitee.Id,
		RewardType:    ReferralRewardTypeFirstTopUpCashback,
		Status:        status,
		SourceTopUpId: &sourceID,
		SourceTradeNo: topUp.TradeNo,
		BaseQuota:     creditedQuota,
		BaseMoney:     topUp.Money,
		RewardRateBps: rateBps,
		RewardQuota:   rewardQuota,
		FailureReason: failureReason,
	}
	_, _, err = CreateReferralRewardIfAbsent(tx, reward)
	return err
}

func referralRewardDB(db *gorm.DB) *gorm.DB {
	if db == nil {
		return DB
	}
	return db
}

// GetReferralRewardByInviteeType returns the idempotency row for an invitee
// and reward type. Callers can use gorm.ErrRecordNotFound to distinguish a
// not-yet-created reward from a database failure.
func GetReferralRewardByInviteeType(db *gorm.DB, inviteeId int, rewardType string) (*ReferralReward, error) {
	if inviteeId <= 0 || !validReferralRewardType(rewardType) {
		return nil, ErrReferralRewardInvalid
	}
	var reward ReferralReward
	if err := referralRewardDB(db).Where("invitee_id = ? AND reward_type = ?", inviteeId, rewardType).First(&reward).Error; err != nil {
		return nil, err
	}
	return &reward, nil
}

// GetReferralRewardBySourceTopUp returns the reward generated from a top-up.
func GetReferralRewardBySourceTopUp(db *gorm.DB, sourceTopUpId int, rewardType string) (*ReferralReward, error) {
	if sourceTopUpId <= 0 || !validReferralRewardType(rewardType) {
		return nil, ErrReferralRewardInvalid
	}
	var reward ReferralReward
	if err := referralRewardDB(db).Where("source_top_up_id = ? AND reward_type = ?", sourceTopUpId, rewardType).First(&reward).Error; err != nil {
		return nil, err
	}
	return &reward, nil
}

// CreateReferralRewardIfAbsent atomically inserts a reward ledger row and
// returns the existing row when another callback won the unique constraint.
// It accepts an existing transaction so the eventual wallet credit can be
// committed atomically with the ledger row.
func CreateReferralRewardIfAbsent(db *gorm.DB, reward *ReferralReward) (created bool, existing *ReferralReward, err error) {
	db = referralRewardDB(db)
	if err := validateReferralReward(reward); err != nil {
		return false, nil, err
	}
	if reward.CreatedAt == 0 {
		reward.CreatedAt = common.GetTimestamp()
	}

	result := db.Clauses(clause.OnConflict{DoNothing: true}).Create(reward)
	if result.Error != nil {
		return false, nil, result.Error
	}
	if result.RowsAffected == 1 {
		return true, reward, nil
	}

	// The unique key can be won by either the invitee/type or source/type
	// index. Check both so callers receive the canonical row in either case.
	if current, findErr := GetReferralRewardByInviteeType(db, reward.InviteeId, reward.RewardType); findErr == nil {
		return false, current, nil
	}
	if reward.SourceTopUpId != nil {
		if current, findErr := GetReferralRewardBySourceTopUp(db, *reward.SourceTopUpId, reward.RewardType); findErr == nil {
			return false, current, nil
		}
	}
	return false, nil, ErrReferralRewardNotFound
}
