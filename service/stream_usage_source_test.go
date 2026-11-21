package service

import (
	"errors"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestAppendStreamStatusLogsDownstreamDiagnosticsForAdmins(t *testing.T) {
	info := &relaycommon.RelayInfo{IsStream: true, StreamStatus: relaycommon.NewStreamStatus()}
	info.StreamStatus.RecordDownstreamFlush()
	other := model.NewLogOther()
	appendStreamStatus(info, other)
	require.NotContains(t, other.Snapshot(), "admin_info")

	info.StreamStatus.RecordDownstreamWriteError(errors.New("connection reset by peer"))
	info.StreamStatus.SetEndReason(relaycommon.StreamEndReasonClientGone, nil)
	info.StreamStatus.RecordDownstreamContextError("request context done: context canceled")
	info.StreamStatus.RecordDownstreamContextError("request context done: context canceled")
	ctx, _ := gin.CreateTestContext(nil)
	markStreamUsageSource(ctx, info, &dto.Usage{TotalTokens: 12})
	appendStreamStatus(info, other)

	fields := other.Snapshot()
	require.NotContains(t, fields, "stream_downstream")
	stream := fields["stream_status"].(map[string]any)
	require.Equal(t, "error", stream["status"])
	require.Equal(t, 1, stream["error_count"])
	require.Equal(t, "upstream_actual", stream["usage_source"])
	details := fields["admin_info"].(map[string]any)["stream_downstream"].(map[string]any)
	require.Equal(t, "connection reset by peer", details["first_write_error"])
	require.Equal(t, 1, details["repeated_context_errors"])
	require.Equal(t, true, details["upstream_usage_received"])
	require.False(t, details["context_done_at"].(time.Time).IsZero())
	require.False(t, details["last_successful_flush_at"].(time.Time).IsZero())
	require.False(t, details["first_write_error_at"].(time.Time).IsZero())
	require.Equal(t, details["first_write_error_at"], details["interrupted_at"])
	require.GreaterOrEqual(t, details["ms_since_last_successful_flush"].(int64), int64(0))
}

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
