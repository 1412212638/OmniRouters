package system_setting

import "testing"

func TestValidateCORSMode(t *testing.T) {
	for _, mode := range []string{CORSModeDisabled, CORSModeAllowlist, CORSModeAll} {
		if err := ValidateCORSMode(mode); err != nil {
			t.Fatalf("expected mode %q to be valid: %v", mode, err)
		}
	}
	if err := ValidateCORSMode("wildcard"); err == nil {
		t.Fatal("expected unknown CORS mode to be rejected")
	}
}

func TestValidateCORSOrigins(t *testing.T) {
	if err := ValidateCORSOrigins(`["https://app.example.com","http://localhost:3000/"]`); err != nil {
		t.Fatalf("expected valid origins: %v", err)
	}
	for _, value := range []string{
		`["https://app.example.com/path"]`,
		`["https://app.example.com?x=1"]`,
		`["javascript:alert(1)"]`,
		`["https://user:pass@app.example.com"]`,
		`{"origin":"https://app.example.com"}`,
	} {
		if err := ValidateCORSOrigins(value); err == nil {
			t.Fatalf("expected origins %s to be rejected", value)
		}
	}
}
