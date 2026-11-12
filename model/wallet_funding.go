package model

import (
	"errors"
	"math/big"
	"strings"
	"unicode/utf8"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

const (
	FundingPaid         = "paid"
	FundingGift         = "gift"
	FundingAdjustment   = "adjustment"
	FundingUnclassified = "unclassified"
)

// WalletFundingEntry records the nature of an inflow, not a second spendable
// balance. Relay reservations, charges and refunds must not write to this table.
type WalletFundingEntry struct {
	Id         int    `json:"id"`
	UserId     int    `json:"user_id" gorm:"not null;index:idx_wallet_funding_user_time,priority:1"`
	Amount     int64  `json:"amount,string" gorm:"type:bigint;not null"`
	Kind       string `json:"kind" gorm:"type:varchar(32);not null"`
	Source     string `json:"source" gorm:"type:varchar(32);not null"`
	Reference  string `json:"reference" gorm:"type:varchar(255)"`
	EventKey   string `json:"-" gorm:"type:varchar(128);not null;uniqueIndex"`
	OperatorId int    `json:"operator_id" gorm:"not null;default:0"`
	Note       string `json:"note" gorm:"type:text"`
	CreatedAt  int64  `json:"created_at" gorm:"not null;index:idx_wallet_funding_user_time,priority:2"`
}

type WalletFundingSummary struct {
	Paid         string `json:"paid"`
	Gift         string `json:"gift"`
	Adjustment   string `json:"adjustment"`
	Unclassified string `json:"unclassified"`
}

func NormalizeFundingKind(kind, fallback string) (string, error) {
	if kind == "" {
		kind = fallback
	}
	switch kind {
	case FundingPaid, FundingGift, FundingAdjustment, FundingUnclassified:
		return kind, nil
	default:
		return "", errors.New("invalid funding kind")
	}
}

func insertWalletFundingTx(tx *gorm.DB, entry WalletFundingEntry) (bool, error) {
	if entry.Amount == 0 {
		return false, nil
	}
	if entry.UserId <= 0 || entry.EventKey == "" || entry.Source == "" ||
		entry.Amount > 2*int64(common.MaxWalletQuota) || entry.Amount < -2*int64(common.MaxWalletQuota) {
		return false, errors.New("invalid wallet funding entry")
	}
	kind, err := NormalizeFundingKind(entry.Kind, FundingUnclassified)
	if err != nil {
		return false, err
	}
	if kind != FundingAdjustment && (entry.Amount < 0 || entry.Amount > int64(common.MaxWalletQuota)) {
		return false, errors.New("invalid non-adjustment funding amount")
	}
	entry.Kind = kind
	if entry.CreatedAt == 0 {
		entry.CreatedAt = common.GetTimestamp()
	}
	// A duplicate must fail the enclosing transaction, including its balance
	// update. Avoid database-specific no-op upsert RowsAffected semantics.
	result := tx.Create(&entry)
	return result.RowsAffected == 1, result.Error
}

func RecordWalletFundingTx(tx *gorm.DB, entry WalletFundingEntry) error {
	created, err := insertWalletFundingTx(tx, entry)
	if err == nil && !created && entry.Amount != 0 {
		return errors.New("funding event already recorded")
	}
	return err
}

// CreditUserFunding writes the balance and its business-origin record together.
// Keep IncreaseUserQuota unchanged: it is also used for settlement refunds.
func CreditUserFunding(userID, quota int, kind, source, eventKey, reference string, operatorID int, note string) error {
	if quota <= 0 || quota > common.MaxWalletQuota {
		return ErrInvalidTopUpQuota
	}
	var err error
	kind, err = NormalizeFundingKind(kind, FundingUnclassified)
	if err != nil {
		return err
	}
	credited := false
	err = DB.Transaction(func(tx *gorm.DB) error {
		var user User
		if err := lockForUpdate(tx).Select("id").First(&user, userID).Error; err != nil {
			return err
		}
		var existing WalletFundingEntry
		err := lockForUpdate(tx).Where("event_key = ?", eventKey).Take(&existing).Error
		if err == nil {
			if existing.UserId != userID || existing.Amount != int64(quota) || existing.Kind != kind || existing.Source != source {
				return errors.New("funding event does not match the original credit")
			}
			return nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		created, err := insertWalletFundingTx(tx, WalletFundingEntry{UserId: userID, Amount: int64(quota), Kind: kind, Source: source, EventKey: eventKey, Reference: reference, OperatorId: operatorID, Note: note})
		if err != nil || !created {
			return err
		}
		if err := creditTopUpQuota(tx, userID, quota, nil); err != nil {
			return err
		}
		credited = true
		return nil
	})
	if err == nil && credited {
		syncCreditUserQuotaCache(userID, quota, source)
	}
	if err != nil {
		// A concurrent retry may have committed the same event while this
		// transaction lost a uniqueness race. Treat that exact event as an
		// idempotent success after the failed transaction has rolled back.
		var existing WalletFundingEntry
		if lookupErr := DB.Where("event_key = ?", eventKey).First(&existing).Error; lookupErr == nil &&
			existing.UserId == userID && existing.Amount == int64(quota) && existing.Kind == kind && existing.Source == source {
			return nil
		}
	}
	return err
}

// AdjustUserWalletFunding preserves a single wallet. Replacements and decreases
// are adjustments; only explicit positive additions may count as paid or gifts.
func AdjustUserWalletFunding(userID, operatorID int, mode string, value int, kind, note string) (int, error) {
	if utf8.RuneCountInString(note) > 500 {
		return 0, errors.New("funding note must not exceed 500 characters")
	}
	if value < -common.MaxWalletQuota {
		return 0, errors.New("wallet adjustment exceeds safe range")
	}
	if err := common.ValidateWalletQuota(value); err != nil {
		return 0, err
	}
	kind, err := NormalizeFundingKind(kind, FundingAdjustment)
	if err != nil || kind == FundingUnclassified {
		return 0, errors.New("invalid funding kind")
	}
	if mode == "override" && value < 0 {
		return 0, errors.New("wallet override must not be negative")
	}
	var before int
	err = DB.Transaction(func(tx *gorm.DB) error {
		var user User
		if err := lockForUpdate(tx).First(&user, userID).Error; err != nil {
			return err
		}
		before = user.Quota
		if before < -common.MaxWalletQuota || before > common.MaxWalletQuota {
			return errors.New("wallet balance exceeds safe adjustment range")
		}
		var delta int64
		switch mode {
		case "add":
			if value <= 0 {
				return ErrInvalidTopUpQuota
			}
			if err := creditTopUpQuota(tx, userID, value, nil); err != nil {
				return err
			}
			delta = int64(value)
		case "subtract":
			if value <= 0 {
				return ErrInvalidTopUpQuota
			}
			result := tx.Model(&User{}).Where("id = ? AND quota >= ?", userID, -common.MaxWalletQuota+value).Update("quota", gorm.Expr("quota - ?", value))
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected != 1 {
				return errors.New("wallet adjustment exceeds safe range")
			}
			delta = -int64(value)
			kind = FundingAdjustment
		case "override":
			if err := tx.Model(&User{}).Where("id = ?", userID).Update("quota", value).Error; err != nil {
				return err
			}
			delta = int64(value) - int64(before)
			kind = FundingAdjustment
		default:
			return errors.New("invalid adjustment mode")
		}
		return RecordWalletFundingTx(tx, WalletFundingEntry{UserId: userID, Amount: delta, Kind: kind, Source: "admin", EventKey: "admin:" + common.GetUUID(), OperatorId: operatorID, Note: strings.TrimSpace(note)})
	})
	if err == nil {
		// Invalidate instead of setting a stale value over concurrent consumption.
		if cacheErr := InvalidateUserCache(userID); cacheErr != nil {
			common.SysError("failed to invalidate adjusted wallet cache: " + cacheErr.Error())
		}
	}
	return before, err
}

func GetWalletFunding(userID, offset, limit int, kind string) ([]WalletFundingEntry, int64, WalletFundingSummary, error) {
	summary := WalletFundingSummary{Paid: "0", Gift: "0", Adjustment: "0", Unclassified: "0"}
	if kind != "" {
		if _, err := NormalizeFundingKind(kind, ""); err != nil {
			return nil, 0, summary, err
		}
	}
	if offset < 0 {
		offset = 0
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	// Sum exact integers in bounded batches. A lifetime total may exceed both
	// the wallet limit and SQLite's signed SUM range, even when each credit is safe.
	totals := map[string]*big.Int{
		FundingPaid: new(big.Int), FundingGift: new(big.Int),
		FundingAdjustment: new(big.Int), FundingUnclassified: new(big.Int),
	}
	var batch []WalletFundingEntry
	err := DB.Model(&WalletFundingEntry{}).Select("id", "kind", "amount").Where("user_id = ?", userID).
		FindInBatches(&batch, 500, func(tx *gorm.DB, batchIndex int) error {
			for _, entry := range batch {
				if total, ok := totals[entry.Kind]; ok {
					total.Add(total, big.NewInt(entry.Amount))
				}
			}
			return nil
		}).Error
	if err != nil {
		return nil, 0, summary, err
	}
	summary.Paid = totals[FundingPaid].String()
	summary.Gift = totals[FundingGift].String()
	summary.Adjustment = totals[FundingAdjustment].String()
	summary.Unclassified = totals[FundingUnclassified].String()
	query := DB.Model(&WalletFundingEntry{}).Where("user_id = ?", userID)
	if kind != "" {
		query = query.Where("kind = ?", kind)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, summary, err
	}
	rows := make([]WalletFundingEntry, 0)
	err = query.Order("created_at desc, id desc").Offset(offset).Limit(limit).Find(&rows).Error
	return rows, total, summary, err
}
