package openai

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	relayconstant "github.com/QuantumNous/new-api/relay/constant"
	"github.com/QuantumNous/new-api/relay/helper"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestGPTImage26JSONGenerationAndEditRelay(t *testing.T) {
	for _, mode := range []int{relayconstant.RelayModeImagesGenerations, relayconstant.RelayModeImagesEdits} {
		for _, format := range []string{"svg", "gif"} {
			t.Run(format+"/mode_"+strconv.Itoa(mode), func(t *testing.T) {
				body := `{"model":"public-image","prompt":"animate this logo","n":1,"complexity":"low","svg":{"model":"vector"},"markup":{"image_url":"https://example.com/layer.png"},"frame_count":2,"frame_duration_ms":0,"output_format":"` + format + `","image":"https://example.com/source.png","private_extension":"must-not-leak"}`
				c, _ := gin.CreateTestContext(httptest.NewRecorder())
				path := "/v1/images/generations"
				if mode == relayconstant.RelayModeImagesEdits {
					path = "/v1/images/edits"
				}
				c.Request = httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
				c.Request.Header.Set("Content-Type", "application/json")
				request, err := helper.GetAndValidOpenAIImageRequest(c, mode)
				require.NoError(t, err)
				// ImageHelper deep-copies and maps before adaptor conversion.
				copied, err := common.DeepCopy(request)
				require.NoError(t, err)
				c.Set("model_mapping", `{"public-image":"gpt-image-2.6-flare"}`)
				info := &relaycommon.RelayInfo{RelayMode: mode, OriginModelName: "public-image", ChannelMeta: &relaycommon.ChannelMeta{UpstreamModelName: "public-image"}}
				require.NoError(t, helper.ModelMappedHelper(c, info, copied))
				converted, err := (&Adaptor{}).ConvertImageRequest(c, info, *copied)
				require.NoError(t, err)
				encoded, err := common.Marshal(converted)
				require.NoError(t, err)
				var payload map[string]any
				require.NoError(t, common.Unmarshal(encoded, &payload))
				require.Equal(t, "gpt-image-2.6-flare", payload["model"])
				require.Equal(t, "low", payload["complexity"])
				require.Equal(t, map[string]any{"model": "vector"}, payload["svg"])
				require.Equal(t, map[string]any{"image_url": "https://example.com/layer.png"}, payload["markup"])
				require.Equal(t, float64(2), payload["frame_count"])
				require.Equal(t, float64(0), payload["frame_duration_ms"])
				require.Equal(t, format, payload["output_format"])
				require.Equal(t, float64(1), payload["n"])
				require.NotContains(t, payload, "private_extension")
				count, err := copied.ImageCount(false)
				require.NoError(t, err)
				require.Equal(t, 1, count) // Frames are not a new local billing multiplier.
			})
		}
	}
}

func TestGPTImage26MultipartMarkupAndParameters(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	fields := map[string]string{
		"model": "gpt-image-2.6-flare", "prompt": "edit the marked area", "group": "internal-group",
		"complexity": "low", "svg": "false", "frame_count": "2", "frame_duration_ms": "0", "output_format": "gif",
	}
	for key, value := range fields {
		require.NoError(t, writer.WriteField(key, value))
	}
	for _, name := range []string{"image", "mask", "markup"} {
		part, err := writer.CreateFormFile(name, name+".png")
		require.NoError(t, err)
		_, err = part.Write([]byte("original " + name))
		require.NoError(t, err)
	}
	require.NoError(t, writer.Close())
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/images/edits", &body)
	c.Request.Header.Set("Content-Type", writer.FormDataContentType())
	request, err := helper.GetAndValidOpenAIImageRequest(c, relayconstant.RelayModeImagesEdits)
	require.NoError(t, err)
	require.JSONEq(t, `"low"`, string(request.Complexity))
	require.JSONEq(t, `false`, string(request.Svg))
	require.JSONEq(t, `2`, string(request.FrameCount))
	require.JSONEq(t, `0`, string(request.FrameDurationMs))
	// Retries can remap the public model without dropping the annotation files.
	request.Model = "mapped-image-model"
	converted, err := (&Adaptor{}).ConvertImageRequest(c, &relaycommon.RelayInfo{RelayMode: relayconstant.RelayModeImagesEdits}, *request)
	require.NoError(t, err)
	upstream := httptest.NewRequest(http.MethodPost, "/v1/images/edits", converted.(*bytes.Buffer))
	upstream.Header.Set("Content-Type", c.Request.Header.Get("Content-Type"))
	require.NoError(t, upstream.ParseMultipartForm(32<<20))
	defer upstream.MultipartForm.RemoveAll()
	for key, value := range fields {
		if key == "model" || key == "group" {
			continue
		}
		require.Equal(t, value, upstream.PostForm.Get(key))
	}
	require.Equal(t, "mapped-image-model", upstream.PostForm.Get("model"))
	require.Empty(t, upstream.PostForm.Get("group"))
	for _, name := range []string{"image", "mask", "markup"} {
		require.Len(t, upstream.MultipartForm.File[name], 1)
		part := upstream.MultipartForm.File[name][0]
		require.Equal(t, "image/png", part.Header.Get("Content-Type"))
		file, err := part.Open()
		require.NoError(t, err)
		data, err := io.ReadAll(file)
		_ = file.Close()
		require.NoError(t, err)
		require.Equal(t, "original "+name, string(data))
	}
}

func TestGPTImage26OutputFormatsKeepResponseAndUsage(t *testing.T) {
	for _, format := range []string{"svg", "gif"} {
		mimeType := "image/gif"
		if format == "svg" {
			mimeType = "image/svg+xml"
		}
		body := `{"model":"gpt-image-2.6","data":[{"url":"https://example.com/result.` + format + `","b64_json":"untouched-base64","mime_type":"` + mimeType + `"}],"usage":{"input_tokens":10,"output_tokens":20,"total_tokens":30}}`
		c, recorder, response, info := newImageTestContext(t, body, "application/json", false)
		usage, err := OpenaiImageHandler(c, info, response)
		require.Nil(t, err)
		require.Equal(t, body, recorder.Body.String())
		require.Equal(t, 10, usage.PromptTokens)
		require.Equal(t, 20, usage.CompletionTokens)
		require.Equal(t, 30, usage.TotalTokens)
	}
}
