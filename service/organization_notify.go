package service

import (
	"fmt"
	"html"
	"net/mail"
	"net/url"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/bytedance/gopkg/util/gopool"
)

const (
	OrganizationNotificationInvitation           = "invitation"
	OrganizationNotificationInvitationExpiring   = "invitation_expiring"
	OrganizationNotificationRoleChanged          = "role_changed"
	OrganizationNotificationRemoved              = "removed"
	OrganizationNotificationOwnershipTransferred = "ownership_transferred"
)

type OrganizationNotification struct {
	Kind           string
	OrganizationId int
	UserId         int
	Email          string
	Role           string
	ExpiresAt      int64
}

type organizationNotificationMail struct {
	Kind             string
	OrganizationName string
	Recipient        string
	Role             string
	ExpiresAt        int64
	WorkspaceURL     string
}

func organizationEmailConfigured() bool {
	common.OptionMapRWMutex.RLock()
	provider := strings.TrimSpace(common.OptionMap["EmailProvider"])
	resendKey := strings.TrimSpace(common.OptionMap["ResendApiKey"])
	resendFrom := strings.TrimSpace(common.OptionMap["ResendFrom"])
	common.OptionMapRWMutex.RUnlock()
	if provider == "resend" {
		return resendKey != "" && resendFrom != ""
	}
	return (provider == "" || provider == "smtp") && strings.TrimSpace(common.SMTPServer) != "" && (strings.TrimSpace(common.SMTPFrom) != "" || strings.TrimSpace(common.SMTPAccount) != "")
}

// TryNotifyOrganization runs only after the caller's database transaction has
// committed. Mail delivery is optional and never changes the saved action.
func TryNotifyOrganization(notification OrganizationNotification) {
	if notification.OrganizationId <= 0 || !organizationEmailConfigured() {
		return
	}
	gopool.Go(func() {
		organization, err := model.GetOrganizationById(notification.OrganizationId)
		if err != nil {
			common.SysLog(fmt.Sprintf("organization notification load failed: organization_id=%d error=%s", notification.OrganizationId, err.Error()))
			return
		}
		recipient := strings.TrimSpace(notification.Email)
		if recipient == "" && notification.UserId > 0 {
			user, err := model.GetUserById(notification.UserId, false)
			if err != nil {
				common.SysLog(fmt.Sprintf("organization notification user load failed: user_id=%d error=%s", notification.UserId, err.Error()))
				return
			}
			recipient = strings.TrimSpace(user.Email)
		}
		details := organizationNotificationMail{Kind: notification.Kind, OrganizationName: organization.Name, Recipient: recipient, Role: notification.Role, ExpiresAt: notification.ExpiresAt, WorkspaceURL: organizationWorkspaceURL()}
		if err := sendOrganizationNotificationMail(details, common.SendEmail); err != nil {
			common.SysLog(fmt.Sprintf("organization notification failed: organization_id=%d user_id=%d kind=%s error=%s", notification.OrganizationId, notification.UserId, notification.Kind, err.Error()))
		}
	})
}

func organizationWorkspaceURL() string {
	base, err := url.Parse(strings.TrimSpace(system_setting.ServerAddress))
	if err != nil || (base.Scheme != "https" && base.Scheme != "http") || base.Host == "" {
		return ""
	}
	base.Path = strings.TrimRight(base.Path, "/") + "/organizations"
	base.RawQuery, base.Fragment = "", ""
	return base.String()
}

func sendOrganizationNotificationMail(details organizationNotificationMail, send func(string, string, string) error) error {
	if !organizationEmailConfigured() || strings.TrimSpace(details.Recipient) == "" {
		return nil
	}
	address, err := mail.ParseAddress(details.Recipient)
	if err != nil || address.Address != details.Recipient {
		return fmt.Errorf("invalid organization notification recipient")
	}
	subject, content := buildOrganizationNotificationMail(details)
	if subject == "" {
		return nil
	}
	return send(subject, details.Recipient, content)
}

func buildOrganizationNotificationMail(details organizationNotificationMail) (string, string) {
	name := html.EscapeString(details.OrganizationName)
	role := details.Role
	if common.IsEmailLanguageEnglish() {
		switch role {
		case model.OrganizationMemberOwner:
			role = "Owner"
		case model.OrganizationMemberAdmin:
			role = "Admin"
		default:
			role = "Member"
		}
	} else {
		switch role {
		case model.OrganizationMemberOwner:
			role = "所有者"
		case model.OrganizationMemberAdmin:
			role = "管理员"
		default:
			role = "成员"
		}
	}
	var subject, content string
	if common.IsEmailLanguageEnglish() {
		switch details.Kind {
		case OrganizationNotificationInvitation:
			subject = "Organization invitation"
			content = fmt.Sprintf("<p>You have been invited to join <strong>%s</strong> as %s.</p><p>Sign in with this email address and accept the invitation in Organizations. The invitation expires at %s UTC.</p>", name, role, time.Unix(details.ExpiresAt, 0).UTC().Format("2006-01-02 15:04"))
		case OrganizationNotificationInvitationExpiring:
			subject = "Organization invitation expires soon"
			content = fmt.Sprintf("<p>Your invitation to join <strong>%s</strong> as %s expires at %s UTC.</p><p>Sign in with this email address and accept the invitation in Organizations before it expires.</p>", name, role, time.Unix(details.ExpiresAt, 0).UTC().Format("2006-01-02 15:04"))
		case OrganizationNotificationRoleChanged:
			subject = "Organization role updated"
			content = fmt.Sprintf("<p>Your role in <strong>%s</strong> is now %s.</p>", name, role)
		case OrganizationNotificationRemoved:
			subject = "Organization membership ended"
			content = fmt.Sprintf("<p>Your membership in <strong>%s</strong> has ended. Your API keys for this organization have been disabled. Historical usage remains with the organization.</p>", name)
		case OrganizationNotificationOwnershipTransferred:
			subject = "Organization ownership transferred"
			content = fmt.Sprintf("<p>Ownership of <strong>%s</strong> has been transferred. Your current role is %s.</p>", name, role)
		}
	} else {
		switch details.Kind {
		case OrganizationNotificationInvitation:
			subject = "组织邀请"
			content = fmt.Sprintf("<p>您已受邀以%s身份加入组织 <strong>%s</strong>。</p><p>请使用此邮箱登录，在“组织”页面接受邀请。邀请有效期至 %s（UTC）。</p>", role, name, time.Unix(details.ExpiresAt, 0).UTC().Format("2006-01-02 15:04"))
		case OrganizationNotificationInvitationExpiring:
			subject = "组织邀请即将到期"
			content = fmt.Sprintf("<p>您以%s身份加入组织 <strong>%s</strong> 的邀请将于 %s（UTC）到期。</p><p>如需加入，请在到期前使用此邮箱登录，在“组织”页面接受邀请。</p>", role, name, time.Unix(details.ExpiresAt, 0).UTC().Format("2006-01-02 15:04"))
		case OrganizationNotificationRoleChanged:
			subject = "组织角色已更新"
			content = fmt.Sprintf("<p>您在组织 <strong>%s</strong> 中的角色已更新为%s。</p>", name, role)
		case OrganizationNotificationRemoved:
			subject = "组织成员身份已结束"
			content = fmt.Sprintf("<p>您在组织 <strong>%s</strong> 中的成员身份已结束，对应组织 API 密钥已停用，历史使用记录仍保留在组织中。</p>", name)
		case OrganizationNotificationOwnershipTransferred:
			subject = "组织所有权已转移"
			content = fmt.Sprintf("<p>组织 <strong>%s</strong> 的所有权已转移，您当前的角色为%s。</p>", name, role)
		}
	}
	if subject == "" {
		return "", ""
	}
	if details.WorkspaceURL != "" {
		label := "打开组织工作区"
		if common.IsEmailLanguageEnglish() {
			label = "Open organization workspace"
		}
		content += fmt.Sprintf(`<p><a href="%s">%s</a></p>`, html.EscapeString(details.WorkspaceURL), label)
	}
	return strings.ReplaceAll(strings.ReplaceAll(common.SystemName, "\r", ""), "\n", "") + " - " + subject, content
}
