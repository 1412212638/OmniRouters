package controller

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/logger"
	relaychannel "github.com/QuantumNous/new-api/relay/channel"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/service"

	"github.com/gin-gonic/gin"
)

const upstreamRelayUploadPath = "/v1/files"

// RelayUpload synchronously forwards a multipart upload to the selected
// channel. The gateway does not persist a file record or expose a later
// retrieval operation; the upstream response is returned as-is.
func RelayUpload(c *gin.Context) {
	contentType := c.Request.Header.Get("Content-Type")
	if !strings.HasPrefix(strings.ToLower(contentType), "multipart/form-data;") {
		c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": gin.H{
			"message": "Content-Type must be multipart/form-data",
			"type":    "invalid_request_error",
			"code":    "unsupported_media_type",
		}})
		return
	}
	if c.Request.Body == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{
			"message": "request body is empty",
			"type":    "invalid_request_error",
			"code":    "invalid_request_body",
		}})
		return
	}

	body := c.Request.Body
	defer body.Close()

	baseURL := common.GetContextKeyString(c, constant.ContextKeyChannelBaseUrl)
	if strings.TrimSpace(baseURL) == "" {
		writeUploadProxyError(c, http.StatusBadGateway, "selected channel has no upstream base URL")
		return
	}
	channelType := common.GetContextKeyInt(c, constant.ContextKeyChannelType)
	upstreamURL := relaycommon.GetFullRequestURL(baseURL, upstreamRelayUploadPath, channelType)
	if c.Request.URL.RawQuery != "" {
		upstreamURL += "?" + c.Request.URL.RawQuery
	}

	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, upstreamURL, body)
	if err != nil {
		writeUploadProxyError(c, http.StatusInternalServerError, "failed to create upstream upload request")
		return
	}
	if c.Request.ContentLength > 0 {
		req.ContentLength = c.Request.ContentLength
	} else {
		req.ContentLength = -1
	}
	req.Header.Set("Content-Type", contentType)
	if accept := c.Request.Header.Get("Accept"); accept != "" {
		req.Header.Set("Accept", accept)
	}
	if userAgent := c.Request.Header.Get("User-Agent"); userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	}
	channelKey := common.GetContextKeyString(c, constant.ContextKeyChannelKey)
	channelInfo := &relaycommon.RelayInfo{
		ChannelMeta: &relaycommon.ChannelMeta{
			ChannelType:     channelType,
			ApiKey:          channelKey,
			HeadersOverride: common.GetContextKeyStringMap(c, constant.ContextKeyChannelHeaderOverride),
		},
	}
	relaychannel.SetupApiRequestHeader(channelInfo, c, &req.Header)
	if channelKey != "" {
		req.Header.Set("Authorization", "Bearer "+channelKey)
	}
	headerOverrides, err := relaychannel.ResolveHeaderOverride(channelInfo, c)
	if err != nil {
		writeUploadProxyError(c, http.StatusBadRequest, "invalid channel header override")
		return
	}
	for key, value := range headerOverrides {
		if strings.EqualFold(key, "Content-Type") || strings.EqualFold(key, "Content-Length") ||
			strings.EqualFold(key, "Transfer-Encoding") || strings.EqualFold(key, "Host") {
			continue
		}
		req.Header.Set(key, value)
	}

	settings, _ := common.GetContextKeyType[dto.ChannelSettings](c, constant.ContextKeyChannelSetting)
	client, err := service.GetHttpClientWithProxySettings(settings.Proxy, settings)
	if err != nil {
		writeUploadProxyError(c, http.StatusBadGateway, "failed to create upstream client")
		return
	}
	// Preserve the upstream response and avoid following redirects with the
	// uploaded body or forwarding credentials to another host.
	clientCopy := *client
	clientCopy.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}
	common.RecordUpstreamAddress(c, upstreamURL)
	common.RecordUpstreamAddress(c, settings.Proxy)

	resp, err := clientCopy.Do(req)
	if err != nil {
		if common.IsRequestBodyTooLargeError(err) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": gin.H{
				"message": "request body exceeds the configured limit",
				"type":    "invalid_request_error",
				"code":    "request_body_too_large",
			}})
			return
		}
		logger.LogError(c, fmt.Sprintf("upload proxy request failed: %s", err.Error()))
		writeUploadProxyError(c, http.StatusBadGateway, "upstream upload request failed")
		return
	}
	defer resp.Body.Close()

	copyUploadResponseHeaders(c, resp.Header)
	c.Status(resp.StatusCode)
	if _, err = io.Copy(c.Writer, resp.Body); err != nil {
		logger.LogError(c, fmt.Sprintf("upload proxy response copy failed: %s", err.Error()))
	}
}

func copyUploadResponseHeaders(c *gin.Context, headers http.Header) {
	for key, values := range headers {
		if strings.EqualFold(key, "Content-Length") || strings.EqualFold(key, "Connection") || strings.EqualFold(key, "Transfer-Encoding") {
			continue
		}
		for _, value := range values {
			c.Writer.Header().Add(key, value)
		}
	}
}

func writeUploadProxyError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": gin.H{
		"message": message,
		"type":    "upstream_error",
		"code":    "upload_proxy_error",
	}})
}
