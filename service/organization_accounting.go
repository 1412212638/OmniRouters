package service

import (
	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
)

// RecordRelayUsage updates the usage counter belonging to the request's
// workspace. Organization funding already updates monetary usage during
// reservation/settlement, so this helper only records the completed request.
func RecordRelayUsage(info *relaycommon.RelayInfo, quota int) {
	if info == nil {
		return
	}
	if info.OrganizationId > 0 {
		if err := model.IncrementOrganizationRequestCount(info.OrganizationId); err != nil {
			common.SysError("failed to increment organization request count: " + err.Error())
			return
		}
		return
	}
	model.UpdateUserUsedQuotaAndRequestCount(info.UserId, quota)
}

// RecordRelayUsageDelta updates only settled monetary usage for an already
// recorded request, preserving the same personal/organization split.
func RecordRelayUsageDelta(info *relaycommon.RelayInfo, quota int) {
	if info == nil || quota == 0 {
		return
	}
	if info.OrganizationId > 0 {
		return
	}
	model.UpdateUserUsedQuota(info.UserId, quota)
}
