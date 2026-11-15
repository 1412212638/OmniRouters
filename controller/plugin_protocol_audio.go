package controller

import (
	"context"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	pluginruntime "github.com/QuantumNous/new-api/pkg/jsplugin"
	"github.com/QuantumNous/new-api/pkg/perf_metrics"
	"github.com/QuantumNous/new-api/relay"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	relayconstant "github.com/QuantumNous/new-api/relay/constant"
	"github.com/QuantumNous/new-api/relaykit/types"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
)

const maxTaskPluginAudioBytes = 16 << 20

// serveTaskPluginAudioProtocol handles the two synchronous speech protocols.
// Submission, persistence, pre-consume and settlement remain in the shared
// task path; only the final wire representation is protocol-specific.
func serveTaskPluginAudioProtocol(c *gin.Context, pinned pluginruntime.PinnedEndpoint, deps pluginProtocolBridgeDeps) {
	deps = deps.withDefaults()
	requestValue, exists := c.Get(pluginruntime.ContextKeyProtocolRequest)
	protocolRequest, ok := requestValue.(pluginruntime.ProtocolRequestContext)
	if !exists || !ok || protocolRequest.Protocol != pinned.Protocol || pinned.Plugin == nil {
		respondPluginProtocolError(c, http.StatusInternalServerError, "task_protocol_error", "Task protocol request failed")
		return
	}
	if protocolRequest.Stream {
		respondPluginProtocolError(c, http.StatusBadRequest, "invalid_request_error", "Synchronous speech requests do not support stream=true")
		return
	}

	clientRequest := c.Request
	var relayInfo *relaycommon.RelayInfo
	var relayInfoErr error
	var outcome *taskSubmissionOutcome
	var taskErr *dto.TaskError
	func() {
		submissionContext, cancelSubmission := context.WithTimeout(
			context.WithoutCancel(clientRequest.Context()), deps.submissionTimeout,
		)
		c.Request = clientRequest.Clone(submissionContext)
		defer func() {
			c.Request = clientRequest
			cancelSubmission()
		}()

		relayInfo, relayInfoErr = relaycommon.GenRelayInfo(c, types.RelayFormatTask, nil, nil)
		if relayInfoErr != nil {
			return
		}
		relayInfo.RelayMode = relayconstant.RelayModeAudioSpeech
		relayInfo.IsStream = false
		relayInfo.OriginModelName = c.GetString("resolved_task_model")
		if action := c.GetString("task_action"); action != "" {
			relayInfo.Action = action
		}
		if taskErr = relay.ResolveOriginTask(c, relayInfo); taskErr != nil {
			return
		}
		if taskErr = relay.ApplyOriginTaskAffinity(c, relayInfo); taskErr != nil {
			return
		}
		outcome, taskErr = deps.submit(c, relayInfo)
	}()
	if clientRequest.Context().Err() != nil {
		return
	}
	if relayInfoErr != nil {
		logger.LogError(c, "build task audio protocol relay info failed: "+relayInfoErr.Error())
		respondPluginProtocolError(c, http.StatusInternalServerError, "task_protocol_error", "Task protocol request failed")
		return
	}
	if taskErr != nil {
		respondTaskPluginAudioError(c, taskErr)
		return
	}
	if outcome == nil || outcome.Task == nil || outcome.RelayInfo == nil ||
		outcome.Task.UserId != relayInfo.UserId ||
		outcome.Task.Platform != constant.TaskPlatform(pinned.Plugin.Meta.Key) {
		logger.LogError(c, "task audio protocol submission returned an invalid durable outcome")
		respondPluginProtocolError(c, http.StatusInternalServerError, "task_protocol_error", "Task protocol request failed")
		return
	}

	task := outcome.Task
	completedAtSubmission := task.Status == model.TaskStatusSuccess || task.Status == model.TaskStatusFailure
	if !completedAtSubmission {
		respondPluginProtocolError(c, http.StatusBadGateway, "task_protocol_error", "Speech provider did not return a synchronous result")
		return
	}
	if completedAtSubmission {
		perfmetrics.RecordRelaySample(outcome.RelayInfo, task.Status == model.TaskStatusSuccess, 0, 0, 0)
	}
	if task.Status == model.TaskStatusFailure {
		reason := strings.TrimSpace(task.FailReason)
		if reason == "" {
			reason = "speech synthesis failed"
		}
		respondTaskPluginAudioError(c, service.TaskErrorWrapperLocal(errors.New(reason), "audio_generation_failed", http.StatusBadRequest))
		return
	}

	view, err := service.BuildTaskPluginView(task)
	if err != nil {
		logger.LogError(c, "build task audio protocol view failed: "+err.Error())
		respondPluginProtocolError(c, http.StatusInternalServerError, "task_protocol_error", "Task protocol request failed")
		return
	}
	viewValue, err := taskPluginProtocolJSONValue(view)
	if err != nil {
		logger.LogError(c, "encode task audio protocol view failed: "+err.Error())
		respondPluginProtocolError(c, http.StatusInternalServerError, "task_protocol_error", "Task protocol request failed")
		return
	}
	payload, err := pinned.Plugin.Engine.CallPathWithAdmissionTimeout(
		c.Request.Context(), deps.admissionTimeout, "protocols", []string{pinned.Protocol, "render"}, protocolRequest.JSValue(), viewValue,
	)
	if err != nil {
		logger.LogError(c, "task audio protocol render hook failed: "+err.Error())
		respondPluginProtocolError(c, http.StatusInternalServerError, "task_protocol_error", "Task protocol request failed")
		return
	}
	response, ok := payload.(map[string]any)
	if !ok || response == nil {
		logger.LogError(c, "task audio protocol render hook must return an object")
		respondPluginProtocolError(c, http.StatusInternalServerError, "task_protocol_error", "Task protocol request failed")
		return
	}
	if pinned.Protocol == pluginruntime.ProtocolMiniMaxT2A {
		c.JSON(http.StatusOK, response)
		return
	}
	writeTaskPluginOpenAIAudio(c, response)
}

func writeTaskPluginOpenAIAudio(c *gin.Context, response map[string]any) {
	audioHex, _ := response["audioHex"].(string)
	audioHex = strings.TrimSpace(audioHex)
	if audioHex == "" || len(audioHex)%2 != 0 || len(audioHex)/2 > maxTaskPluginAudioBytes {
		respondPluginProtocolError(c, http.StatusInternalServerError, "task_protocol_error", "Task protocol audio result is invalid")
		return
	}
	audio, err := hex.DecodeString(audioHex)
	if err != nil || len(audio) == 0 {
		respondPluginProtocolError(c, http.StatusInternalServerError, "task_protocol_error", "Task protocol audio result is invalid")
		return
	}
	format, _ := response["format"].(string)
	mimeType, ok := taskPluginAudioMimeType(format)
	if !ok {
		respondPluginProtocolError(c, http.StatusInternalServerError, "task_protocol_error", "Task protocol audio format is invalid")
		return
	}
	c.Data(http.StatusOK, mimeType, audio)
}

func taskPluginAudioMimeType(format string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "mp3", "mpeg":
		return "audio/mpeg", true
	case "wav":
		return "audio/wav", true
	case "flac":
		return "audio/flac", true
	case "aac":
		return "audio/aac", true
	case "opus":
		return "audio/ogg", true
	case "pcm", "pcm_s16le":
		return "audio/pcm", true
	default:
		return "", false
	}
}

func respondTaskPluginAudioError(c *gin.Context, taskErr *dto.TaskError) {
	status := http.StatusInternalServerError
	if taskErr != nil && taskErr.StatusCode >= 400 && taskErr.StatusCode <= 599 {
		status = taskErr.StatusCode
	}
	message := "Task protocol request failed"
	code := "task_protocol_error"
	if taskErr != nil {
		message = taskErr.Message
		code = taskErr.Code
		var upstream struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}
		if common.UnmarshalJsonStr(message, &upstream) == nil && strings.TrimSpace(upstream.Message) != "" {
			message = upstream.Message
			if strings.TrimSpace(upstream.Code) != "" {
				code = upstream.Code
			}
		}
	}
	c.JSON(status, gin.H{"error": gin.H{
		"message": common.MessageWithRequestId(message, c.GetString(common.RequestIdKey)),
		"type":    "new_api_error",
		"code":    code,
	}})
}
