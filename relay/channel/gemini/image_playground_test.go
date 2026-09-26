package gemini

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestConvertGeminiImageModelRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/pg/images/generations", nil)
	info := &relaycommon.RelayInfo{ChannelMeta: &relaycommon.ChannelMeta{
		UpstreamModelName: "gemini-3.1-flash-lite-image",
	}}

	converted, err := (&Adaptor{}).ConvertImageRequest(c, info, dto.ImageRequest{
		Prompt:  "a red fox",
		Size:    "16:9",
		Quality: "high",
	})
	require.NoError(t, err)
	request, ok := converted.(*dto.GeminiChatRequest)
	require.True(t, ok)
	require.Equal(t, "a red fox", request.Contents[0].Parts[0].Text)
	require.Equal(t, []string{"TEXT", "IMAGE"}, request.GenerationConfig.ResponseModalities)
	var imageConfig map[string]string
	require.NoError(t, common.Unmarshal(request.GenerationConfig.ImageConfig, &imageConfig))
	require.Equal(t, "16:9", imageConfig["aspectRatio"])
	require.Equal(t, "2K", imageConfig["imageSize"])
}

func TestConvertGeminiImageEditRequestAddsInlineDataContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/pg/images/edits", nil)
	info := &relaycommon.RelayInfo{ChannelMeta: &relaycommon.ChannelMeta{
		UpstreamModelName: "gemini-3.1-flash-lite-image",
	}}

	converted, err := (&Adaptor{}).ConvertImageRequest(c, info, dto.ImageRequest{
		Prompt: "make the sky blue",
		Image:  []byte(`"data:image/png;base64,aGVsbG8="`),
	})
	require.NoError(t, err)
	request, ok := converted.(*dto.GeminiChatRequest)
	require.True(t, ok)
	require.Len(t, request.Contents[0].Parts, 2)
	require.Equal(t, "image/png", request.Contents[0].Parts[0].InlineData.MimeType)
	require.Equal(t, "aGVsbG8=", request.Contents[0].Parts[0].InlineData.Data)
	require.Equal(t, "make the sky blue", request.Contents[0].Parts[1].Text)
}

func TestGeminiGenerateContentImageHandlerReturnsOpenAIImageResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/pg/images/generations", nil)
	responseBody := `{"candidates":[{"content":{"role":"model","parts":[{"text":"Here is the image"},{"inlineData":{"mimeType":"image/png","data":"aW1hZ2UtYnl0ZXM="}}]}}]}`
	resp := &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(responseBody))}

	usage, apiErr := GeminiGenerateContentImageHandler(c, &relaycommon.RelayInfo{}, resp)
	require.Nil(t, apiErr)
	require.Equal(t, 258, usage.TotalTokens)
	var imageResponse dto.ImageResponse
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &imageResponse))
	require.Len(t, imageResponse.Data, 1)
	require.Equal(t, "aW1hZ2UtYnl0ZXM=", imageResponse.Data[0].B64Json)
	require.Equal(t, "image/png", imageResponse.Data[0].MimeType)
}
