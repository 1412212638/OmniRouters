package controller

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
)

const impersonateUserAction = "user.impersonation.start"

func ImpersonateUser(c *gin.Context) {
	c.Set("audit_action", impersonateUserAction)
	c.Set("audit_category", model.AuditCategorySecurity)
	common.SetContextKey(c, constant.ContextKeyAuditLogged, true)
	actor, ok := requireBrowserSession(c)
	if !ok {
		return
	}
	targetID, err := strconv.Atoi(c.Param("id"))
	if err != nil || targetID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false})
		return
	}
	targetLabel := "user #" + strconv.Itoa(targetID)
	c.Set("audit_target", targetLabel)
	details, err := common.Marshal(model.AuditDetails{
		Version:    1,
		Method:     c.Request.Method,
		Route:      c.FullPath(),
		Target:     targetLabel,
		AuthMethod: "session",
		HTTPStatus: http.StatusOK,
		Outcome:    "success",
		Params:     map[string]any{"target_user_id": targetID},
	})
	if err != nil {
		writeAuthSessionError(c, err)
		return
	}
	audit := &model.AuditLog{
		UserId:    actor.UserID,
		Category:  model.AuditCategorySecurity,
		Action:    impersonateUserAction,
		Ip:        c.ClientIP(),
		Success:   true,
		Outcome:   "success",
		RequestId: c.GetString(common.RequestIdKey),
		Other:     string(details),
	}
	bundle, target, err := service.SwitchToUserSession(c.Request.Context(), actor, targetID, c.ClientIP(), c.Request.UserAgent(), audit)
	if errors.Is(err, model.ErrImpersonationUnavailable) {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "code": "USER_IMPERSONATION_UNAVAILABLE"})
		return
	}
	if err != nil {
		writeAuthSessionError(c, err)
		return
	}
	c.Set("audit_independent_recorded", true)
	service.WriteRefreshCookie(c, bundle.RefreshToken)
	data := authRotationData(bundle)
	data["user"] = buildSelfUserData(target)
	common.ApiSuccess(c, data)
}
