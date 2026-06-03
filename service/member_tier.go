package service

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/ratio_setting"

	"github.com/bytedance/gopkg/util/gopool"
)

const (
	memberTierBatchSize = 500
)

type MemberTierProgress struct {
	Rule             setting.MemberTierRule `json:"rule"`
	Qualified        bool                   `json:"qualified"`
	Current          bool                   `json:"current"`
	Next             bool                   `json:"next"`
	TopupRemaining   int64                  `json:"topup_remaining"`
	UsedRemaining    int64                  `json:"used_remaining"`
	GroupRatio       float64                `json:"group_ratio"`
	TopupGroupRatio  float64                `json:"topup_group_ratio"`
}

type MemberTierEvaluation struct {
	Enabled                    bool                       `json:"enabled"`
	CurrentGroup               string                     `json:"current_group"`
	TargetGroup                string                     `json:"target_group"`
	TotalTopupQuota            int64                      `json:"total_topup_quota"`
	UsedQuota                  int64                      `json:"used_quota"`
	CurrentTier                *setting.MemberTierRule    `json:"current_tier,omitempty"`
	TargetTier                 *setting.MemberTierRule    `json:"target_tier,omitempty"`
	NextTier                   *setting.MemberTierRule    `json:"next_tier,omitempty"`
	Progress                   []MemberTierProgress       `json:"progress"`
	Upgraded                   bool                       `json:"upgraded"`
	SkippedByActiveSubscription bool                       `json:"skipped_by_active_subscription"`
}

type MemberTierRecalculateResult struct {
	Scanned                    int `json:"scanned"`
	Upgraded                   int `json:"upgraded"`
	SkippedByActiveSubscription int `json:"skipped_by_active_subscription"`
}

var (
	memberTierDailyOnce    sync.Once
	memberTierDailyRunning atomic.Bool
)

func EvaluateMemberTierForUser(userId int, applyUpgrade bool) (*MemberTierEvaluation, error) {
	user, err := model.GetUserById(userId, false)
	if err != nil {
		return nil, err
	}
	totalTopup, err := model.GetUserSuccessfulTopUpQuotaTotal(userId)
	if err != nil {
		return nil, err
	}
	return evaluateMemberTierForUser(user, totalTopup, applyUpgrade)
}

func CheckMemberTierUpgradeAfterTopUp(userId int) {
	if userId <= 0 {
		return
	}
	gopool.Go(func() {
		evaluation, err := EvaluateMemberTierForUser(userId, true)
		if err != nil {
			logger.LogWarn(context.Background(), fmt.Sprintf("member tier top-up check failed: user_id=%d error=%v", userId, err))
			return
		}
		if evaluation.Upgraded {
			logger.LogInfo(context.Background(), fmt.Sprintf("member tier upgraded after top-up: user_id=%d group=%s", userId, evaluation.TargetGroup))
		}
	})
}

func RecalculateAllMemberTiers() (MemberTierRecalculateResult, error) {
	result := MemberTierRecalculateResult{}
	if len(enabledMemberTierRules()) == 0 {
		return result, nil
	}

	topupTotals, err := model.GetSuccessfulTopUpQuotaTotals()
	if err != nil {
		return result, err
	}

	afterId := 0
	for {
		users, err := model.ListUsersForMemberTierEvaluation(afterId, memberTierBatchSize)
		if err != nil {
			return result, err
		}
		if len(users) == 0 {
			break
		}

		for _, user := range users {
			if user == nil {
				continue
			}
			afterId = user.Id
			evaluation, err := evaluateMemberTierForUser(user, topupTotals[user.Id], true)
			if err != nil {
				return result, err
			}
			result.Scanned++
			if evaluation.Upgraded {
				result.Upgraded++
			}
			if evaluation.SkippedByActiveSubscription {
				result.SkippedByActiveSubscription++
			}
		}

		if len(users) < memberTierBatchSize {
			break
		}
	}
	return result, nil
}

func StartMemberTierDailyTask() {
	memberTierDailyOnce.Do(func() {
		if !common.IsMasterNode {
			return
		}
		gopool.Go(func() {
			logger.LogInfo(context.Background(), "member tier daily task started: schedule=local midnight")
			for {
				now := time.Now()
				next := nextLocalMidnight(now)
				timer := time.NewTimer(time.Until(next))
				<-timer.C
				runMemberTierDailyOnce()
			}
		})
	})
}

func runMemberTierDailyOnce() {
	if !memberTierDailyRunning.CompareAndSwap(false, true) {
		return
	}
	defer memberTierDailyRunning.Store(false)

	result, err := RecalculateAllMemberTiers()
	if err != nil {
		logger.LogWarn(context.Background(), fmt.Sprintf("member tier daily task failed: %v", err))
		return
	}
	if result.Scanned > 0 || result.Upgraded > 0 || result.SkippedByActiveSubscription > 0 {
		logger.LogInfo(context.Background(), fmt.Sprintf(
			"member tier daily task finished: scanned=%d upgraded=%d skipped_by_active_subscription=%d",
			result.Scanned,
			result.Upgraded,
			result.SkippedByActiveSubscription,
		))
	}
}

func nextLocalMidnight(now time.Time) time.Time {
	return time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
}

func enabledMemberTierRules() []setting.MemberTierRule {
	rules := setting.GetMemberTierRulesCopy()
	enabled := make([]setting.MemberTierRule, 0, len(rules))
	for _, rule := range rules {
		if rule.Enabled {
			enabled = append(enabled, rule)
		}
	}
	return enabled
}

func evaluateMemberTierForUser(user *model.User, totalTopup int64, applyUpgrade bool) (*MemberTierEvaluation, error) {
	if user == nil {
		return nil, fmt.Errorf("user is nil")
	}

	rules := enabledMemberTierRules()
	usedQuota := int64(user.UsedQuota)
	evaluation := &MemberTierEvaluation{
		Enabled:         len(rules) > 0,
		CurrentGroup:    user.Group,
		TotalTopupQuota: totalTopup,
		UsedQuota:       usedQuota,
		Progress:        buildMemberTierProgress(rules, user.Group, totalTopup, usedQuota),
	}
	if len(rules) == 0 {
		return evaluation, nil
	}

	currentIndex := -1
	matchedIndex := -1
	for idx, rule := range rules {
		if rule.Group == user.Group {
			currentIndex = idx
		}
		if totalTopup >= rule.MinTopupQuota && usedQuota >= rule.MinUsedQuota {
			matchedIndex = idx
		}
	}

	if currentIndex >= 0 {
		currentTier := rules[currentIndex]
		evaluation.CurrentTier = &currentTier
	}
	if matchedIndex >= 0 {
		targetTier := rules[matchedIndex]
		evaluation.TargetTier = &targetTier
		evaluation.TargetGroup = targetTier.Group
	}

	nextIndex := findNextMemberTierIndex(rules, matchedIndex, totalTopup, usedQuota)
	if nextIndex >= 0 {
		nextTier := rules[nextIndex]
		evaluation.NextTier = &nextTier
	}
	evaluation.Progress = markNextMemberTierProgress(evaluation.Progress, nextIndex)

	if matchedIndex < 0 {
		return evaluation, nil
	}
	if currentIndex >= matchedIndex {
		return evaluation, nil
	}

	targetRule := rules[matchedIndex]
	if user.Group == targetRule.Group {
		return evaluation, nil
	}

	activeSubscriptionGroup, err := model.HasActiveUpgradeSubscriptionGroup(user.Id, user.Group)
	if err != nil {
		return evaluation, err
	}
	if activeSubscriptionGroup {
		evaluation.SkippedByActiveSubscription = true
		return evaluation, nil
	}

	if !applyUpgrade {
		return evaluation, nil
	}

	if err := model.UpdateUserGroupById(user.Id, targetRule.Group); err != nil {
		return evaluation, err
	}
	evaluation.CurrentGroup = targetRule.Group
	evaluation.CurrentTier = &targetRule
	evaluation.Progress = markNextMemberTierProgress(
		buildMemberTierProgress(rules, targetRule.Group, totalTopup, usedQuota),
		nextIndex,
	)
	evaluation.Upgraded = true
	return evaluation, nil
}

func buildMemberTierProgress(rules []setting.MemberTierRule, currentGroup string, totalTopup int64, usedQuota int64) []MemberTierProgress {
	progress := make([]MemberTierProgress, 0, len(rules))
	for _, rule := range rules {
		topupRatio := common.GetTopupGroupRatio(rule.Group)
		if topupRatio == 0 {
			topupRatio = 1
		}
		item := MemberTierProgress{
			Rule:            rule,
			Qualified:       totalTopup >= rule.MinTopupQuota && usedQuota >= rule.MinUsedQuota,
			Current:         rule.Group == currentGroup,
			TopupRemaining:  positiveDiff(rule.MinTopupQuota, totalTopup),
			UsedRemaining:   positiveDiff(rule.MinUsedQuota, usedQuota),
			GroupRatio:      ratio_setting.GetGroupRatio(rule.Group),
			TopupGroupRatio: topupRatio,
		}
		progress = append(progress, item)
	}
	return progress
}

func markNextMemberTierProgress(progress []MemberTierProgress, nextIndex int) []MemberTierProgress {
	if nextIndex < 0 || nextIndex >= len(progress) {
		return progress
	}
	progress[nextIndex].Next = true
	return progress
}

func findNextMemberTierIndex(rules []setting.MemberTierRule, matchedIndex int, totalTopup int64, usedQuota int64) int {
	for idx := matchedIndex + 1; idx < len(rules); idx++ {
		rule := rules[idx]
		if totalTopup < rule.MinTopupQuota || usedQuota < rule.MinUsedQuota {
			return idx
		}
	}
	return -1
}

func positiveDiff(target int64, value int64) int64 {
	if target <= value {
		return 0
	}
	return target - value
}
