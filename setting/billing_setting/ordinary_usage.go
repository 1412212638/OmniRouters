package billing_setting

import (
	"fmt"
	"strings"

	"github.com/QuantumNous/new-api/pkg/billingexpr"
	"github.com/QuantumNous/new-api/pkg/jsplugin"
)

// OrdinaryResponseUsageSchema returns the bounded usage fields exposed to an
// ordinary OpenAI-compatible response profile. The schema is deliberately
// separate from task-plugin metadata: the same model may have different
// pricing expressions on the two channel types.
func OrdinaryResponseUsageSchema(model string) (map[string]jsplugin.UsageFieldSchema, bool) {
	if CanonicalOrdinaryResponseUsageModel(model) == "" {
		return nil, false
	}
	return map[string]jsplugin.UsageFieldSchema{
		"images_up_to_1_5k":   {Type: "number", Unit: "count"},
		"images_above_1_5k":   {Type: "number", Unit: "count"},
		"generated_images":    {Type: "number", Unit: "count"},
		"input_images":        {Type: "number", Unit: "count"},
		"layer_decomposition": {Type: "boolean"},
	}, true
}

// CanonicalOrdinaryResponseUsageModel identifies models whose ordinary
// OpenAI-compatible response contains the Seedream image usage profile.
func CanonicalOrdinaryResponseUsageModel(model string) string {
	switch strings.ToLower(strings.TrimSpace(model)) {
	case "doubao-seedream-5.0-pro", "doubao-seedream-5-0-pro-260628":
		return "doubao-seedream-5.0-pro"
	case "doubao-seedream-5.0-flash", "doubao-seedream-5-0-flash-260915":
		return "doubao-seedream-5.0-flash"
	default:
		return ""
	}
}

// ValidateOrdinaryResponseUsageExpr checks literal u() fields against the
// selected ordinary response profile. Expressions without u() remain valid
// ordinary token/request expressions and are handled by the existing checks.
func ValidateOrdinaryResponseUsageExpr(model, expression string) error {
	schema, ok := OrdinaryResponseUsageSchema(model)
	if !ok {
		return fmt.Errorf("model %s has no ordinary response usage profile", model)
	}
	if _, err := billingexpr.CompileFromCache(expression); err != nil {
		return err
	}
	for key := range billingexpr.UsedUsageKeys(expression) {
		if _, exists := schema[key]; !exists {
			return fmt.Errorf("usage key %q is not declared by the ordinary response profile", key)
		}
	}
	return nil
}

// OrdinaryResponseUsageExprCompatibleWithTaskSchema reports whether the
// literal usage fields in an ordinary response expression have the same
// declared type and unit in a task plugin schema. An ordinary expression can
// be inherited by a task plugin only when both sides measure every referenced
// field identically.
func OrdinaryResponseUsageExprCompatibleWithTaskSchema(model, expression string, taskSchema map[string]jsplugin.UsageFieldSchema) bool {
	ordinarySchema, ok := OrdinaryResponseUsageSchema(model)
	if !ok || ValidateOrdinaryResponseUsageExpr(model, expression) != nil || !TaskExprCompatible(expression, taskSchema) {
		return false
	}
	for key := range billingexpr.UsedUsageKeys(expression) {
		ordinaryField, ordinaryExists := ordinarySchema[key]
		taskField, taskExists := taskSchema[key]
		if !ordinaryExists || !taskExists || !usageFieldSchemaCompatible(ordinaryField, taskField) {
			return false
		}
	}
	return true
}

// UsageExprCompatibleWithTaskSchema applies the ordinary response profile's
// type/unit compatibility rules when the model has one, while retaining the
// existing task-schema check for all other models.
func UsageExprCompatibleWithTaskSchema(model, expression string, taskSchema map[string]jsplugin.UsageFieldSchema) bool {
	if len(billingexpr.UsedUsageKeys(expression)) > 0 {
		if _, ordinary := OrdinaryResponseUsageSchema(model); ordinary {
			return OrdinaryResponseUsageExprCompatibleWithTaskSchema(model, expression, taskSchema)
		}
	}
	return TaskExprCompatible(expression, taskSchema)
}

func usageFieldSchemaCompatible(left, right jsplugin.UsageFieldSchema) bool {
	if left.Type != right.Type || left.Unit != right.Unit {
		return false
	}
	if len(left.Enum) != len(right.Enum) {
		return false
	}
	if len(left.Enum) > 0 {
		values := make(map[string]struct{}, len(left.Enum))
		for _, value := range left.Enum {
			values[value] = struct{}{}
		}
		for _, value := range right.Enum {
			if _, exists := values[value]; !exists {
				return false
			}
		}
	}
	return true
}
