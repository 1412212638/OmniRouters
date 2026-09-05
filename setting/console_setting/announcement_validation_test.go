package console_setting

import (
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
)

func TestAnnouncementContentLength(t *testing.T) {
	for _, tc := range []struct {
		name    string
		content string
		wantErr bool
	}{
		{"above old limit", strings.Repeat("a", 501), false},
		{"at limit", strings.Repeat("a", 5000), false},
		{"above limit", strings.Repeat("a", 5001), true},
		{"Chinese at limit", strings.Repeat("\u4e2d", 5000), false},
		{"UTF16 at limit", strings.Repeat("\U0001f600", 2500), false},
		{"UTF16 above limit", strings.Repeat("\U0001f600", 2500) + "a", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data, err := common.Marshal([]map[string]string{{
				"content": tc.content, "publishDate": "2026-09-05T12:00:00Z", "type": "default",
			}})
			if err != nil {
				t.Fatal(err)
			}
			err = ValidateConsoleSettings(string(data), "Announcements")
			if (err != nil) != tc.wantErr {
				t.Fatalf("validation error = %v, want error = %v", err, tc.wantErr)
			}
			if tc.wantErr && !strings.Contains(err.Error(), "5000") {
				t.Fatalf("error does not report the new limit: %v", err)
			}
		})
	}
}
