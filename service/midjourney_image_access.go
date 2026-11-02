package service

import (
	"net/url"
	"strings"

	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/system_setting"
)

// BuildMidjourneyImageURL keeps existing personal image links compatible.
// Organization images require a capability bound to that exact task.
func BuildMidjourneyImageURL(task *model.Midjourney) string {
	if task == nil {
		return ""
	}
	result := strings.TrimRight(system_setting.ServerAddress, "/") + "/mj/image/" + url.PathEscape(task.MjId)
	if task.OrganizationId > 0 {
		access, err := IssueTaskArtifactAccess("mj_"+task.MjId, "image")
		if err != nil {
			return ""
		}
		result += "?access=" + url.QueryEscape(access)
	}
	return result
}
