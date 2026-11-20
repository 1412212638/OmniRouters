package controller

import (
	"crypto/hmac"
	"net/http"
	"net/url"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/gin-gonic/gin"
)

func marketingSESCallbackKey() string {
	return common.GenerateHMAC("marketing-ses-callback-v1")
}

func marketingSMTPSettingsResponse(cfg common.MarketingSMTPConfig) gin.H {
	configured := cfg.Password != ""
	cfg.Password = ""
	var callback string
	base, err := url.Parse(strings.TrimSpace(system_setting.ServerAddress))
	if err == nil && base.Scheme == "https" && base.Hostname() != "" && base.User == nil && base.RawQuery == "" && base.Fragment == "" {
		callback = strings.TrimRight(base.String(), "/") + "/api/marketing_email/ses_callback?key=" + marketingSESCallbackKey()
	}
	return gin.H{"settings": cfg, "password_configured": configured, "callback_url": callback}
}

func GetMarketingSMTPSettings(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	cfg, err := model.GetMarketingSMTPConfig()
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, marketingSMTPSettingsResponse(cfg))
}

func UpdateMarketingSMTPSettings(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	var cfg common.MarketingSMTPConfig
	if err := common.DecodeJson(http.MaxBytesReader(c.Writer, c.Request.Body, 8192), &cfg); err != nil {
		common.ApiErrorMsg(c, "invalid marketing SMTP settings")
		return
	}
	current, err := model.GetMarketingSMTPConfig()
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if cfg.Password == "" && strings.TrimSpace(cfg.Server) == current.Server && cfg.Port == current.Port &&
		strings.TrimSpace(cfg.Account) == current.Account {
		cfg.Password = current.Password
	}
	if err := cfg.Validate(); err != nil {
		common.ApiError(c, err)
		return
	}
	raw, err := common.Marshal(cfg)
	if err == nil {
		err = model.UpdateOption(common.MarketingSMTPOption, string(raw))
	}
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, marketingSMTPSettingsResponse(cfg))
}

func TestMarketingSMTPSettings(c *gin.Context) {
	var req struct {
		Recipient string `json:"recipient"`
	}
	if err := common.DecodeJson(http.MaxBytesReader(c.Writer, c.Request.Body, 1024), &req); err != nil ||
		!common.IsPlainEmailAddress(strings.TrimSpace(req.Recipient)) {
		common.ApiErrorMsg(c, "invalid test recipient")
		return
	}
	cfg, err := model.GetMarketingSMTPConfig()
	if err == nil {
		err = cfg.Validate()
	}
	if err == nil {
		err = service.SendMarketingMessage(c.Request.Context(), cfg, "Marketing email test", strings.TrimSpace(req.Recipient), "<p>Marketing email delivery test.</p>", "")
	}
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, nil)
}

func MarketingSESEvent(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	if !hmac.Equal([]byte(c.Query("key")), []byte(marketingSESCallbackKey())) {
		c.Status(http.StatusForbidden)
		return
	}
	if c.Request.Method == http.MethodGet {
		c.Status(http.StatusOK)
		return
	}
	var event struct {
		Event      string `json:"event"`
		BounceType string `json:"bounceType"`
		Email      string `json:"email"`
		From       string `json:"from"`
		MessageID  string `json:"messageId"`
		Proof      string `json:"X-Tencentcloudses-Cb-Proof"`
	}
	if err := common.DecodeJson(http.MaxBytesReader(c.Writer, c.Request.Body, 32768), &event); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	reason := event.Event
	switch {
	case event.Event == "bounce" && event.BounceType == "hard_bounce":
		reason = event.BounceType
	case event.Event == "spamreport":
	default:
		c.Status(http.StatusOK)
		return
	}
	if !common.IsPlainEmailAddress(event.Email) || !common.IsPlainEmailAddress(event.From) ||
		event.MessageID == "" || !hmac.Equal([]byte(event.Proof), []byte(service.MarketingSESProof(event.MessageID, event.From, event.Email))) {
		c.Status(http.StatusForbidden)
		return
	}
	if err := model.SuppressMarketingEmail(event.Email, reason); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusOK)
}
