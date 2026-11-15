package jsplugin

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/require"
)

func TestDecodeUsageSchemaAcceptsFieldSpecificMaximum(t *testing.T) {
	schema, err := decodeUsageSchema(map[string]any{
		"characters": map[string]any{
			"type": "number",
			"unit": "count",
			"max": 10000.0,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, schema["characters"].Max)
	require.Equal(t, 10000.0, *schema["characters"].Max)
}

func TestDecodeUsageSchemaRejectsUnsafeFieldSpecificMaximum(t *testing.T) {
	_, err := decodeUsageSchema(map[string]any{
		"characters": map[string]any{
			"type": "number",
			"unit": "count",
			"max": float64(common.MaxQuota) + 1,
		},
	})
	require.ErrorContains(t, err, "max must be greater than 0")
}
