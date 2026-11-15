package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/relay"
	relaychannel "github.com/QuantumNous/new-api/relay/channel"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/gin-gonic/gin"
)

// Only file-backed Hailuo tasks need this lookup. H3's direct CDN results,
// other plugins and ordinary OpenAI content keep their existing handlers.
func isHailuoFileTask(task *model.Task) bool {
	if task == nil {
		return false
	}
	key := string(task.Platform)
	if execution := task.PrivateData.Execution; execution != nil && execution.TaskPlugin != nil {
		key = execution.TaskPlugin.Key
	}
	if key != "hailuo" && key != strconv.Itoa(constant.ChannelTypeMiniMax) {
		return false
	}
	var data map[string]any
	if common.Unmarshal(task.Data, &data) != nil {
		return false
	}
	if nested, ok := data["data"].(map[string]any); ok && nested["task_id"] != nil {
		if value, exists := nested["data"]; exists {
			data, _ = value.(map[string]any)
		}
	}
	switch fileID := data["file_id"].(type) {
	case string:
		return strings.TrimSpace(fileID) != ""
	case float64:
		return fileID > 0
	default:
		return false
	}
}

func buildHailuoFileContentRequest(task *model.Task, channel *model.Channel, method string) (*relaychannel.TaskContentRequest, error) {
	adaptor := relay.GetTaskAdaptor(task.Platform)
	provider, ok := adaptor.(relaychannel.TaskContentRequestProvider)
	if !ok {
		return nil, fmt.Errorf("Hailuo video plugin is unavailable")
	}
	key := task.PrivateData.Key
	if key == "" {
		if channel.ChannelInfo.IsMultiKey {
			return nil, fmt.Errorf("Hailuo task credential is unavailable")
		}
		key = channel.Key
	}
	if strings.TrimSpace(key) == "" || strings.ContainsAny(key, "\r\n") {
		return nil, fmt.Errorf("Hailuo task credential is unavailable")
	}
	baseURL := channel.GetBaseURL()
	if baseURL == "" {
		baseURL = constant.GetChannelBaseURL(channel.Type)
	}
	adaptor.Init(&relaycommon.RelayInfo{ChannelMeta: &relaycommon.ChannelMeta{
		ChannelId: channel.Id, ChannelType: channel.Type, ChannelBaseUrl: baseURL,
		ApiKey: key, ChannelSetting: channel.GetSetting(),
	}})
	return provider.BuildContentRequest(task, "video", relaychannel.TaskArtifactClientRequest{Method: method})
}

func serveHailuoFileTaskContent(c *gin.Context, task *model.Task) {
	channel, err := model.CacheGetChannel(task.ChannelId)
	if err != nil || channel == nil {
		videoProxyError(c, http.StatusBadGateway, "server_error", "Failed to retrieve video channel")
		return
	}
	descriptor, err := buildHailuoFileContentRequest(task, channel, c.Request.Method)
	if err != nil || descriptor == nil || !descriptor.Credentialless || len(descriptor.Headers) != 0 || len(descriptor.Body) != 0 {
		videoProxyError(c, http.StatusBadGateway, "server_error", "Failed to resolve video content")
		return
	}
	proxyTaskArtifact(c, descriptor.URL, descriptor)
}
