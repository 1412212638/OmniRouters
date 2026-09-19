package service

import (
	"testing"

	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/stretchr/testify/require"
)

func TestResponseModelLogIsAdminOnlyAndOmitsRedundantNames(t *testing.T) {
	info := &relaycommon.RelayInfo{OriginModelName: "public", ChannelMeta: &relaycommon.ChannelMeta{UpstreamModelName: "public"}}
	info.ObserveResponseModel("public")
	other := model.NewLogOther()
	AppendResponseModelLogInfo(info, other)
	require.Empty(t, other.Snapshot())
	info.ObserveResponseModel("private-provider-model")
	AppendResponseModelLogInfo(info, other)
	snapshot := other.Snapshot()
	require.NotContains(t, snapshot, "response_model")
	require.Contains(t, snapshot["admin_info"], "response_model")
	require.NotContains(t, other.JSONString(), "\"mismatch\"")
	info.ObserveResponseModel("public")
	require.Equal(t, "private-provider-model", info.ResponseModel.ReturnedModel)
}
