package dto_test

import (
	"encoding/json"
	"testing"

	"github.com/QuantumNous/new-api/common"
	relaydto "github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestMessageToolCallsPreserveExtraContentThroughParseAndSet(t *testing.T) {
	msg := relaydto.Message{
		Role: "assistant",
		ToolCalls: json.RawMessage(`[
			{
				"id":"call_123",
				"type":"function",
				"function":{"name":"lookup","arguments":"{\"city\":\"Shanghai\"}"},
				"extra_content":{"google":{"thought_signature":"sig-123"}}
			}
		]`),
	}

	toolCalls := msg.ParseToolCalls()
	require.Len(t, toolCalls, 1)
	require.Equal(t, "sig-123", gjson.GetBytes(toolCalls[0].ExtraContent, "google.thought_signature").String())

	msg.SetToolCalls(toolCalls)
	require.Equal(t, "sig-123", gjson.GetBytes(msg.ToolCalls, "0.extra_content.google.thought_signature").String())
}

func TestGeneralOpenAIRequestDeepCopyPreservesToolCallExtraContent(t *testing.T) {
	request := relaydto.GeneralOpenAIRequest{
		Model: "gemini-3.1-pro-preview",
		Messages: []relaydto.Message{
			{
				Role: "assistant",
				ToolCalls: json.RawMessage(`[
					{
						"id":"call_123",
						"type":"function",
						"function":{"name":"lookup","arguments":"{}"},
						"extra_content":{"google":{"thought_signature":"sig-123"}}
					}
				]`),
			},
		},
	}

	copiedRequest, err := common.DeepCopy(&request)
	require.NoError(t, err)
	require.Equal(t, "sig-123", gjson.GetBytes(copiedRequest.Messages[0].ToolCalls, "0.extra_content.google.thought_signature").String())
}

func TestStreamToolCallResponsePreservesExtraContent(t *testing.T) {
	var response relaydto.ChatCompletionsStreamResponse
	err := common.Unmarshal([]byte(`{
		"choices":[
			{
				"index":0,
				"delta":{
					"tool_calls":[
						{
							"index":0,
							"id":"call_123",
							"type":"function",
							"function":{"name":"lookup","arguments":"{}"},
							"extra_content":{"google":{"thought_signature":"sig-123"}}
						}
					]
				}
			}
		]
	}`), &response)
	require.NoError(t, err)

	out, err := common.Marshal(response)
	require.NoError(t, err)
	require.Equal(t, "sig-123", gjson.GetBytes(out, "choices.0.delta.tool_calls.0.extra_content.google.thought_signature").String())
}
