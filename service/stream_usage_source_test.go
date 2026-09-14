package service

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestMarkStreamUsageSource(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name       string
		usage      *dto.Usage
		localCount bool
		want       string
	}{
		{name: "missing", want: "missing"},
		{name: "provider usage", usage: &dto.Usage{BillingUsage: &dto.BillingUsage{}}, want: "upstream_actual"},
		{name: "estimated billing usage", usage: &dto.Usage{BillingUsage: &dto.BillingUsage{Estimated: true}}, want: "local_estimated"},
		{name: "local token counter", usage: &dto.Usage{}, localCount: true, want: "local_estimated"},
		{name: "provider usage wins over local marker", usage: &dto.Usage{BillingUsage: &dto.BillingUsage{}}, localCount: true, want: "upstream_actual"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(nil)
			if tt.localCount {
				common.SetContextKey(c, constant.ContextKeyLocalCountTokens, true)
			}
			info := &relaycommon.RelayInfo{StreamStatus: relaycommon.NewStreamStatus()}
			markStreamUsageSource(c, info, tt.usage)
			require.Equal(t, tt.want, info.StreamStatus.UsageSource)
			require.Equal(t, tt.usage != nil, info.StreamStatus.UsagePresent)
		})
	}
}

func TestShouldSkipEstimatedClientGoneUsage(t *testing.T) {
	tests := []struct {
		name   string
		status relaycommon.StreamEndReason
		source string
		want   bool
	}{
		{name: "normal stream", status: relaycommon.StreamEndReasonDone, source: "local_estimated", want: false},
		{name: "client gone without usage", status: relaycommon.StreamEndReasonClientGone, source: "missing", want: true},
		{name: "client gone with local estimate", status: relaycommon.StreamEndReasonClientGone, source: "local_estimated", want: true},
		{name: "client gone with provider usage", status: relaycommon.StreamEndReasonClientGone, source: "upstream_actual", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &relaycommon.RelayInfo{StreamStatus: relaycommon.NewStreamStatus()}
			info.StreamStatus.SetEndReason(tt.status, nil)
			info.StreamStatus.UsageSource = tt.source
			require.Equal(t, tt.want, shouldSkipEstimatedClientGoneUsage(info))
		})
	}
}
