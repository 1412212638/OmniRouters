package gemini

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/dto"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestResponseToolCallIncludesGoogleThoughtSignatureExtraContent(t *testing.T) {
	part := dto.GeminiPart{
		FunctionCall: &dto.FunctionCall{
			FunctionName: "lookup_weather",
			Arguments: map[string]interface{}{
				"city": "Shanghai",
			},
		},
		ThoughtSignature: json.RawMessage(`"sig-123"`),
	}

	call := getResponseToolCall(&part)
	require.NotNil(t, call)
	require.Equal(t, "sig-123", gjson.GetBytes(call.ExtraContent, "google.thought_signature").String())
}

func TestCovertOpenAI2GeminiRestoresGoogleThoughtSignatureExtraContent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	info := &relaycommon.RelayInfo{
		ChannelMeta: &relaycommon.ChannelMeta{
			ChannelType:       constant.ChannelTypeGemini,
			UpstreamModelName: "gemini-3.1-pro-preview",
		},
	}
	request := dto.GeneralOpenAIRequest{
		Model: "gemini-3.1-pro-preview",
		Messages: []dto.Message{
			{
				Role:    "user",
				Content: "weather?",
			},
			{
				Role:    "assistant",
				Content: "",
				ToolCalls: json.RawMessage(`[
					{
						"id":"call_123",
						"type":"function",
						"function":{"name":"lookup_weather","arguments":"{\"city\":\"Shanghai\"}"},
						"extra_content":{"google":{"thought_signature":"sig-123"}}
					}
				]`),
			},
			{
				Role:       "tool",
				ToolCallId: "call_123",
				Content:    `{"temperature":25}`,
			},
		},
	}

	geminiRequest, err := CovertOpenAI2Gemini(c, request, info)
	require.NoError(t, err)
	require.Len(t, geminiRequest.Contents, 3)

	modelPart := geminiRequest.Contents[1].Parts[0]
	require.NotNil(t, modelPart.FunctionCall)
	var signature string
	err = common.Unmarshal(modelPart.ThoughtSignature, &signature)
	require.NoError(t, err)
	require.Equal(t, "sig-123", signature)

	toolPart := geminiRequest.Contents[2].Parts[0]
	require.NotNil(t, toolPart.FunctionResponse)
	require.Equal(t, "lookup_weather", toolPart.FunctionResponse.Name)
}
