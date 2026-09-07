package controller

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/oauth"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func StartTelegramOAuth(c *gin.Context) {
	intent := strings.TrimSpace(c.DefaultQuery("intent", model.AuthFlowIntentLogin))
	userID := 0
	if intent == model.AuthFlowIntentBind {
		userID = c.GetInt("id")
		if userID <= 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "未登录"})
			return
		}
	}
	state, flow, err := service.CreateTelegramAuthFlow(userID, intent)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{
		"state": state, "authorization_url": flow.AuthorizationURL(state),
	}})
}

func HandleTelegramOAuth(c *gin.Context) {
	state, code := strings.TrimSpace(c.Query("state")), strings.TrimSpace(c.Query("code"))
	if state == "" || code == "" {
		common.ApiError(c, errors.New("Telegram OAuth 回调参数无效"))
		return
	}
	session := sessions.Default(c)
	intent := model.AuthFlowIntentLogin
	userID := 0
	if id, ok := session.Get("id").(int); ok && id > 0 {
		intent, userID = model.AuthFlowIntentBind, id
	}
	flowRecord, flow, err := service.ReadTelegramAuthFlow(state, intent)
	if err != nil || (intent == model.AuthFlowIntentBind && flowRecord.UserId != userID) {
		common.ApiError(c, errors.New("Telegram OAuth 状态无效或已过期"))
		return
	}
	provider := oauth.NewTelegramOAuthProvider(nil)
	c.Set(oauth.TelegramOAuthFlowContextKey, flow)
	token, err := provider.ExchangeToken(c.Request.Context(), code, c)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	oauthUser, err := provider.GetUserInfo(c.Request.Context(), token)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if intent == model.AuthFlowIntentBind {
		if err = service.CommitTelegramBind(state, oauthUser.ProviderUserID, userID); err != nil {
			common.ApiError(c, err)
			return
		}
		c.Redirect(http.StatusFound, common.ThemeAwarePath("/console/personal"))
		return
	}
	if !common.RegisterEnabled {
		if !model.IsTelegramIdAlreadyTaken(oauthUser.ProviderUserID) {
			common.ApiError(c, errors.New("注册功能未开启"))
			return
		}
	}
	if model.IsTelegramIdAlreadyTaken(oauthUser.ProviderUserID) {
		user := &model.User{TelegramId: oauthUser.ProviderUserID}
		if err = user.FillUserByTelegramId(); err != nil || user.Status != common.UserStatusEnabled {
			common.ApiError(c, errors.New("Telegram 用户不可用"))
			return
		}
		if err = service.CommitTelegramExistingLogin(state, oauthUser.ProviderUserID, user.Id); err != nil {
			common.ApiError(c, err)
			return
		}
		setupLogin(user, c)
		return
	}
	user := &model.User{
		Username:    "telegram_" + strconv.Itoa(model.GetMaxUserId()+1),
		DisplayName: oauthUser.DisplayName,
		Role:        common.RoleCommonUser, Status: common.UserStatusEnabled,
	}
	if user.DisplayName == "" {
		user.DisplayName = oauthUser.Username
	}
	if user.DisplayName == "" {
		user.DisplayName = "Telegram User"
	}
	if oauthUser.Email != "" {
		user.Email = model.NormalizeEmail(oauthUser.Email)
	}
	inviterID := 0
	if aff, ok := session.Get("aff").(string); ok && aff != "" {
		inviterID, _ = model.GetUserIdByAffCode(aff)
	}
	if err = service.CommitTelegramLogin(state, oauthUser.ProviderUserID, user, inviterID); err != nil {
		common.ApiError(c, err)
		return
	}
	user.FinalizeOAuthUserCreation(inviterID)
	setupLogin(user, c)
}
