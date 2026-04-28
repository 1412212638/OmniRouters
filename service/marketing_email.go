package service

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
)

const MaxMarketingEmailRecipients = 200

type MarketingEmailUserResult struct {
	UserId   int    `json:"user_id"`
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty"`
	Reason   string `json:"reason"`
}

type MarketingEmailResult struct {
	Total        int                        `json:"total"`
	Sent         int                        `json:"sent"`
	Skipped      int                        `json:"skipped"`
	Failed       int                        `json:"failed"`
	SkippedUsers []MarketingEmailUserResult `json:"skipped_users,omitempty"`
	FailedUsers  []MarketingEmailUserResult `json:"failed_users,omitempty"`
}

func SendMarketingEmailToUsers(userIds []int, subjectTemplate string, contentTemplate string) (*MarketingEmailResult, error) {
	ids := uniquePositiveUserIds(userIds)
	if len(ids) == 0 {
		return nil, fmt.Errorf("请选择至少一个收件人")
	}
	if len(ids) > MaxMarketingEmailRecipients {
		return nil, fmt.Errorf("单次最多发送 %d 个用户", MaxMarketingEmailRecipients)
	}

	users, err := model.GetUsersByIds(ids)
	if err != nil {
		return nil, err
	}

	result := &MarketingEmailResult{Total: len(ids)}
	userById := make(map[int]model.User, len(users))
	for _, user := range users {
		userById[user.Id] = user
	}

	for _, userId := range ids {
		user, ok := userById[userId]
		if !ok {
			result.Skipped++
			result.SkippedUsers = append(result.SkippedUsers, MarketingEmailUserResult{
				UserId: userId,
				Reason: "user not found",
			})
			continue
		}

		userSetting := user.GetSetting()
		emailToUse := strings.TrimSpace(userSetting.NotificationEmail)
		if emailToUse == "" {
			emailToUse = strings.TrimSpace(user.Email)
		}
		if emailToUse == "" {
			result.Skipped++
			result.SkippedUsers = append(result.SkippedUsers, MarketingEmailUserResult{
				UserId:   user.Id,
				Username: user.Username,
				Reason:   "user has no email",
			})
			continue
		}

		subject, content := common.BuildMarketingMail(common.MarketingMailParams{
			SystemName:      common.SystemName,
			Username:        user.Username,
			DisplayName:     user.DisplayName,
			Email:           emailToUse,
			UserId:          strconv.Itoa(user.Id),
			SubjectTemplate: subjectTemplate,
			ContentTemplate: contentTemplate,
		})

		if err := common.SendEmail(subject, emailToUse, content); err != nil {
			result.Failed++
			result.FailedUsers = append(result.FailedUsers, MarketingEmailUserResult{
				UserId:   user.Id,
				Username: user.Username,
				Email:    common.MaskEmail(emailToUse),
				Reason:   err.Error(),
			})
			continue
		}
		result.Sent++
	}

	common.SysLog(fmt.Sprintf("marketing email sent: total=%d sent=%d skipped=%d failed=%d", result.Total, result.Sent, result.Skipped, result.Failed))
	return result, nil
}

func uniquePositiveUserIds(userIds []int) []int {
	seen := make(map[int]struct{}, len(userIds))
	ids := make([]int, 0, len(userIds))
	for _, id := range userIds {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	return ids
}
