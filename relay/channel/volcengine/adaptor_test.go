package volcengine

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
)

func TestConvertClaudeRequestStreamUsage(t *testing.T) {
	oldMode := gin.Mode()
	gin.SetMode(gin.TestMode)
	t.Cleanup(func() { gin.SetMode(oldMode) })

	for _, tc := range []struct {
		name      string
		baseURL   string
		stream    bool
		wantUsage bool
	}{
		{name: "chat completions streaming", stream: true, wantUsage: true},
		{name: "chat completions non-streaming"},
		{name: "native Messages streaming", baseURL: "glm-coding-plan", stream: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request := &dto.ClaudeRequest{
				Model:     "glm-5.3",
				MaxTokens: lo.ToPtr(uint(32)),
				Stream:    lo.ToPtr(tc.stream),
				Messages:  []dto.ClaudeMessage{{Role: "user", Content: "Hello"}},
			}
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)
			common.SetContextKey(c, constant.ContextKeyOriginalModel, request.Model)
			common.SetContextKey(c, constant.ContextKeyChannelType, constant.ChannelTypeVolcEngine)
			common.SetContextKey(c, constant.ContextKeyChannelBaseUrl, tc.baseURL)
			info := relaycommon.GenRelayInfoClaude(c, request)
			info.InitChannelMeta(c)

			adaptor := Adaptor{}
			adaptor.Init(info)
			converted, err := adaptor.ConvertClaudeRequest(c, info, request)
			require.NoError(t, err)
			if tc.baseURL == "" {
				require.IsType(t, &dto.GeneralOpenAIRequest{}, converted)
			} else {
				require.IsType(t, &dto.ClaudeRequest{}, converted)
			}
			body, err := common.Marshal(converted)
			require.NoError(t, err)
			body, err = relaycommon.RemoveDisabledFields(body, info.ChannelOtherSettings, info.ChannelSetting.PassThroughBodyEnabled)
			require.NoError(t, err)
			var payload struct {
				Stream        bool               `json:"stream"`
				StreamOptions *dto.StreamOptions `json:"stream_options"`
			}
			require.NoError(t, common.Unmarshal(body, &payload))
			require.Equal(t, tc.stream, payload.Stream)
			if tc.wantUsage {
				require.NotNil(t, payload.StreamOptions)
				require.True(t, payload.StreamOptions.IncludeUsage)
			} else {
				require.Nil(t, payload.StreamOptions)
			}
		})
	}
}
