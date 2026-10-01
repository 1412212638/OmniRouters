package controller

import (
	"net/http"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

type referralFriendResponse struct {
	ID                  int    `json:"id"`
	Username            string `json:"username,omitempty"`
	DisplayName         string `json:"display_name,omitempty"`
	Email               string `json:"email,omitempty"`
	CreatedAt           int64  `json:"created_at"`
	FirstTopUpStatus    string `json:"first_top_up_status"`
	FirstTopUpBaseQuota int    `json:"first_top_up_base_quota"`
	RegistrationRewardQuota int `json:"registration_reward_quota"`
	CashbackQuota       int    `json:"cashback_quota"`
	CashbackRateBps     int    `json:"cashback_rate_bps"`
	CashbackStatus      string `json:"cashback_status,omitempty"`
}

func buildReferralFriendResponse(friend model.ReferralFriend) referralFriendResponse {
	username := maskReferralName(friend.User.Username)
	displayName := maskReferralName(friend.User.DisplayName)
	response := referralFriendResponse{
		ID:          friend.User.Id,
		Username:    username,
		DisplayName: displayName,
		Email:       common.MaskEmail(friend.User.Email),
		CreatedAt:   friend.User.CreatedAt,
		FirstTopUpStatus: "pending",
		RegistrationRewardQuota: common.QuotaForInviter,
	}
	if friend.Cashback != nil {
		response.FirstTopUpStatus = "completed"
		response.FirstTopUpBaseQuota = friend.Cashback.BaseQuota
		response.CashbackQuota = friend.Cashback.RewardQuota
		response.CashbackRateBps = friend.Cashback.RewardRateBps
		response.CashbackStatus = friend.Cashback.Status
	}
	return response
}

func maskReferralName(value string) string {
	runes := []rune(value)
	if len(runes) == 0 {
		return ""
	}
	if len(runes) <= 2 {
		return "***"
	}
	return string(runes[:1]) + "***" + string(runes[len(runes)-1:])
}

// GetFriendsSummary returns the registration and first-top-up reward counters.
func GetFriendsSummary(c *gin.Context) {
	summary, user, err := model.GetReferralFriendsSummary(c.GetInt("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{
		"aff_code":                    user.AffCode,
		"invite_code":                 user.AffCode,
		"invite_count":                summary.InviteCount,
		"first_top_up_invite_count":   summary.FirstTopUpInviteCount,
		"first_topup_count":            summary.FirstTopUpInviteCount,
		"inviter_registration_reward_quota": summary.InviterRegistrationRewardQuota,
		"invitee_registration_reward_quota": summary.InviteeRegistrationRewardQuota,
		"registration_pending_quota": summary.RegistrationPendingQuota,
		"registration_total_quota":   summary.RegistrationTotalQuota,
		"cashback_pending_quota":     summary.CashbackPendingQuota,
		"cashback_total_quota":       summary.CashbackTotalQuota,
		"wallet_quota":               user.Quota,
		"cashback_rate":              common.ReferralFirstTopUpCashbackRate,
		"referral_rate":              common.ReferralFirstTopUpCashbackRate,
	})
}

// GetFriends lists invitees without exposing credentials or OAuth identifiers.
func GetFriends(c *gin.Context) {
	page := common.GetPageQuery(c)
	items, total, err := model.GetReferralFriends(c.GetInt("id"), page.GetStartIdx(), page.GetPageSize())
	if err != nil {
		common.ApiError(c, err)
		return
	}
	responses := make([]referralFriendResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, buildReferralFriendResponse(item))
	}
	page.SetTotal(int(total))
	page.SetItems(responses)
	common.ApiSuccess(c, page)
}

type referralRewardResponse struct {
	ID             int    `json:"id"`
	RewardType     string `json:"reward_type"`
	InviteeID      int    `json:"invitee_id"`
	InviteeName    string `json:"invitee_name,omitempty"`
	SourceTopUpID  *int   `json:"source_topup_id,omitempty"`
	SourceTradeNo  string `json:"source_trade_no,omitempty"`
	BaseQuota      int    `json:"base_quota"`
	BaseMoney      float64 `json:"base_money"`
	RewardRateBps  int    `json:"reward_rate_bps"`
	RateBps        int    `json:"rate_bps"`
	RewardQuota    int    `json:"reward_quota"`
	Status         string `json:"status"`
	CreatedAt      int64  `json:"created_at"`
	CreditedAt     *int64 `json:"credited_at,omitempty"`
	TransferredAt  *int64 `json:"transferred_at,omitempty"`
}

func buildReferralRewardResponse(reward model.ReferralReward) referralRewardResponse {
	return referralRewardResponse{
		ID: reward.Id, RewardType: reward.RewardType, InviteeID: reward.InviteeId, InviteeName: maskReferralName(reward.InviteeName), SourceTopUpID: reward.SourceTopUpId,
		SourceTradeNo: reward.SourceTradeNo, BaseQuota: reward.BaseQuota, BaseMoney: reward.BaseMoney,
		RewardRateBps: reward.RewardRateBps, RateBps: reward.RewardRateBps, RewardQuota: reward.RewardQuota, Status: reward.Status,
		CreatedAt: reward.CreatedAt, CreditedAt: reward.CreditedAt, TransferredAt: reward.CreditedAt,
	}
}

func GetFriendsRewards(c *gin.Context) {
	page := common.GetPageQuery(c)
	items, total, err := model.GetReferralRewards(c.GetInt("id"), page.GetStartIdx(), page.GetPageSize())
	if err != nil {
		common.ApiError(c, err)
		return
	}
	responses := make([]referralRewardResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, buildReferralRewardResponse(item))
	}
	page.SetTotal(int(total))
	page.SetItems(responses)
	common.ApiSuccess(c, page)
}

type transferReferralCashbackRequest struct {
	RewardIDs []int `json:"reward_ids"`
}

func TransferReferralCashback(c *gin.Context) {
	request := transferReferralCashbackRequest{}
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&request); err != nil {
			common.ApiError(c, err)
			return
		}
	}
	quota, count, err := model.TransferPendingReferralCashback(c.GetInt("id"), request.RewardIDs)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": gin.H{"quota": quota, "count": count}})
}
