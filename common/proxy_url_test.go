package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseProxyURLStrict(t *testing.T) {
	tests := []struct {
		name       string
		raw        string
		want       string
		wantErr    string
		wantLegacy bool
	}{
		{name: "empty"},
		{name: "http", raw: " HTTP://proxy.example:8080/ ", want: "http://proxy.example:8080"},
		{name: "socks default port", raw: "socks5://user:pass@proxy.example", want: "socks5://user:pass@proxy.example:1080"},
		{name: "unsupported scheme", raw: "ftp://proxy.example", wantErr: "must use"},
		{name: "missing host", raw: "socks5:///path", wantErr: "include a host"},
		{name: "invalid port", raw: "http://proxy.example:70000", wantErr: "valid port"},
		{name: "path", raw: "http://proxy.example/path", wantErr: "must not include a path"},
		{name: "query", raw: "http://proxy.example/?mode=legacy", wantErr: "must not include a query"},
		{name: "fragment", raw: "http://proxy.example/#legacy", wantErr: "must not include a fragment"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := ParseProxyURLStrict(tt.raw)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			if tt.want == "" {
				assert.Nil(t, parsed)
				return
			}
			assert.Equal(t, tt.want, parsed.String())
		})
	}
}

func TestParseProxyURLRuntimeStripsLegacySuffix(t *testing.T) {
	parsed, stripped, err := ParseProxyURLRuntime("socks5h://proxy.example/legacy?mode=1#old")

	require.NoError(t, err)
	assert.True(t, stripped)
	assert.Equal(t, "socks5h://proxy.example:1080", parsed.String())
}
