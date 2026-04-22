package openai_image

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/relay/channel"
	openaichannel "github.com/QuantumNous/new-api/relay/channel/openai"
	taskcommon "github.com/QuantumNous/new-api/relay/channel/task/taskcommon"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting/model_setting"

	"github.com/gin-gonic/gin"
)

type TaskAdaptor struct {
	taskcommon.BaseBilling
	openAIAdaptor openaichannel.Adaptor
}

type asyncImageSubmitResponse struct {
	ID     string `json:"id,omitempty"`
	TaskID string `json:"task_id,omitempty"`
	Status string `json:"status,omitempty"`
	Error  *struct {
		Code    string `json:"code,omitempty"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type imageTaskError struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
}

func (a *TaskAdaptor) Init(info *relaycommon.RelayInfo) {
	a.openAIAdaptor.Init(info)
}

func (a *TaskAdaptor) ValidateRequestAndSetAction(c *gin.Context, info *relaycommon.RelayInfo) *dto.TaskError {
	if _, ok := info.Request.(*dto.ImageRequest); !ok {
		return service.TaskErrorWrapperLocal(fmt.Errorf("invalid request type"), "invalid_request", http.StatusBadRequest)
	}
	info.Action = constant.TaskActionGenerate
	return nil
}

func (a *TaskAdaptor) BuildRequestURL(info *relaycommon.RelayInfo) (string, error) {
	requestURL, err := a.openAIAdaptor.GetRequestURL(info)
	if err != nil {
		return "", err
	}
	parsedURL, err := neturl.Parse(requestURL)
	if err != nil {
		return "", err
	}
	query := parsedURL.Query()
	query.Set("async", "true")
	parsedURL.RawQuery = query.Encode()
	return parsedURL.String(), nil
}

func (a *TaskAdaptor) BuildRequestHeader(c *gin.Context, req *http.Request, info *relaycommon.RelayInfo) error {
	if err := a.openAIAdaptor.SetupRequestHeader(c, &req.Header, info); err != nil {
		return err
	}
	contentType := c.Request.Header.Get("Content-Type")
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	return nil
}

func (a *TaskAdaptor) BuildRequestBody(c *gin.Context, info *relaycommon.RelayInfo) (io.Reader, error) {
	request, ok := info.Request.(*dto.ImageRequest)
	if !ok {
		return nil, fmt.Errorf("invalid request type")
	}

	if model_setting.GetGlobalSettings().PassThroughRequestEnabled || info.ChannelSetting.PassThroughBodyEnabled {
		storage, err := common.GetBodyStorage(c)
		if err != nil {
			return nil, err
		}
		return common.ReaderOnly(storage), nil
	}

	convertedRequest, err := a.openAIAdaptor.ConvertImageRequest(c, info, *request)
	if err != nil {
		return nil, err
	}
	relaycommon.AppendRequestConversionFromRequest(info, convertedRequest)

	switch converted := convertedRequest.(type) {
	case *bytes.Buffer:
		return converted, nil
	default:
		jsonData, err := common.Marshal(convertedRequest)
		if err != nil {
			return nil, err
		}
		if len(info.ParamOverride) > 0 {
			jsonData, err = relaycommon.ApplyParamOverrideWithRelayInfo(jsonData, info)
			if err != nil {
				return nil, err
			}
		}
		return bytes.NewBuffer(jsonData), nil
	}
}

func (a *TaskAdaptor) DoRequest(c *gin.Context, info *relaycommon.RelayInfo, requestBody io.Reader) (*http.Response, error) {
	return channel.DoTaskApiRequest(a, c, info, requestBody)
}

func (a *TaskAdaptor) DoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (taskID string, taskData []byte, taskErr *dto.TaskError) {
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, service.TaskErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError)
	}
	_ = resp.Body.Close()

	var submitResp asyncImageSubmitResponse
	if err := common.Unmarshal(responseBody, &submitResp); err != nil {
		return "", nil, service.TaskErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError)
	}
	if submitResp.Error != nil && submitResp.Error.Message != "" {
		return "", nil, service.TaskErrorWrapperLocal(fmt.Errorf(submitResp.Error.Message), "invalid_response", http.StatusBadRequest)
	}

	upstreamTaskID := strings.TrimSpace(submitResp.TaskID)
	if upstreamTaskID == "" {
		upstreamTaskID = strings.TrimSpace(submitResp.ID)
	}
	if upstreamTaskID == "" {
		return "", nil, service.TaskErrorWrapper(fmt.Errorf("task_id is empty"), "invalid_response", http.StatusInternalServerError)
	}

	status := toImageTaskFetchStatus(submitResp.Status)
	if status == "" {
		status = dto.VideoStatusQueued
	}

	c.JSON(http.StatusOK, dto.OpenAIImageTaskSubmitResponse{
		TaskID: info.PublicTaskID,
		Status: status,
	})
	return upstreamTaskID, responseBody, nil
}

func (a *TaskAdaptor) FetchTask(baseURL, key string, body map[string]any, proxy string) (*http.Response, error) {
	taskID, ok := body["task_id"].(string)
	if !ok || strings.TrimSpace(taskID) == "" {
		return nil, fmt.Errorf("invalid task_id")
	}
	baseURL = strings.TrimRight(baseURL, "/")
	path := fmt.Sprintf("/v1/images/generations/%s", taskID)
	if strings.HasSuffix(baseURL, "/v1") {
		path = fmt.Sprintf("/images/generations/%s", taskID)
	}
	uri := baseURL + path
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+key)
	client, err := service.GetHttpClientWithProxy(proxy)
	if err != nil {
		return nil, fmt.Errorf("new proxy http client failed: %w", err)
	}
	return client.Do(req)
}

func (a *TaskAdaptor) ParseTaskResult(respBody []byte) (*relaycommon.TaskInfo, error) {
	var raw map[string]any
	if err := common.Unmarshal(respBody, &raw); err != nil {
		return nil, err
	}

	status := toTaskStatus(common.Interface2String(raw["status"]))
	result := &relaycommon.TaskInfo{
		TaskID: common.Interface2String(raw["task_id"]),
		Status: status,
	}
	if result.TaskID == "" {
		result.TaskID = common.Interface2String(raw["id"])
	}

	if errInfo := extractImageTaskError(raw); errInfo != nil {
		result.Reason = errInfo.Message
		if result.Status == "" {
			result.Status = string(model.TaskStatusFailure)
		}
	}

	imageData := ExtractImageData(respBody)
	if len(imageData) > 0 {
		result.Url = firstImageURL(imageData)
		if result.Status == "" {
			result.Status = string(model.TaskStatusSuccess)
		}
	}

	if progress := common.Interface2String(raw["progress"]); progress != "" {
		result.Progress = progress
	}
	if result.Status == "" {
		switch {
		case len(imageData) > 0:
			result.Status = string(model.TaskStatusSuccess)
		case result.Reason != "":
			result.Status = string(model.TaskStatusFailure)
		default:
			result.Status = string(model.TaskStatusQueued)
		}
	}
	return result, nil
}

func (a *TaskAdaptor) GetModelList() []string {
	return nil
}

func (a *TaskAdaptor) GetChannelName() string {
	return "OpenAI Image"
}

func ExtractImageData(respBody []byte) []dto.ImageData {
	var response struct {
		Data   []dto.ImageData `json:"data,omitempty"`
		Images []dto.ImageData `json:"images,omitempty"`
		Output []dto.ImageData `json:"output,omitempty"`
		URL    string          `json:"url,omitempty"`
	}
	if err := common.Unmarshal(respBody, &response); err == nil {
		switch {
		case len(response.Data) > 0:
			return response.Data
		case len(response.Images) > 0:
			return response.Images
		case len(response.Output) > 0:
			return response.Output
		case response.URL != "":
			return []dto.ImageData{{Url: response.URL}}
		}
	}

	var raw map[string]any
	if err := common.Unmarshal(respBody, &raw); err != nil {
		return nil
	}
	for _, key := range []string{"data", "images", "output"} {
		items, ok := raw[key].([]any)
		if !ok || len(items) == 0 {
			continue
		}
		result := make([]dto.ImageData, 0, len(items))
		for _, item := range items {
			switch value := item.(type) {
			case string:
				if value != "" {
					result = append(result, dto.ImageData{Url: value})
				}
			case map[string]any:
				imageURL := common.Interface2String(value["url"])
				if imageURL == "" {
					imageURL = common.Interface2String(value["image_url"])
				}
				if imageURL != "" {
					result = append(result, dto.ImageData{Url: imageURL})
				}
			}
		}
		if len(result) > 0 {
			return result
		}
	}

	if url := common.Interface2String(raw["url"]); url != "" {
		return []dto.ImageData{{Url: url}}
	}
	return nil
}

func extractImageTaskError(raw map[string]any) *imageTaskError {
	if raw == nil {
		return nil
	}
	if message := common.Interface2String(raw["message"]); message != "" {
		return &imageTaskError{Message: message}
	}
	errorValue, ok := raw["error"]
	if !ok {
		return nil
	}
	switch value := errorValue.(type) {
	case string:
		if value == "" {
			return nil
		}
		return &imageTaskError{Message: value}
	case map[string]any:
		message := common.Interface2String(value["message"])
		if message == "" {
			message = common.Interface2String(value["detail"])
		}
		if message == "" {
			return nil
		}
		return &imageTaskError{
			Code:    common.Interface2String(value["code"]),
			Message: message,
		}
	}
	return nil
}

func firstImageURL(items []dto.ImageData) string {
	for _, item := range items {
		if item.Url != "" {
			return item.Url
		}
	}
	return ""
}

func toTaskStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "", "submitted", "pending", "queued":
		return string(model.TaskStatusQueued)
	case "processing", "in_progress", "running", "generating":
		return string(model.TaskStatusInProgress)
	case "completed", "complete", "succeeded", "success", "finished":
		return string(model.TaskStatusSuccess)
	case "failed", "failure", "error", "cancelled", "canceled":
		return string(model.TaskStatusFailure)
	default:
		return ""
	}
}

func toImageTaskFetchStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "", "submitted", "pending", "queued":
		return dto.VideoStatusQueued
	case "processing", "in_progress", "running", "generating":
		return dto.VideoStatusInProgress
	case "completed", "complete", "succeeded", "success", "finished":
		return dto.VideoStatusCompleted
	case "failed", "failure", "error", "cancelled", "canceled":
		return dto.VideoStatusFailed
	default:
		return dto.VideoStatusQueued
	}
}
