package tencent

import (
	"testing"

	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/relay/channel/openai"
	relaycommon "github.com/QuantumNous/new-api/relay/common"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDispatchAdaptorInit(t *testing.T) {
	tests := []struct {
		name        string
		apiKey      string
		baseURL     string
		wantTC3     bool
		wantBaseURL string
	}{
		{"legacy key uses TC3", "1300000000|AKIDxxxxxxxx|secretxxxxxxxx", constant.ChannelBaseURLs[constant.ChannelTypeTencent], true, constant.ChannelBaseURLs[constant.ChannelTypeTencent]},
		{"TokenHub key rewrites default URL", "sk-xxxxxxxxxxxxxxxx", constant.ChannelBaseURLs[constant.ChannelTypeTencent], false, tokenHubBaseURL},
		{"TokenHub key rewrites empty URL", "sk-xxxxxxxxxxxxxxxx", "", false, tokenHubBaseURL},
		{"TokenHub key preserves custom URL", "sk-xxxxxxxxxxxxxxxx", "https://proxy.example.com", false, "https://proxy.example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &relaycommon.RelayInfo{ChannelMeta: &relaycommon.ChannelMeta{
				ChannelType:    constant.ChannelTypeTencent,
				ApiKey:         tt.apiKey,
				ChannelBaseUrl: tt.baseURL,
			}}

			dispatch := &DispatchAdaptor{}
			dispatch.Init(info)

			require.NotNil(t, dispatch.Adaptor)
			if tt.wantTC3 {
				assert.IsType(t, &Adaptor{}, dispatch.Adaptor)
			} else {
				assert.IsType(t, &openai.Adaptor{}, dispatch.Adaptor)
			}
			assert.Equal(t, tt.wantBaseURL, info.ChannelBaseUrl)
		})
	}
}
