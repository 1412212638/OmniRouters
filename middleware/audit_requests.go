package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

// AuditRequests runs outside authentication so rejected login attempts are
// visible as anonymous events. No request body, query or credential is stored.
func AuditRequests() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/audit") {
			c.Next()
			return
		}
		writer := &auditResponseWriter{ResponseWriter: c.Writer, body: bytes.NewBuffer(nil), maxSize: 16 * 1024}
		c.Writer = writer
		c.Next()
		c.Writer = writer.ResponseWriter
		if c.GetBool("audit_independent_recorded") { return }
		action := c.GetString("audit_action")
		category := c.GetString("audit_category")
		route := c.FullPath()
		if route == "" { return }
		if action == "" {
			switch route {
			case "/api/user/login", "/api/user/login/2fa", "/api/user/passkey/login/finish":
				action, category = "login.attempt", "security"
			case "/api/oauth/:provider", "/api/oauth/wechat", "/api/oauth/telegram", "/api/oauth/telegram/login":
				action, category = "oauth.callback", "security"
			case "/api/user/auth/logout", "/api/user/logout":
				action, category = "session.logout", "security"
			}
		}
		userID := c.GetInt("id")
		if userID <= 0 && category != "security" { return }
		if action == "" {
			if known := auditRouteActions[c.Request.Method+" "+route]; known != "" {
				action, category = known, "operation"
			} else if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead || route == "/api/user/auth/refresh" {
				action, category = "request", "access"
			} else {
				action, category = auditRouteActions[c.Request.Method+" "+route], "operation"
				if action == "" { action = "generic" }
			}
		}
		method := "session"
		if userID == 0 { method = "anonymous" }
		switch route {
		case "/api/user/login": method = "password"
		case "/api/user/login/2fa": method = "two_factor"
		case "/api/user/passkey/login/finish": method = "passkey"
		case "/api/oauth/:provider", "/api/oauth/wechat", "/api/oauth/telegram", "/api/oauth/telegram/login": method = "oauth"
		}
		if c.GetBool("use_access_token") {
			if _, ok := GetAuthIdentity(c); !ok { method = "personal_token" }
		}
		target := c.GetString("audit_target")
		if target == "" {
			target = route
			for _, key := range []string{"id", "sid", "provider_id"} {
				if value := c.Param(key); value != "" { target += " #"+value }
			}
		}
		if len(target) > 512 { target = target[:512] }
		outcome, failure := auditOutcome(writer.Status(), writer.body.Bytes(), writer.body.Len() >= writer.maxSize)
		details := model.AuditDetails{Version: 1, Method: c.Request.Method, Route: route, Target: target, AuthMethod: method, HTTPStatus: writer.Status(), Outcome: outcome, Failure: failure}
		if value, ok := c.Get("audit_changes"); ok { details.Changes, _ = value.(map[string]model.AuditChange) }
		if value, ok := c.Get("audit_label_params"); ok { details.Params, _ = value.(map[string]interface{}) }
		encoded, err := common.Marshal(details)
		if err != nil { common.SysError("encode audit metadata: "+err.Error()); return }
		if len(encoded) > 64 * 1024 {
			for field := range details.Changes { details.Changes[field] = model.AuditChange{Before: "[REDACTED]", After: "[REDACTED]"} }
			encoded, err = common.Marshal(details)
			if err != nil { common.SysError("encode bounded audit metadata: "+err.Error()); return }
		}
		entry := &model.AuditLog{UserId: userID, Username: c.GetString("username"), Category: category, Action: action, Ip: c.ClientIP(), Success: outcome == "success", Outcome: outcome, RequestId: c.GetString(common.RequestIdKey), Other: string(encoded)}
		if method == "personal_token" { entry.TokenRef = model.AccessTokenFingerprint(c.GetHeader("Authorization")) }
		if ref := c.GetString("audit_token_ref"); ref != "" { entry.TokenRef = ref }
		if err := model.RecordAuditLog(entry); err != nil { common.SysError("write request audit: "+err.Error()) }
	}
}

func auditOutcome(status int, body []byte, truncated bool) (string, string) {
	if status >= 400 { return "failed", fmt.Sprintf("HTTP %d %s", status, http.StatusText(status)) }
	if status == http.StatusNoContent { return "success", "" }
	if truncated { return "unknown", "" }
	var result struct {
		Success *bool `json:"success"`
		Data json.RawMessage `json:"data"`
	}
	if common.Unmarshal(body, &result) == nil && result.Success != nil {
		if !*result.Success { return "failed", "business_rejected" }
		var data struct { Require2FA bool `json:"require_2fa"` }
		if common.Unmarshal(result.Data, &data) == nil && data.Require2FA { return "pending", "" }
		return "success", ""
	}
	return "unknown", ""
}
