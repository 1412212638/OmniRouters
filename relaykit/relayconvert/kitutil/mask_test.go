package kitutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaskSensitiveInfoMasksOnlyConfiguredUpstreamAddress(t *testing.T) {
	message := "request to https://private.gateway.example/v1/chat failed; see https://docs.provider.com/errors?code=123"
	masked := MaskSensitiveInfo(message, "https://private.gateway.example")
	assert.NotContains(t, masked, "private.gateway.example")
	assert.Contains(t, masked, "https://docs.provider.com/errors?code=***")
	assert.Contains(t, masked, "https://***.example/***")
}

func TestMaskSensitiveInfoPreservesConfiguredEndpointPathAndMasksMatchingPort(t *testing.T) {
	message := "request failed at https://private.gateway.example:8443/v1/chat; docs https://docs.provider.com/errors"
	masked := MaskSensitiveInfo(message, "https://private.gateway.example:8443/path?key=secret")
	assert.NotContains(t, masked, "private.gateway.example")
	assert.Contains(t, masked, ":8443/***")
	assert.Contains(t, masked, "https://docs.provider.com/errors")
}
