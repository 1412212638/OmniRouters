package controller

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	relaychannel "github.com/QuantumNous/new-api/relay/channel"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestOmniDashboardOmitsMediaButPreservesUsageAndSteps(t *testing.T) {
	raw := []byte(`{"steps":[{"type":"user_input","content":[{"type":"image","data":"private-input","mime_type":"image/png"}]},{"type":"thought","signature":"keep-signature"},{"type":"model_output","content":[{"type":"video","data":"large-output","mime_type":"video/mp4"}]}],"usage":{"total_input_tokens":19,"total_output_tokens":57920}}`)
	result := omitOmniTaskMedia(raw)
	require.NotContains(t, string(result), "private-input")
	require.NotContains(t, string(result), "large-output")
	require.Contains(t, string(result), "keep-signature")
	require.Contains(t, string(result), "video/mp4")
	require.Contains(t, string(result), "57920")
	require.True(t, common.IsValidJSON(result))
	require.True(t, strings.Contains(string(raw), "large-output"), "persisted native results must not be modified")
}

func TestOmniInlineVideoSupportsRangeAndHead(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodHead} {
		response := httptest.NewRecorder()
		context, _ := gin.CreateTestContext(response)
		context.Request = httptest.NewRequest(method, "/content", nil)
		context.Request.Header.Set("Range", "bytes=2-5")
		serveInlineTaskContent(context, &relaychannel.TaskContentRequest{InlineData: []byte("0123456789"), MimeType: "video/mp4"})
		require.Equal(t, http.StatusPartialContent, response.Code)
		require.Equal(t, "bytes 2-5/10", response.Header().Get("Content-Range"))
		require.Equal(t, "video/mp4", response.Header().Get("Content-Type"))
		require.Equal(t, "private, no-store", response.Header().Get("Cache-Control"))
		if method == http.MethodHead {
			require.Empty(t, response.Body.String())
		} else {
			require.Equal(t, "2345", response.Body.String())
		}
	}
}
