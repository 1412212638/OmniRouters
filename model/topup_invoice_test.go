package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
)

func TestMarkTopUpInvoiceAvailability(t *testing.T) {
	tests := []struct {
		name     string
		topUp    *TopUp
		wantShow bool
	}{
		{
			name: "stripe invoice",
			topUp: &TopUp{
				PaymentProvider: PaymentProviderStripe,
				StripeInvoiceID: "in_123",
				Status:          common.TopUpStatusSuccess,
			},
			wantShow: true,
		},
		{
			name: "legacy stripe without invoice",
			topUp: &TopUp{
				PaymentMethod: PaymentMethodStripe,
				Status:        common.TopUpStatusSuccess,
			},
		},
		{
			name: "legacy Pancake provider",
			topUp: &TopUp{
				PaymentMethod: PaymentMethodWaffoPancake,
				Status:        common.TopUpStatusSuccess,
			},
			wantShow: true,
		},
		{
			name: "pending Pancake order",
			topUp: &TopUp{
				PaymentProvider: PaymentProviderWaffoPancake,
				Status:          common.TopUpStatusPending,
			},
		},
		{
			name: "other provider",
			topUp: &TopUp{
				PaymentProvider: PaymentProviderEpay,
				Status:          common.TopUpStatusSuccess,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			markTopUpInvoiceAvailability([]*TopUp{tt.topUp})
			if tt.topUp.InvoiceAvailable != tt.wantShow {
				t.Errorf("InvoiceAvailable = %t, want %t", tt.topUp.InvoiceAvailable, tt.wantShow)
			}
		})
	}
}
