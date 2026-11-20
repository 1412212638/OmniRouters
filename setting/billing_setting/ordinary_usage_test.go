package billing_setting

import (
	"testing"

	"github.com/QuantumNous/new-api/pkg/jsplugin"
	"github.com/stretchr/testify/require"
)

func TestOrdinaryResponseUsageProfileIsModelScoped(t *testing.T) {
	require.Equal(t, "doubao-seedream-5.0-pro", CanonicalOrdinaryResponseUsageModel("doubao-seedream-5-0-pro-260628"))
	require.Equal(t, "doubao-seedream-5.5", CanonicalOrdinaryResponseUsageModel("doubao-seedream-5-5-261028"))
	require.Equal(t, "doubao-seedream-5.5", CanonicalOrdinaryResponseUsageModel("DOUBAO-SEEDREAM-5.5"))
	require.Empty(t, CanonicalOrdinaryResponseUsageModel("minimax-h3"))
	require.NoError(t, ValidateOrdinaryResponseUsageExpr("doubao-seedream-5.0-pro", `u("generated_images") * 0.2`))
	require.NoError(t, ValidateOrdinaryResponseUsageExpr("doubao-seedream-5.5", `u("images_up_to_1_5k") * 0.1 + u("images_above_1_5k") * 0.2 + max(u("input_images") - 1, 0) * 0.02`))
	require.Error(t, ValidateOrdinaryResponseUsageExpr("doubao-seedream-5.0-pro", `u("unknown") * 0.2`))
	require.Error(t, ValidateOrdinaryResponseUsageExpr("minimax-h3", `u("generated_images") * 0.2`))
}

func TestOrdinaryResponseUsageExpressionSharesOnlyCompatibleTaskSchema(t *testing.T) {
	compatible := map[string]jsplugin.UsageFieldSchema{
		"generated_images": {Type: "number", Unit: "count"},
	}
	incompatibleUnit := map[string]jsplugin.UsageFieldSchema{
		"generated_images": {Type: "number", Unit: "second"},
	}
	incompatibleType := map[string]jsplugin.UsageFieldSchema{
		"generated_images": {Type: "boolean"},
	}
	expression := `tier("base", u("generated_images") * 0.2)`

	require.True(t, OrdinaryResponseUsageExprCompatibleWithTaskSchema("doubao-seedream-5.0-pro", expression, compatible))
	require.True(t, UsageExprCompatibleWithTaskSchema("doubao-seedream-5.0-pro", expression, compatible))
	require.False(t, OrdinaryResponseUsageExprCompatibleWithTaskSchema("doubao-seedream-5.0-pro", expression, incompatibleUnit))
	require.False(t, OrdinaryResponseUsageExprCompatibleWithTaskSchema("doubao-seedream-5.0-pro", expression, incompatibleType))
	require.False(t, OrdinaryResponseUsageExprCompatibleWithTaskSchema("doubao-seedream-5.0-pro", `tier("base", u("input_images") * 0.2)`, compatible))
	require.True(t, UsageExprCompatibleWithTaskSchema("minimax-h3", `tier("base", u("generated_images") * 0.2)`, compatible))
}
