package sora

import (
	"math"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/types"

	"github.com/gin-gonic/gin"
)

func TestIsSoraAudioGenerationEnabled(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  bool
	}{
		{name: "bool true", value: true, want: true},
		{name: "enabled string", value: "Enabled", want: true},
		{name: "true string", value: "true", want: true},
		{name: "typo string", value: "ture", want: true},
		{name: "numeric one", value: float64(1), want: true},
		{name: "disabled string", value: "disabled", want: false},
		{name: "false string", value: "false", want: false},
		{name: "nil", value: nil, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSoraAudioGenerationEnabled(tt.value); got != tt.want {
				t.Fatalf("isSoraAudioGenerationEnabled(%v) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestEstimateBillingRejectsAudioSurchargeOverflow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)
	common.SetContextKey(c, constant.ContextKeySoraPricingContext, soraPricingContext{
		Seconds:                  1,
		Multiplier:               1,
		AudioGeneration:          true,
		AudioGenerationSurcharge: math.MaxFloat64,
	})
	info := &relaycommon.RelayInfo{
		PriceData: types.PriceData{
			GroupRatioInfo: types.GroupRatioInfo{GroupRatio: 1},
		},
	}

	ratios := (&TaskAdaptor{}).EstimateBilling(c, info)

	if ratios["seconds"] != 1 || ratios["resolution"] != 1 {
		t.Fatalf("unexpected Sora ratios: %#v", ratios)
	}
	if info.QuotaClamp == nil || info.QuotaClamp.Kind != common.QuotaClampOverflow {
		t.Fatalf("expected overflow clamp, got %#v", info.QuotaClamp)
	}
	if info.PriceData.FixedQuotaTotal() != 0 {
		t.Fatalf("overflowed audio surcharge must not become a fixed charge")
	}
}
