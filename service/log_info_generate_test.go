package service

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/pkg/billingexpr"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/stretchr/testify/require"
)

func TestInjectTieredBillingInfoRecordsOriginalQuota(t *testing.T) {
	relayInfo := &relaycommon.RelayInfo{
		TieredBillingSnapshot: &billingexpr.BillingSnapshot{
			BillingMode:              "tiered_expr",
			ExprString:                `tier("base", p * 2)`,
			EstimatedQuotaBeforeGroup: 80.4,
		},
	}

	other := model.NewLogOther()
	InjectTieredBillingInfo(other, relayInfo, &billingexpr.TieredResult{
		ActualQuotaBeforeGroup: 123.5,
	})

	var values map[string]any
	require.NoError(t, common.UnmarshalJsonStr(other.JSONString(), &values))
	require.Equal(t, float64(124), values["original_quota"])
}

func TestInjectTieredBillingInfoFallsBackToEstimatedOriginalQuota(t *testing.T) {
	relayInfo := &relaycommon.RelayInfo{
		TieredBillingSnapshot: &billingexpr.BillingSnapshot{
			BillingMode:              "tiered_expr",
			ExprString:                `tier("base", p * 2)`,
			EstimatedQuotaBeforeGroup: 80.4,
		},
	}

	other := model.NewLogOther()
	InjectTieredBillingInfo(other, relayInfo, nil)

	var values map[string]any
	require.NoError(t, common.UnmarshalJsonStr(other.JSONString(), &values))
	require.Equal(t, float64(80), values["original_quota"])
}
