package dto

import "testing"

func TestValidateGPTImageParameters(t *testing.T) {
	tests := []struct {
		name    string
		request ImageRequest
		wantErr bool
	}{
		{
			name:    "valid maximum count and quality",
			request: ImageRequest{Model: "gpt-image-2.5-sunburst", N: pointer(uint(10)), Quality: "max", Size: "3840x2160"},
		},
		{
			name:    "valid inclusive aspect ratio boundary",
			request: ImageRequest{Model: "gpt-image-2.5-sunburst", Size: "3840x1280"},
		},
		{
			name:    "valid minimum pixel boundary",
			request: ImageRequest{Model: "gpt-image-2.5-sunburst", Size: "1024x640"},
		},
		{
			name:    "auto size",
			request: ImageRequest{Model: "gpt-image-2.5-flare", N: pointer(uint(1)), Quality: "xhigh", Size: "auto"},
		},
		{
			name:    "other image models keep their own rules",
			request: ImageRequest{Model: "gemini-3.1-flash-lite-image", N: pointer(uint(11)), Quality: "ultra", Size: "17x17"},
		},
		{
			name:    "count above GPT maximum",
			request: ImageRequest{Model: "gpt-image-2.5-sunburst", N: pointer(uint(11))},
			wantErr: true,
		},
		{
			name:    "zero count is outside GPT bounds",
			request: ImageRequest{Model: "gpt-image-2.5-sunburst", N: pointer(uint(0))},
			wantErr: true,
		},
		{
			name:    "unsupported quality",
			request: ImageRequest{Model: "gpt-image-2.5-sunburst", Quality: "ultra"},
			wantErr: true,
		},
		{
			name:    "dimensions must be multiples of sixteen",
			request: ImageRequest{Model: "gpt-image-2.5-sunburst", Size: "1000x1000"},
			wantErr: true,
		},
		{
			name:    "dimensions must meet pixel minimum",
			request: ImageRequest{Model: "gpt-image-2.5-sunburst", Size: "512x512"},
			wantErr: true,
		},
		{
			name:    "dimensions must meet aspect ratio",
			request: ImageRequest{Model: "gpt-image-2.5-sunburst", Size: "3840x1264"},
			wantErr: true,
		},
		{
			name:    "dimensions cannot exceed maximum pixels",
			request: ImageRequest{Model: "gpt-image-2.5-sunburst", Size: "3840x2176"},
			wantErr: true,
		},
		{
			name:    "dimensions cannot exceed maximum side",
			request: ImageRequest{Model: "gpt-image-2.5-sunburst", Size: "3856x2160"},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.request.ValidateGPTImageParameters()
			if (err != nil) != test.wantErr {
				t.Fatalf("ValidateGPTImageParameters() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}

func pointer[T any](value T) *T {
	return &value
}
