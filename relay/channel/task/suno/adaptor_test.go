package suno

import (
	"testing"

	"github.com/QuantumNous/new-api/constant"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
)

func TestBuildRequestURLUsesLowercaseAction(t *testing.T) {
	adaptor := &TaskAdaptor{}
	info := &relaycommon.RelayInfo{
		ChannelBaseUrl: "https://api.example.com",
		TaskRelayInfo: &relaycommon.TaskRelayInfo{
			Action: constant.SunoActionMusic,
		},
	}

	got, err := adaptor.BuildRequestURL(info)
	if err != nil {
		t.Fatalf("BuildRequestURL returned error: %v", err)
	}

	want := "https://api.example.com/suno/submit/music"
	if got != want {
		t.Fatalf("BuildRequestURL = %q, want %q", got, want)
	}
}
