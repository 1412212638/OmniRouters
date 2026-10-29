package controller

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	perfmetrics "github.com/QuantumNous/new-api/pkg/perf_metrics"
	relaychannel "github.com/QuantumNous/new-api/relay/channel"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	relayhelper "github.com/QuantumNous/new-api/relay/helper"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/relaykit/types"
	"github.com/QuantumNous/new-api/service"

	"github.com/bytedance/gopkg/util/gopool"
	"github.com/gin-gonic/gin"
)

const relayUploadPath = "/v1/upload"

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

	relayInfo, err := relaycommon.GenRelayInfo(c, types.RelayFormatTask, nil, nil)
	if err != nil {
		writeUploadProxyError(c, http.StatusInternalServerError, "failed to initialize upload request")
		return
	}
	relayInfo.InitChannelMeta(c)
	if err := service.ValidateSubscriptionRequest(relayInfo); err != nil {
		writeUploadProxyError(c, http.StatusForbidden, err.Error())
		return
	}

	baseURL := common.GetContextKeyString(c, constant.ContextKeyChannelBaseUrl)
	if strings.TrimSpace(baseURL) == "" {
		writeUploadProxyError(c, http.StatusBadGateway, "selected channel has no upstream base URL")
		return
	}
	priceData, err := relayhelper.ModelPriceHelperPerCall(c, relayInfo)
	if err != nil {
		writeUploadProxyError(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := validateUploadPrice(priceData.UsePrice); err != nil {
		writeUploadProxyError(c, http.StatusBadRequest, err.Error())
		return
	}
	relayInfo.PriceData = priceData
	if apiErr := service.PreConsumeBilling(c, priceData.Quota, relayInfo); apiErr != nil {
		writeUploadProxyError(c, apiErr.StatusCode, apiErr.Error())
		return
	}
	billingSettled := false
	defer func() {
		if !billingSettled && relayInfo.Billing != nil {
			relayInfo.Billing.Refund(c)
		}
	}()
	channelType := common.GetContextKeyInt(c, constant.ContextKeyChannelType)
	upstreamURL := relaycommon.GetFullRequestURL(baseURL, relayUploadPath, channelType)
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
	relaychannel.SetupApiRequestHeader(relayInfo, c, &req.Header)
	if channelKey != "" {
		req.Header.Set("Authorization", "Bearer "+channelKey)
	}
	headerOverrides, err := relaychannel.ResolveHeaderOverride(relayInfo, c)
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
		recordUploadFailure(c, relayInfo, http.StatusBadGateway)
		writeUploadProxyError(c, http.StatusBadGateway, "upstream upload request failed")
		return
	}
	defer resp.Body.Close()

	if isSuccessfulUploadStatus(resp.StatusCode) {
		if err := service.SettleBilling(c, relayInfo, priceData.Quota); err != nil {
			logger.LogError(c, fmt.Sprintf("upload billing settlement failed: %s", err.Error()))
			writeUploadProxyError(c, http.StatusInternalServerError, "upload billing settlement failed")
			return
		}
		billingSettled = true
		recordUploadSuccess(c, relayInfo, resp.StatusCode)
	} else {
		recordUploadFailure(c, relayInfo, resp.StatusCode)
	}

	copyUploadResponseHeaders(c, resp.Header)
	c.Status(resp.StatusCode)
	if _, err = io.Copy(c.Writer, resp.Body); err != nil {
		logger.LogError(c, fmt.Sprintf("upload proxy response copy failed: %s", err.Error()))
	}
}

func isSuccessfulUploadStatus(status int) bool {
	return status >= http.StatusOK && status < http.StatusMultipleChoices
}

func validateUploadPrice(useFixedPrice bool) error {
	if !useFixedPrice {
		return fmt.Errorf("model pricing for /v1/upload must use a fixed per-request price; token-ratio pricing cannot be calculated from an upload response")
	}
	return nil
}

func uploadLogOther(c *gin.Context, info *relaycommon.RelayInfo, status int) *model.LogOther {
	other := model.NewLogOther()
	other.SetPublic("request_path", relayUploadPath)
	other.SetPublic("operation", "multipart_upload")
	other.SetPublic("billing_unit", "request")
	other.SetPublic("status_code", status)
	if info != nil {
		other.SetPublic("model_price", info.PriceData.ModelPrice)
		other.SetPublic("group_ratio", info.PriceData.GroupRatioInfo.GroupRatio)
	}
	service.AppendRelayLogAdminInfo(c, info, other)
	return other
}

func uploadUseTimeSeconds(info *relaycommon.RelayInfo) int {
	if info == nil || info.StartTime.IsZero() {
		return 0
	}
	seconds := int(time.Since(info.StartTime).Seconds())
	if seconds < 0 {
		return 0
	}
	return seconds
}

func recordUploadSuccess(c *gin.Context, info *relaycommon.RelayInfo, status int) {
	if info == nil {
		return
	}
	if info.UserId > 0 {
		model.UpdateUserUsedQuotaAndRequestCount(info.UserId, info.PriceData.Quota)
		model.UpdateChannelUsedQuota(info.ChannelId, info.PriceData.Quota)
		model.RecordConsumeLog(c, info.UserId, model.RecordConsumeLogParams{
			ChannelId:      info.ChannelId,
			ModelName:      info.OriginModelName,
			TokenName:      c.GetString("token_name"),
			Quota:          info.PriceData.Quota,
			Content:        "POST /v1/upload",
			TokenId:        info.TokenId,
			UseTimeSeconds: uploadUseTimeSeconds(info),
			Group:          info.UsingGroup,
			Other:          uploadLogOther(c, info, status),
		})
	}
	gopool.Go(func() {
		perfmetrics.RecordRelaySample(info, true, 0, 0, 0)
	})
}

func recordUploadFailure(c *gin.Context, info *relaycommon.RelayInfo, status int) {
	if info == nil {
		return
	}
	gopool.Go(func() {
		perfmetrics.RecordRelayFailureSample(info, status)
	})
	if !constant.ErrorLogEnabled || info.UserId <= 0 {
		return
	}
	other := uploadLogOther(c, info, status)
	other.SetPublic("error_type", "upstream_error")
	other.SetPublic("error_code", "upload_proxy_error")
	model.RecordErrorLog(c, info.UserId, info.ChannelId, info.OriginModelName, c.GetString("token_name"),
		fmt.Sprintf("Upstream upload failed with HTTP status %d", status), info.TokenId,
		uploadUseTimeSeconds(info), false, info.UsingGroup, other)
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
