package jsplugin

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/model"
	pluginruntime "github.com/QuantumNous/new-api/pkg/jsplugin"
	"github.com/QuantumNous/new-api/relay/channel"
	"github.com/QuantumNous/new-api/relay/channel/task/sora"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
)

type requestDescriptor struct {
	URL      string            `json:"url"`
	Method   string            `json:"method"`
	Headers  map[string]string `json:"headers"`
	Body     any               `json:"body"`
	BodyType string            `json:"bodyType"`
}

type submitResponse struct {
	TaskID   string `json:"taskId"`
	TaskData any    `json:"taskData"`
}

type taskResult struct {
	Status   string `json:"status"`
	Progress string `json:"progress"`
	Reason   string `json:"reason"`
}

// TaskAdaptor is the first compatibility bridge between the declarative task
// runtime and the repository's existing task lifecycle. Billing and fallback
// behavior come from the native Sora adaptor; the plugin only owns the JSON
// wire conversion and response parsing for requests selected below.
type TaskAdaptor struct {
	legacy *sora.TaskAdaptor
	plugin *pluginruntime.LoadedPlugin
	info   *relaycommon.RelayInfo
	useLegacy bool
	submit *requestDescriptor
}

var _ channel.TaskAdaptor = (*TaskAdaptor)(nil)

func New(plugin *pluginruntime.LoadedPlugin) channel.TaskAdaptor {
	return &TaskAdaptor{legacy: &sora.TaskAdaptor{}, plugin: plugin}
}

func (a *TaskAdaptor) Init(info *relaycommon.RelayInfo) {
	a.info = info
	a.legacy.Init(info)
}

func (a *TaskAdaptor) ValidateRequestAndSetAction(c *gin.Context, info *relaycommon.RelayInfo) *dto.TaskError {
	// Keep multipart and opaque requests on the proven native path until the
	// plugin bridge has a complete file-reference implementation.
	a.useLegacy = strings.Contains(strings.ToLower(c.GetHeader("Content-Type")), "multipart/")
	if a.useLegacy {
		return a.legacy.ValidateRequestAndSetAction(c, info)
	}
	if taskErr := a.legacy.ValidateRequestAndSetAction(c, info); taskErr != nil {
		return taskErr
	}
	if _, err := a.buildSubmit(c, info); err != nil {
		return service.TaskErrorWrapperLocal(err, "plugin_request_invalid", http.StatusBadRequest)
	}
	return nil
}

func (a *TaskAdaptor) BuildRequestURL(info *relaycommon.RelayInfo) (string, error) {
	if a.useLegacy {
		return a.legacy.BuildRequestURL(info)
	}
	if a.submit == nil {
		return "", fmt.Errorf("plugin submit request was not built")
	}
	return a.submit.URL, pluginruntime.ValidateRequestURL(a.submit.URL, info.ChannelBaseUrl, a.plugin.Meta.AllowedHosts)
}

func (a *TaskAdaptor) BuildRequestHeader(c *gin.Context, req *http.Request, info *relaycommon.RelayInfo) error {
	if a.useLegacy {
		return a.legacy.BuildRequestHeader(c, req, info)
	}
	if a.submit == nil {
		return fmt.Errorf("plugin submit request was not built")
	}
	for key, value := range a.submit.Headers {
		req.Header.Set(key, value)
	}
	return nil
}

func (a *TaskAdaptor) BuildRequestBody(c *gin.Context, info *relaycommon.RelayInfo) (io.Reader, error) {
	if a.useLegacy {
		return a.legacy.BuildRequestBody(c, info)
	}
	descriptor, err := a.buildSubmit(c, info)
	if err != nil {
		return nil, err
	}
	if descriptor.Body == nil {
		return nil, nil
	}
	if body, ok := descriptor.Body.(string); ok {
		return strings.NewReader(body), nil
	}
	encoded, err := common.Marshal(descriptor.Body)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(encoded), nil
}

func (a *TaskAdaptor) DoRequest(c *gin.Context, info *relaycommon.RelayInfo, body io.Reader) (*http.Response, error) {
	if a.useLegacy {
		return a.legacy.DoRequest(c, info, body)
	}
	return channel.DoTaskApiRequest(a, c, info, body)
}

func (a *TaskAdaptor) DoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (string, []byte, *dto.TaskError) {
	if a.useLegacy {
		return a.legacy.DoResponse(c, resp, info)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, service.TaskErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError)
	}
	var decoded any
	if err = common.Unmarshal(body, &decoded); err != nil {
		return "", nil, service.TaskErrorWrapper(err, "unmarshal_response_body_failed", http.StatusBadGateway)
	}
	value, err := a.plugin.Engine.Call(c.Request.Context(), "parseSubmitResponse", a.submitContext(info), map[string]any{
		"statusCode": resp.StatusCode,
		"headers":    resp.Header,
		"body":       decoded,
	})
	if err != nil {
		return "", nil, service.TaskErrorWrapper(err, "plugin_submit_response_failed", http.StatusBadGateway)
	}
	var parsed submitResponse
	if err = convert(value, &parsed); err != nil || strings.TrimSpace(parsed.TaskID) == "" {
		if err == nil {
			err = fmt.Errorf("plugin returned an empty taskId")
		}
		return "", nil, service.TaskErrorWrapper(err, "plugin_submit_response_invalid", http.StatusBadGateway)
	}
	clientBody := cloneMap(decoded)
	clientBody["id"] = info.PublicTaskID
	clientBody["task_id"] = info.PublicTaskID
	c.JSON(http.StatusOK, clientBody)
	var taskData []byte
	if parsed.TaskData != nil {
		taskData, err = common.Marshal(parsed.TaskData)
		if err != nil {
			return "", nil, service.TaskErrorWrapper(err, "plugin_submit_response_invalid", http.StatusBadGateway)
		}
	} else {
		taskData = body
	}
	return parsed.TaskID, taskData, nil
}

func (a *TaskAdaptor) GetModelList() []string { return append([]string(nil), a.plugin.Meta.Models...) }
func (a *TaskAdaptor) GetChannelName() string { return a.plugin.Meta.Name }

func (a *TaskAdaptor) FetchTask(baseURL, key string, body map[string]any, proxy string) (*http.Response, error) {
	ctx := map[string]any{
		"baseUrl": baseURL, "apiKey": key,
		"taskId": body["task_id"], "action": body["action"], "requestBody": body,
	}
	value, err := a.plugin.Engine.Call(context.Background(), "buildQueryRequest", ctx)
	if err != nil {
		return nil, err
	}
	var descriptor requestDescriptor
	if err = convert(value, &descriptor); err != nil {
		return nil, err
	}
	if err = pluginruntime.ValidateRequestURL(descriptor.URL, baseURL, a.plugin.Meta.AllowedHosts); err != nil {
		return nil, err
	}
	request, err := http.NewRequest(strings.ToUpper(descriptor.Method), descriptor.URL, nil)
	if err != nil {
		return nil, err
	}
	for name, value := range descriptor.Headers {
		request.Header.Set(name, value)
	}
	client, err := service.GetHttpClientWithProxy(proxy)
	if err != nil {
		return nil, err
	}
	return client.Do(request)
}

func (a *TaskAdaptor) ParseTaskResult(body []byte) (*relaycommon.TaskInfo, error) {
	var decoded any
	if err := common.Unmarshal(body, &decoded); err != nil {
		return nil, err
	}
	value, err := a.plugin.Engine.Call(context.Background(), "parseTaskResult", map[string]any{}, decoded)
	if err != nil {
		return nil, err
	}
	var parsed taskResult
	if err = convert(value, &parsed); err != nil {
		return nil, err
	}
	result := &relaycommon.TaskInfo{Status: parsed.Status, Progress: parsed.Progress, Reason: parsed.Reason}
	switch parsed.Status {
	case "QUEUED", "IN_PROGRESS", "SUCCESS", "FAILURE":
	default:
		result.Status = "UNKNOWN"
	}
	return result, nil
}

func (a *TaskAdaptor) EstimateBilling(c *gin.Context, info *relaycommon.RelayInfo) map[string]float64 {
	return a.legacy.EstimateBilling(c, info)
}
func (a *TaskAdaptor) AdjustBillingOnSubmit(info *relaycommon.RelayInfo, data []byte) map[string]float64 {
	return a.legacy.AdjustBillingOnSubmit(info, data)
}
func (a *TaskAdaptor) AdjustBillingOnComplete(task *model.Task, result *relaycommon.TaskInfo) int {
	return a.legacy.AdjustBillingOnComplete(task, result)
}
func (a *TaskAdaptor) ConvertToOpenAIVideo(task *model.Task) ([]byte, error) {
	return a.legacy.ConvertToOpenAIVideo(task)
}

func (a *TaskAdaptor) buildSubmit(c *gin.Context, info *relaycommon.RelayInfo) (*requestDescriptor, error) {
	value, exists := c.Get("task_request")
	if !exists {
		return nil, fmt.Errorf("task_request not found in context")
	}
	encoded, err := common.Marshal(value)
	if err != nil {
		return nil, err
	}
	var requestBody any
	if err = common.Unmarshal(encoded, &requestBody); err != nil {
		return nil, err
	}
	result, err := a.plugin.Engine.Call(c.Request.Context(), "buildSubmitRequest", a.submitContextWithRequest(info, requestBody))
	if err != nil {
		return nil, err
	}
	var descriptor requestDescriptor
	if err = convert(result, &descriptor); err != nil {
		return nil, err
	}
	if err = pluginruntime.ValidateRequestURL(descriptor.URL, info.ChannelBaseUrl, a.plugin.Meta.AllowedHosts); err != nil {
		return nil, err
	}
	a.submit = &descriptor
	return &descriptor, nil
}

func (a *TaskAdaptor) submitContext(info *relaycommon.RelayInfo) map[string]any {
	return a.submitContextWithRequest(info, nil)
}

func (a *TaskAdaptor) submitContextWithRequest(info *relaycommon.RelayInfo, requestBody any) map[string]any {
	return map[string]any{
		"baseUrl": info.ChannelBaseUrl, "apiKey": info.ApiKey,
		"model": info.OriginModelName, "upstreamModel": info.UpstreamModelName,
		"action": info.Action, "originTaskId": info.OriginTaskID,
		"requestBody": requestBody,
	}
}

func cloneMap(value any) map[string]any {
	if object, ok := value.(map[string]any); ok {
		cloned := make(map[string]any, len(object)+2)
		for key, item := range object {
			cloned[key] = item
		}
		return cloned
	}
	return map[string]any{}
}

func convert(value any, target any) error {
	encoded, err := common.Marshal(value)
	if err != nil {
		return err
	}
	return common.Unmarshal(encoded, target)
}
