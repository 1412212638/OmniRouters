package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"

	"github.com/bytedance/gopkg/util/gopool"
)

func TryNotifyTopUpSuccess(topUp *model.TopUp, quotaAdded int) {
	if !common.TopUpSuccessEmailEnabled {
		return
	}
	if topUp == nil || topUp.UserId <= 0 || quotaAdded <= 0 {
		return
	}

	topUpSnapshot := *topUp
	gopool.Go(func() {
		sendTopUpSuccessEmail(&topUpSnapshot, quotaAdded)
	})
}

func sendTopUpSuccessEmail(topUp *model.TopUp, quotaAdded int) {
	user, err := model.GetUserById(topUp.UserId, false)
	if err != nil {
		common.SysLog(fmt.Sprintf("failed to load user for top-up success email: user_id=%d trade_no=%s error=%s", topUp.UserId, topUp.TradeNo, err.Error()))
		return
	}

	emailToUse := strings.TrimSpace(user.GetSetting().NotificationEmail)
	if emailToUse == "" {
		emailToUse = strings.TrimSpace(user.Email)
	}
	if emailToUse == "" {
		common.SysLog(fmt.Sprintf("user %d has no email, skip top-up success email", topUp.UserId))
		return
	}

	paidAt := topUp.CompleteTime
	if paidAt == 0 {
		paidAt = common.GetTimestamp()
	}
	paymentProvider := topUp.EffectivePaymentProvider()
	params := common.TopUpSuccessMailParams{
		SystemName:      common.SystemName,
		Username:        user.Username,
		TradeNo:         topUp.TradeNo,
		PaymentMethod:   topUp.PaymentMethod,
		PaymentProvider: paymentProvider,
		Amount:          fmt.Sprintf("%.2f", topUp.Money),
		Quota:           logger.FormatQuota(quotaAdded),
		Balance:         logger.FormatQuota(user.Quota),
		PaidAt:          time.Unix(paidAt, 0).Local().Format("2006-01-02 15:04:05"),
	}

	subject, content := common.BuildTopUpSuccessMail(params)
	if err := common.SendEmail(subject, emailToUse, content); err != nil {
		common.SysLog(fmt.Sprintf("failed to send top-up success email: user_id=%d trade_no=%s email=%s error=%s", topUp.UserId, topUp.TradeNo, common.MaskEmail(emailToUse), err.Error()))
		return
	}
	common.SysLog(fmt.Sprintf("top-up success email sent: user_id=%d trade_no=%s email=%s", topUp.UserId, topUp.TradeNo, common.MaskEmail(emailToUse)))
}
