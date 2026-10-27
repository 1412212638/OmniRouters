package jsplugin

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFilteredClientRequestHeaders(t *testing.T) {
	headers := make(http.Header)
	headers.Set("Content-Type", "application/json")
	headers.Set("TypeSafe-Danger-Zone-Long-Context", "true")
	headers.Add("X-Trace", "first")
	headers.Add("X-Trace", "second")
	for _, name := range []string{"Authorization", "Proxy-Authorization", "Cookie", "X-Api-Key", "X-Goog-Api-Key", "Mj-Api-Secret"} {
		headers.Set(name, "gateway-secret")
	}

	got := filteredClientRequestHeaders(headers)
	require.Equal(t, "application/json", got["Content-Type"])
	require.Equal(t, "", got["Accept"])
	require.Equal(t, "true", got["Typesafe-Danger-Zone-Long-Context"])
	require.Equal(t, "first", got["X-Trace"])
	for _, name := range []string{"Authorization", "Proxy-Authorization", "Cookie", "X-Api-Key", "X-Goog-Api-Key", "Mj-Api-Secret"} {
		_, exists := got[name]
		require.False(t, exists, name)
	}
}
