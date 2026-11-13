package billingexpr

import "testing"

// The visual editor distinguishes an absent category from an explicitly free
// category. Keep zero-priced identifiers available for usage normalization.
func TestEditorExplicitZeroRetainsCategoryVariables(t *testing.T) {
	for _, variable := range []string{"cr", "cc", "cc1h", "cc12h", "img", "img_cr", "img_o", "ai", "ao"} {
		t.Run(variable, func(t *testing.T) {
			expression := `tier("base", p * 5 + c * 30 + ` + variable + ` * 0)`
			if !UsedVars(expression)[variable] {
				t.Fatalf("zero-price category %q must remain visible to usage normalization", variable)
			}
		})
	}
}
