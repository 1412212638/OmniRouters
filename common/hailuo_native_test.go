package common_test

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/stretchr/testify/require"
)

func TestH3NativeEndpointDefaults(t *testing.T) {
	for _, tc := range []struct {
		endpoint constant.EndpointType
		path     string
		method   string
	}{
		{constant.EndpointTypeMiniMaxH3Create, "/v2/video_generation", "POST"},
		{constant.EndpointTypeMiniMaxH3Query, "/v2/query/video_generation/{task_id}", "GET"},
	} {
		info, ok := common.GetDefaultEndpointInfo(tc.endpoint)
		require.True(t, ok)
		require.Equal(t, common.EndpointInfo{Path: tc.path, Method: tc.method}, info)
		require.False(t, common.IsOpenAIChatEndpointTypes([]constant.EndpointType{constant.EndpointTypeOpenAI, tc.endpoint}))
	}
}
