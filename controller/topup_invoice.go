package controller

import (
	"net/http"
	neturl "net/url"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting"
	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/invoice"
)

// GetTopUpInvoice returns a fresh, provider-hosted invoice URL for a
// successful order. The order owner is derived from the authenticated request;
// clients cannot choose another user's buyer identity.
func GetTopUpInvoice(c *gin.Context) {
	topUpID, parseErr := strconv.Atoi(c.Param("id"))
	if parseErr != nil || topUpID <= 0 {
		common.ApiErrorMsg(c, "充值记录不存在")
		return
	}
	topUp := model.GetTopUpById(topUpID)
	if topUp == nil {
		common.ApiErrorMsg(c, "充值记录不存在")
		return
	}
	if c.GetInt("id") != topUp.UserId && c.GetInt("role") < common.RoleAdminUser {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"success": false, "message": "无权访问该充值记录"})
		return
	}
	if topUp.Status != common.TopUpStatusSuccess {
		common.ApiErrorMsg(c, "只有成功的充值记录可以获取发票")
		return
	}

	var url string
	var err error
	switch topUp.EffectivePaymentProvider() {
	case model.PaymentProviderStripe:
		if topUp.StripeInvoiceID == "" {
			common.ApiErrorMsg(c, "该充值暂无可用发票")
			return
		}
		stripe.Key = setting.StripeApiSecret
		stripeInvoice, getErr := invoice.Get(topUp.StripeInvoiceID, nil)
		if getErr != nil || stripeInvoice == nil {
			err = getErr
		} else {
			url = stripeInvoice.InvoicePDF
			if url == "" {
				url = stripeInvoice.HostedInvoiceURL
			}
		}
	case model.PaymentProviderWaffoPancake:
		url, err = service.CreateWaffoPancakeCustomerPortalLink(c.Request.Context(), service.WaffoPancakeBuyerIdentityFromUserID(topUp.UserId))
	default:
		common.ApiErrorMsg(c, "该支付渠道暂不支持发票")
		return
	}
	if err != nil || url == "" {
		common.ApiErrorMsg(c, "发票暂时不可用，请稍后重试")
		return
	}
	parsedURL, parseErr := neturl.Parse(url)
	if parseErr != nil || parsedURL.Scheme != "https" || parsedURL.User != nil ||
		!(parsedURL.Host == "stripe.com" || strings.HasSuffix(parsedURL.Host, ".stripe.com") || parsedURL.Host == "pancake.waffo.ai") {
		common.ApiErrorMsg(c, "发票暂时不可用，请稍后重试")
		return
	}
	c.Header("Cache-Control", "no-store")
	c.Header("Referrer-Policy", "no-referrer")
	common.ApiSuccess(c, gin.H{"url": url})
}
