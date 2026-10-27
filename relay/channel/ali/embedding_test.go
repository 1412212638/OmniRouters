package ali

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relay/constant"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/relaykit/types"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMultimodalEmbeddingModelRouting(t *testing.T) {
	tests := []struct {
		model string
		want  bool
	}{
		{"qwen3-vl-embedding", true},
		{"qwen2.5-vl-embedding-instruct", true},
		{"tongyi-embedding-vision-plus-2026-03-06", true},
		{"multimodal-embedding-v1", true},
		{"text-embedding-v4", false},
	}
	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			assert.Equal(t, tt.want, isMultimodalEmbeddingModel(tt.model))
		})
	}
}

func TestAliEmbeddingEndpointAndRequestConversion(t *testing.T) {
	const base = "https://dashscope.example"
	tests := []struct {
		name    string
		model   string
		input   string
		wantURL string
		want    string
		wantErr bool
	}{
		{
			name:    "native text-only input",
			model:   "qwen3-vl-embedding",
			input:   `"input":"a cat"`,
			wantURL: base + "/api/v1/services/embeddings/multimodal-embedding/multimodal-embedding",
			want:    `{"model":"qwen3-vl-embedding","input":{"contents":[{"text":"a cat"}]}}`,
		},
		{
			name:    "native mixed image and video",
			model:   "multimodal-embedding-v1",
			input:   `"input":[{"type":"text","text":"a cat"},{"type":"image_url","image_url":{"url":"data:image/png;base64,AAAA"}},{"type":"video_url","video_url":{"url":"https://example.test/cat.mp4"}}]`,
			wantURL: base + "/api/v1/services/embeddings/multimodal-embedding/multimodal-embedding",
			want:    `{"model":"multimodal-embedding-v1","input":{"contents":[{"text":"a cat"},{"image":"data:image/png;base64,AAAA"},{"video":"https://example.test/cat.mp4"}]}}`,
		},
		{
			name:    "text embedding stays compatible",
			model:   "text-embedding-v4",
			input:   `"input":"a cat","dimensions":1024`,
			wantURL: base + "/compatible-mode/v1/embeddings",
			want:    `{"model":"text-embedding-v4","input":"a cat","dimensions":1024}`,
		},
		{
			name:    "reject missing media url",
			model:   "qwen3-vl-embedding",
			input:   `"input":[{"type":"image_url","image_url":{"url":""}}]`,
			wantURL: base + "/api/v1/services/embeddings/multimodal-embedding/multimodal-embedding",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &relaycommon.RelayInfo{
				RelayMode: constant.RelayModeEmbeddings,
				ChannelMeta: &relaycommon.ChannelMeta{
					ChannelBaseUrl:   base,
					UpstreamModelName: tt.model,
				},
			}
			url, err := (&Adaptor{}).GetRequestURL(info)
			require.NoError(t, err)
			assert.Equal(t, tt.wantURL, url)

			var request dto.EmbeddingRequest
			require.NoError(t, common.UnmarshalJsonStr(`{"model":"`+tt.model+`",`+tt.input+`}`, &request))
			converted, err := (&Adaptor{}).ConvertEmbeddingRequest(nil, info, request)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			wire, err := common.Marshal(converted)
			require.NoError(t, err)
			assert.JSONEq(t, tt.want, string(wire))
		})
	}
}

func TestMultimodalEmbeddingHandlerUsageAndFallback(t *testing.T) {
	tests := []struct {
		name        string
		model       string
		body        string
		estimate    int
		wantUsage   dto.Usage
		wantErrCode bool
	}{
		{
			name:  "image tokens included, output tokens excluded",
			model: "qwen3-vl-embedding",
			body: `{"output":{"embeddings":[{"index":0,"embedding":[0.1,-0.2]}]},` +
				`"usage":{"input_tokens":43,"image_tokens":1247,"output_tokens":4,"total_tokens":1294}}`,
			wantUsage: dto.Usage{PromptTokens: 1290, TotalTokens: 1290,
				PromptTokensDetails: dto.InputTokenDetails{TextTokens: 43, ImageTokens: 1247}},
		},
		{
			name:  "vision model input detail split",
			model: "tongyi-embedding-vision-plus",
			body: `{"output":{"embeddings":[{"index":0,"embedding":[0.5]}]},` +
				`"usage":{"input_tokens":903,"input_tokens_details":{"text_tokens":7,"image_tokens":896},"output_tokens":3,"total_tokens":906}}`,
			wantUsage: dto.Usage{PromptTokens: 903, TotalTokens: 903,
				PromptTokensDetails: dto.InputTokenDetails{TextTokens: 7, ImageTokens: 896}},
		},
		{
			name:      "missing usage uses request estimate",
			model:     "qwen3-vl-embedding",
			body:      `{"output":{"embeddings":[{"index":0,"embedding":[0.5]}]}}`,
			estimate:  12,
			wantUsage: dto.Usage{PromptTokens: 12, TotalTokens: 12},
		},
		{
			name:        "error returned in a successful http response",
			model:       "qwen3-vl-embedding",
			body:        `{"code":"InvalidParameter","message":"The input is invalid.","request_id":"r4"}`,
			wantErrCode: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			info := &relaycommon.RelayInfo{
				RelayMode: constant.RelayModeEmbeddings,
				ChannelMeta: &relaycommon.ChannelMeta{
					UpstreamModelName: tt.model,
				},
			}
			info.SetEstimatePromptTokens(tt.estimate)
			resp := &http.Response{StatusCode: http.StatusOK, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(tt.body))}
			apiErr, usage := MultimodalEmbeddingHandler(c, resp, info)
			if tt.wantErrCode {
				require.NotNil(t, apiErr)
				assert.Equal(t, http.StatusInternalServerError, apiErr.StatusCode)
				assert.Empty(t, w.Body.String())
				return
			}
			require.Nil(t, apiErr)
			require.NotNil(t, usage)
			assert.Equal(t, tt.wantUsage, *usage)
			var response dto.EmbeddingResponse
			require.NoError(t, common.Unmarshal(w.Body.Bytes(), &response))
			assert.Equal(t, "list", response.Object)
			assert.Equal(t, tt.model, response.Model)
			assert.Equal(t, tt.wantUsage, response.Usage)
		})
	}
}
