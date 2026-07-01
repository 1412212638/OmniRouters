package doubao

import (
	"math"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	relaycommon "github.com/QuantumNous/new-api/relay/common"

	"github.com/gin-gonic/gin"
)

func TestGetVideoInputRatio(t *testing.T) {
	tests := []struct {
		name       string
		model      string
		resolution string
		hasVideo   bool
		want       float64
		wantOK     bool
	}{
		{
			name:       "standard baseline",
			model:      "doubao-seedance-2-0-260128",
			resolution: "720p",
			want:       1,
			wantOK:     true,
		},
		{
			name:       "standard baseline with video",
			model:      "doubao-seedance-2-0-260128",
			resolution: "720p",
			hasVideo:   true,
			want:       28.0 / 46.0,
			wantOK:     true,
		},
		{
			name:       "standard 1080p",
			model:      "doubao-seedance-2-0-260128",
			resolution: "1080p",
			want:       51.0 / 46.0,
			wantOK:     true,
		},
		{
			name:       "standard 1080p with video normalizes resolution",
			model:      "doubao-seedance-2-0-260128",
			resolution: " 1080P ",
			hasVideo:   true,
			want:       31.0 / 46.0,
			wantOK:     true,
		},
		{
			name:       "standard 4k",
			model:      "doubao-seedance-2-0-260128",
			resolution: "4k",
			want:       26.0 / 46.0,
			wantOK:     true,
		},
		{
			name:       "standard 4k with video",
			model:      "doubao-seedance-2-0-260128",
			resolution: "4K",
			hasVideo:   true,
			want:       16.0 / 46.0,
			wantOK:     true,
		},
		{
			name:       "fast baseline",
			model:      "doubao-seedance-2-0-fast-260128",
			resolution: "720p",
			want:       1,
			wantOK:     true,
		},
		{
			name:       "fast baseline with video",
			model:      "doubao-seedance-2-0-fast-260128",
			resolution: "720p",
			hasVideo:   true,
			want:       22.0 / 37.0,
			wantOK:     true,
		},
		{
			name:       "fast unsupported resolution falls back to baseline",
			model:      "doubao-seedance-2-0-fast-260128",
			resolution: "4k",
			hasVideo:   true,
			want:       1,
			wantOK:     true,
		},
		{
			name:       "unknown model",
			model:      "unknown-model",
			resolution: "1080p",
			want:       0,
			wantOK:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := GetVideoInputRatio(tt.model, tt.resolution, tt.hasVideo)
			if ok != tt.wantOK {
				t.Fatalf("GetVideoInputRatio() ok = %v, want %v", ok, tt.wantOK)
			}
			if math.Abs(got-tt.want) > 1e-12 {
				t.Fatalf("GetVideoInputRatio() = %.12f, want %.12f", got, tt.want)
			}
		})
	}
}

func TestEstimateBillingUsesResolutionAndVideoInput(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		metadata map[string]interface{}
		want     float64
		wantNone bool
	}{
		{
			name:     "baseline has no extra ratio",
			metadata: map[string]interface{}{"resolution": "720p"},
			wantNone: true,
		},
		{
			name:     "1080p output",
			metadata: map[string]interface{}{"resolution": "1080p"},
			want:     51.0 / 46.0,
		},
		{
			name: "4k output with video input",
			metadata: map[string]interface{}{
				"resolution": "4k",
				"content": []interface{}{
					map[string]interface{}{
						"type":      "video_url",
						"video_url": map[string]interface{}{"url": "https://example.com/input.mp4"},
					},
				},
			},
			want: 16.0 / 46.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Set("task_request", relaycommon.TaskSubmitReq{Metadata: tt.metadata})

			got := (&TaskAdaptor{}).EstimateBilling(c, &relaycommon.RelayInfo{
				OriginModelName: "doubao-seedance-2-0-260128",
			})
			if tt.wantNone {
				if len(got) != 0 {
					t.Fatalf("EstimateBilling() = %#v, want no extra ratio", got)
				}
				return
			}

			ratio, ok := got["video_input"]
			if !ok {
				t.Fatalf("EstimateBilling() = %#v, want video_input ratio", got)
			}
			if math.Abs(ratio-tt.want) > 1e-12 {
				t.Fatalf("EstimateBilling() ratio = %.12f, want %.12f", ratio, tt.want)
			}
		})
	}
}

func TestConvertToRequestPayloadPreservesExplicitZeroPriority(t *testing.T) {
	payload, err := (&TaskAdaptor{}).convertToRequestPayload(&relaycommon.TaskSubmitReq{
		Model:  "doubao-seedance-2-0-260128",
		Prompt: "   ",
		Metadata: map[string]interface{}{
			"priority":          0,
			"safety_identifier": "user-hash",
			"resolution":       "4k",
		},
	})
	if err != nil {
		t.Fatalf("convertToRequestPayload() error = %v", err)
	}
	if payload.Priority == nil || int(*payload.Priority) != 0 {
		t.Fatalf("priority = %#v, want explicit zero", payload.Priority)
	}
	if payload.SafetyIdentifier != "user-hash" {
		t.Fatalf("safety_identifier = %q, want user-hash", payload.SafetyIdentifier)
	}
	if len(payload.Content) != 0 {
		t.Fatalf("content = %#v, want empty content for blank prompt", payload.Content)
	}

	data, err := common.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	body := string(data)
	if !strings.Contains(body, `"priority":0`) {
		t.Fatalf("marshaled payload omitted explicit zero priority: %s", body)
	}
	if !strings.Contains(body, `"safety_identifier":"user-hash"`) {
		t.Fatalf("marshaled payload omitted safety_identifier: %s", body)
	}
}
