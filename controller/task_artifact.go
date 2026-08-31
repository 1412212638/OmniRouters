package controller

import (
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/QuantumNous/new-api/middleware"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/relay"
	relaychannel "github.com/QuantumNous/new-api/relay/channel"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
)

var taskArtifactKeyPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._~-]{0,127}$`)

func GetTask(c *gin.Context) {
	task, exists, err := model.GetByTaskId(c.GetInt("id"), c.Param("key"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": "Failed to query task"}})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"message": "Task not found"}})
		return
	}
	createdAt := task.CreatedAt
	if createdAt == 0 {
		createdAt = task.SubmitTime
	}
	c.JSON(http.StatusOK, gin.H{"task_id": task.TaskID, "platform": task.Platform, "status": task.Status, "progress": task.Progress, "fail_reason": task.FailReason, "created_at": createdAt, "finished_at": task.FinishTime})
}

type taskArtifactResponse struct {
	Key        string `json:"key"`
	Type       string `json:"type"`
	MimeType   string `json:"mime_type,omitempty"`
	ContentURL string `json:"content_url"`
}

func GetTaskArtifacts(c *gin.Context) {
	task, exists, err := model.GetByTaskId(c.GetInt("id"), c.Param("key"))
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "artifact_not_found", "message": "Task or artifact not found"}})
		return
	}
	artifacts, err := projectTaskArtifacts(task)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": gin.H{"code": "artifact_plugin_error", "message": err.Error()}})
		return
	}
	items := make([]taskArtifactResponse, 0, len(artifacts))
	for _, artifact := range artifacts {
		contentURL, urlErr := service.BuildTaskArtifactContentURL(task.TaskID, artifact.Key)
		if urlErr != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		items = append(items, taskArtifactResponse{Key: artifact.Key, Type: artifact.Type, MimeType: artifact.MimeType, ContentURL: contentURL})
	}
	c.JSON(http.StatusOK, gin.H{"task_id": task.TaskID, "artifacts": items})
}

func projectTaskArtifacts(task *model.Task) ([]relaychannel.TaskArtifact, error) {
	if task == nil || task.Status != model.TaskStatusSuccess || task.PrivateData.Execution == nil || task.PrivateData.Execution.TaskPlugin == nil {
		return []relaychannel.TaskArtifact{}, nil
	}
	adaptor := relay.GetTaskAdaptor(task.Platform)
	provider, ok := adaptor.(relaychannel.TaskArtifactProvider)
	if !ok {
		return []relaychannel.TaskArtifact{}, nil
	}
	return provider.ListArtifacts(task)
}

func TaskArtifactContent(c *gin.Context) {
	taskID, artifactKey := c.Param("key"), strings.TrimSpace(c.Param("artifact_key"))
	if !taskArtifactKeyPattern.MatchString(artifactKey) {
		c.Status(http.StatusNotFound)
		return
	}
	var task *model.Task
	var exists bool
	var err error
	if middleware.IsTaskArtifactAccess(c) {
		task, exists, err = model.GetByOnlyTaskId(taskID)
	} else {
		task, exists, err = model.GetByTaskId(c.GetInt("id"), taskID)
	}
	if err != nil || !exists || task == nil || task.Status != model.TaskStatusSuccess {
		c.Status(http.StatusNotFound)
		return
	}
	adaptor := relay.GetTaskAdaptor(task.Platform)
	provider, ok := adaptor.(relaychannel.TaskContentRequestProvider)
	if !ok {
		if artifactKey == "video" && task.GetResultURL() != "" {
			proxyTaskArtifact(c, task.GetResultURL(), nil)
			return
		}
		c.Status(http.StatusNotFound)
		return
	}
	descriptor, err := provider.BuildContentRequest(task, artifactKey, relaychannel.TaskArtifactClientRequest{Method: c.Request.Method})
	if err != nil || descriptor == nil {
		c.Status(http.StatusBadGateway)
		return
	}
	proxyTaskArtifact(c, descriptor.URL, descriptor)
}

func proxyTaskArtifact(c *gin.Context, rawURL string, descriptor *relaychannel.TaskContentRequest) {
	if rawURL == "" || service.ValidateSSRFProtectedFetchURL(rawURL) != nil {
		c.Status(http.StatusForbidden)
		return
	}
	method := c.Request.Method
	if descriptor != nil && descriptor.Method != "" {
		method = strings.ToUpper(descriptor.Method)
	}
	if method != http.MethodGet && method != http.MethodHead {
		c.Status(http.StatusBadGateway)
		return
	}
	req, err := http.NewRequestWithContext(c.Request.Context(), method, rawURL, nil)
	if err != nil {
		c.Status(http.StatusBadGateway)
		return
	}
	resp, err := service.GetSSRFProtectedHTTPClient().Do(req)
	if err != nil {
		c.Status(http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	for _, name := range []string{"Content-Type", "Content-Length", "Content-Range", "Accept-Ranges", "ETag", "Last-Modified"} {
		if value := resp.Header.Get(name); value != "" {
			c.Header(name, value)
		}
	}
	c.Header("Cache-Control", "private, no-store")
	c.Status(resp.StatusCode)
	if method != http.MethodHead {
		_, _ = io.Copy(c.Writer, resp.Body)
	}
}
