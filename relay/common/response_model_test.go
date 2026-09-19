package common

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestResponseModelObservationDoesNotChangeRouting(t *testing.T) {
	info := &RelayInfo{OriginModelName: "public-alias", ChannelMeta: &ChannelMeta{UpstreamModelName: "gpt-test"}}
	info.ObserveResponseModel("")
	require.Nil(t, info.ResponseModel)
	info.ObserveResponseModel("GPT-TEST-2026-09-20")
	require.False(t, info.ResponseModel.Mismatch())
	info.ObserveResponseModel("other-model")
	require.True(t, info.ResponseModel.Mismatch())
	info.ObserveResponseModel("gpt-test")
	require.Equal(t, "other-model", info.ResponseModel.ReturnedModel)
	require.Equal(t, "gpt-test", info.UpstreamModelName)
	require.Equal(t, "public-alias", info.OriginModelName)
	var absent *RelayInfo
	absent.ObserveResponseModel("model")
}

func TestResponseModelMatchesProviderPathsAndAliases(t *testing.T) {
	for _, returned := range []string{"public-alias", "GPT-TEST", "gpt-test-2026", "provider/gpt-test"} {
		observation := &ResponseModel{RequestedModel: "public-alias", UpstreamModel: "gpt-test", ReturnedModel: returned}
		require.False(t, observation.Mismatch(), returned)
	}
}

func TestResponseModelIsResetOnChannelRetry(t *testing.T) {
	info := &RelayInfo{OriginModelName: "public-alias", ChannelMeta: &ChannelMeta{UpstreamModelName: "first-model"}}
	info.ObserveResponseModel("unexpected-model")
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	info.InitChannelMeta(c)
	require.Nil(t, info.ResponseModel)
	info.ChannelMeta.UpstreamModelName = "retry-model"
	info.ObserveResponseModel("retry-model")
	require.False(t, info.ResponseModel.Mismatch())
	require.Equal(t, "retry-model", info.ResponseModel.UpstreamModel)
}
