package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProxyClientCacheReusesCanonicalURL(t *testing.T) {
	ResetProxyClientCache()
	t.Cleanup(ResetProxyClientCache)

	first, err := GetHttpClientWithProxy(" HTTP://proxy.example:8080/ ")
	require.NoError(t, err)
	second, err := GetHttpClientWithProxy("http://proxy.example:8080")
	require.NoError(t, err)

	assert.Same(t, first, second)
}

func TestInvalidateProxyClientOnlyRemovesMatchingProxy(t *testing.T) {
	ResetProxyClientCache()
	t.Cleanup(ResetProxyClientCache)

	first, err := GetHttpClientWithProxy("http://first.example:8080")
	require.NoError(t, err)
	second, err := GetHttpClientWithProxy("http://second.example:8080")
	require.NoError(t, err)

	InvalidateProxyClient("http://first.example:8080/")

	firstAfterInvalidation, err := GetHttpClientWithProxy("http://first.example:8080")
	require.NoError(t, err)
	secondAfterInvalidation, err := GetHttpClientWithProxy("http://second.example:8080")
	require.NoError(t, err)
	assert.NotSame(t, first, firstAfterInvalidation)
	assert.Same(t, second, secondAfterInvalidation)
}
