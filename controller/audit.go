package controller

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"

	"github.com/gin-gonic/gin"
)

// auditContentTemplates 将稳定的操作标识 action 映射为英文兜底模板，渲染后写入
// Log.Content（供导出 / 经典前端等非本地化消费者使用）。占位符为 ${name}，由该
// action 的 params 填充。本地化展示文案在前端 i18n 模板中维护，本表是语言中立的
// 英文基线——调用方因此无需在每个埋点处手写句子（避免与 params 重复书写同一份值）。
var auditContentTemplates = map[string]string{
	"user.create":           "Created user ${username} (role ${role})",
	"user.update":           "Updated user ${username} (ID: ${id})",
	"user.delete":           "Deleted user ${username} (ID: ${id})",
	"user.manage":           "Performed ${action} on user ${username} (ID: ${id})",
	"user.quota_add":        "Increased user quota by ${quota}",
	"user.quota_subtract":   "Decreased user quota by ${quota}",
	"user.quota_override":   "Overrode user quota from ${from} to ${to}",
	"user.binding_clear":    "Cleared ${bindingType} binding for user ${username}",
	"user.2fa_disable":      "Force-disabled two-factor authentication for the user",
	"user.passkey_register": "Registered a passkey",
	"user.passkey_delete":   "Deleted a passkey",
	"user.reset_passkey":    "Reset the user passkey",
	"option.update":         "Updated system setting ${key}",

	"option.passkey_domains":           "Updated Passkey domains: removed ${domains}; affected ${known}; unknown ${unknown}",
	"option.passkey_domains_confirmed": "Confirmed removal of Passkey domains: ${domains}; affected ${known}; unknown ${unknown}",
	"option.passkey_domains_blocked":   "Passkey domain change blocked: ${domains}; affected ${known}; unknown ${unknown}",
	"option.passkey_domains_failed":    "Passkey domain update failed",

	"channel.create":             "Created channel ${name} (type ${type}, count ${count})",
	"channel.update":             "Updated channel ${name} (ID: ${id})",
	"channel.delete":             "Deleted channel ${name} (ID: ${id})",
	"channel.delete_batch":       "Batch deleted ${count} channels",
	"channel.delete_disabled":    "Deleted all disabled channels (${count})",
	"channel.key_view":           "Viewed channel key ${name} (ID: ${id})",
	"channel.tag_disable":        "Disabled channels with tag ${tag}",
	"channel.tag_enable":         "Enabled channels with tag ${tag}",
	"channel.tag_edit":           "Edited channels with tag ${tag}",
	"channel.tag_batch_set":      "Batch set tag for ${count} channels",
	"channel.copy":               "Copied channel (source ID: ${sourceId}) to ${name} (new ID: ${id})",
	"channel.multi_key_manage":   "Multi-key management ${action} on channel (ID: ${id})",
	"channel.upstream_apply":     "Applied upstream model changes to channel (ID: ${id})",
	"channel.upstream_apply_all": "Applied upstream model changes to ${count} channels",

	"redemption.create": "Created ${count} redemption codes named ${name} (${quota} each)",

	"subscription.plan_reset":      "Reset active subscriptions for plan ${plan_id}",
	"subscription.user_plan_reset": "Reset active plan ${plan_id} subscriptions for user ${target_user_id}",
}

func recordPasskeyDomainAudit(c *gin.Context, change *model.PasskeyDomainChange, confirmed bool, err error) {
	confirmed = confirmed && err == nil && change != nil && len(change.RemovedRPIDs) > 0
	params := map[string]any{"success": err == nil, "confirmed": confirmed}
	if change != nil {
		params["domains"] = strings.Join(change.RemovedRPIDs, ", ")
		params["removed_rp_ids"] = change.RemovedRPIDs
		params["known"] = change.AffectedCredentials
		params["unknown"] = change.UnknownCredentials
		params["previous_rp_id"] = change.PreviousRPID
		params["effective_rp_id"] = change.EffectiveRPID
	}
	action := "option.passkey_domains"
	if errors.Is(err, model.ErrPasskeyDomainRemovalConfirmation) {
		action = "option.passkey_domains_blocked"
	} else if err != nil {
		action = "option.passkey_domains_failed"
	} else if confirmed && change != nil && len(change.RemovedRPIDs) > 0 {
		action = "option.passkey_domains_confirmed"
	}
	auditInfo := map[string]interface{}{
		"method": c.Request.Method, "route": c.FullPath(), "path": c.FullPath(),
		"status": c.Writer.Status(), "success": err == nil,
	}
	model.RecordOperationAuditLog(c.GetInt("id"), auditContentEN(action, params), c.ClientIP(), action, params, auditOperatorInfo(c), auditInfo)
	markAuditLogged(c)
}

// auditContentEN 按 action 模板渲染英文兜底文本；未登记的 action 退回 action 本身。
func auditContentEN(action string, params map[string]interface{}) string {
	tmpl, ok := auditContentTemplates[action]
	if !ok {
		return action
	}
	return os.Expand(tmpl, func(key string) string {
		if v, ok := params[key]; ok {
			return fmt.Sprintf("%v", v)
		}
		return ""
	})
}

// auditOperatorInfo 从上下文构建操作者身份信息（管理员 id/用户名/角色）。
func auditOperatorInfo(c *gin.Context) map[string]interface{} {
	return map[string]interface{}{
		"admin_id":       c.GetInt("id"),
		"admin_username": c.GetString("username"),
		"admin_role":     c.GetInt("role"),
		"auth_method":    auditAuthMethod(c),
	}
}

func auditAuthMethod(c *gin.Context) string {
	if c.GetBool("use_access_token") {
		return "access_token"
	}
	return "session"
}

// Defer the independent audit entry until the actual response is known. Keep
// the authenticated operator distinct from the account affected by an action.
func recordSecurityEvent(c *gin.Context, entry *model.AuditLog) error {
	c.Set("audit_action", entry.Action)
	c.Set("audit_category", entry.Category)
	c.Set("audit_target", fmt.Sprintf("user #%d", entry.UserId))
	if c.GetInt("id") == 0 { c.Set("id", entry.UserId); c.Set("username", entry.Username) }
	if entry.TokenRef != "" { c.Set("audit_token_ref", entry.TokenRef) }
	return nil
}

// markAuditLogged 标记当前请求已在 handler 内手动记录审计日志，
// 使鉴权链路中的审计兜底（finishAdminAudit）跳过兜底记录，避免重复。
func markAuditLogged(c *gin.Context) {
	common.SetContextKey(c, constant.ContextKeyAuditLogged, true)
}

// recordManageAudit 记录一条由操作者本人归属的管理/高危审计日志（资源类操作：
// 渠道 / 系统设置 / 兑换码等）。content 由 action+params 自动渲染。
func recordManageAudit(c *gin.Context, action string, params map[string]interface{}) {
	recordManageAuditFor(c, c.GetInt("id"), action, params)
}

// recordManageAuditFor 记录一条管理审计日志，日志归属于操作者；targetUserId
// 只表示被操作用户，用于在结构化参数中保留目标上下文。
func recordManageAuditFor(c *gin.Context, targetUserId int, action string, params map[string]interface{}) {
	c.Set("audit_action", action)
	c.Set("audit_category", "operation")
	// Only known non-secret scalar parameters used by localized action labels.
	labelParams := map[string]interface{}{}
	for _, key := range []string{"id", "name", "username", "role", "count", "type", "quota", "from", "to", "plan_id", "target_user_id", "bindingType", "sourceId", "action"} {
		if value, ok := params[key]; ok {
			switch v := value.(type) {
			case string: if len(v) <= 256 { labelParams[key] = v }
			case int, int64, float64, bool: labelParams[key] = v
			}
		}
	}
	if action == "option.update" { labelParams["key"] = params["key"] }
	c.Set("audit_label_params", labelParams)
	if action == "user.quota_override" {
		c.Set("audit_changes", map[string]model.AuditChange{"quota": {Before: labelParams["from"], After: labelParams["to"]}})
	}
	if params != nil {
		for _, key := range []string{"target_user_id", "id", "plan_id", "key", "name"} {
			if value, ok := params[key]; ok { c.Set("audit_target", fmt.Sprintf("%s: %v", key, value)); break }
		}
	}
	if targetUserId > 0 && targetUserId != c.GetInt("id") { c.Set("audit_target", fmt.Sprintf("user #%d", targetUserId)) }
	if params == nil {
		params = map[string]interface{}{}
	}
	operatorUserId := c.GetInt("id")
	if _, ok := params["target_user_id"]; !ok && targetUserId > 0 && targetUserId != operatorUserId {
		params["target_user_id"] = targetUserId
	}
	model.RecordOperationAuditLog(operatorUserId, auditContentEN(action, params), c.ClientIP(), action, params, auditOperatorInfo(c), nil)
	markAuditLogged(c)
}

// recordUserSecurityAudit 记录普通用户自己的安全敏感操作（如 passkey 绑定/解绑）。
// 这类日志没有管理员操作者，不写 admin_info；同时不依赖 AdminAuth/RootAuth 的兜底。
func recordUserSecurityAudit(c *gin.Context, userId int, action string, params map[string]interface{}) {
	c.Set("audit_action", action)
	c.Set("audit_category", "security")
	c.Set("audit_target", fmt.Sprintf("user #%d", userId))
	model.RecordOperationAuditLog(userId, auditContentEN(action, params), c.ClientIP(), action, params, nil, nil)
}
