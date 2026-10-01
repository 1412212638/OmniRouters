package claudemessages

import (
	"testing"

	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/relaykit/relayconvert/convmeta"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClaudeMessagesRequestToOpenAIChatHoistsToolResultImages(t *testing.T) {
	const imageData = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJ"
	const dataURL = "data:image/png;base64," + imageData

	request, err := ClaudeMessagesRequestToOpenAIChat(dto.ClaudeRequest{
		Model: "claude-test",
		Messages: []dto.ClaudeMessage{
			{Role: "user", Content: "take a screenshot"},
			{Role: "assistant", Content: []dto.ClaudeMediaMessage{{
				Type: "tool_use", Id: "toolu_1", Name: "screenshot", Input: map[string]any{},
			}}},
			{Role: "user", Content: []dto.ClaudeMediaMessage{
				{Type: "tool_result", ToolUseId: "toolu_1", Content: []dto.ClaudeMediaMessage{
					{Type: "image", Source: &dto.ClaudeMessageSource{Type: "base64", MediaType: "image/png", Data: imageData}},
				}},
				{Type: "text", Text: stringPtr("what do you see?")},
			}},
		},
	}, &convmeta.Values{})
	require.NoError(t, err)
	require.Len(t, request.Messages, 4)

	tool := request.Messages[2]
	assert.Equal(t, "tool", tool.Role)
	assert.Equal(t, "[image]", tool.StringContent())
	assert.NotContains(t, tool.StringContent(), imageData)

	user := request.Messages[3]
	assert.Equal(t, "user", user.Role)
	parts := user.ParseContent()
	require.Len(t, parts, 2)
	assert.Equal(t, dto.ContentTypeImageURL, parts[0].Type)
	image := parts[0].GetImageMedia()
	require.NotNil(t, image)
	assert.Equal(t, dataURL, image.Url)
	assert.Equal(t, dto.ContentTypeText, parts[1].Type)
	assert.Equal(t, "what do you see?", parts[1].Text)
}

func stringPtr(value string) *string { return &value }
