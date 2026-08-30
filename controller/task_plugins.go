package controller

import (
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	pluginruntime "github.com/QuantumNous/new-api/pkg/jsplugin"
	"github.com/QuantumNous/new-api/setting"
	"github.com/gin-gonic/gin"
)

type taskPluginAdminView struct {
	Key            string                    `json:"key"`
	Name           string                    `json:"name"`
	Version        string                    `json:"version"`
	Icon           string                    `json:"icon,omitempty"`
	Description    pluginruntime.LocalizedText `json:"description,omitempty"`
	Author         pluginruntime.AuthorMeta  `json:"author"`
	ChannelTypes   []int                     `json:"channel_types,omitempty"`
	Models         []string                  `json:"models,omitempty"`
	Enabled        bool                      `json:"enabled"`
	RuntimeEnabled bool                      `json:"runtime_enabled"`
}

type updateTaskPluginRequest struct { Enabled bool `json:"enabled"` }

func taskPluginViews() []taskPluginAdminView {
	snapshot := pluginruntime.DefaultRegistry.Snapshot()
	disabled := make(map[string]struct{}, len(snapshot.DisabledFactory))
	for _, key := range snapshot.DisabledFactory { disabled[key] = struct{}{} }
	plugins := make([]taskPluginAdminView, 0, len(snapshot.Factory)+len(snapshot.Override))
	for _, meta := range append(snapshot.Factory, snapshot.Override...) {
		_, off := disabled[meta.Key]
		plugins = append(plugins, taskPluginAdminView{Key: meta.Key, Name: meta.Name, Version: meta.Version, Icon: meta.Icon, Description: meta.Description, Author: meta.Author, ChannelTypes: meta.ChannelTypes, Models: meta.Models, Enabled: !off, RuntimeEnabled: pluginruntime.DefaultRegistry.Enabled()})
	}
	return plugins
}

func GetTaskPlugins(c *gin.Context) {
	common.ApiSuccess(c, gin.H{"runtime_enabled": pluginruntime.DefaultRegistry.Enabled(), "plugins": taskPluginViews()})
}

func UpdateTaskPlugin(c *gin.Context) {
	key := strings.TrimSpace(c.Param("key"))
	var request updateTaskPluginRequest
	if err := common.DecodeJson(c.Request.Body, &request); err != nil { c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid request parameters"}); return }
	snapshot := pluginruntime.DefaultRegistry.Snapshot()
	found := false
	for _, meta := range snapshot.Factory { if meta.Key == key { found = true; break } }
	if !found { common.ApiErrorMsg(c, "plugin not found"); return }
	keys := setting.GetTaskPluginDisabledFactoryKeys()
	next := make([]string, 0, len(keys)+1)
	for _, item := range keys { if item != key { next = append(next, item) } }
	if !request.Enabled { next = append(next, key) }
	if err := setting.SetTaskPluginDisabledFactoryKeysOption(next); err != nil { common.ApiError(c, err); return }
	encoded, err := common.Marshal(next)
	if err != nil { common.ApiError(c, err); return }
	if err = model.UpdateOption(setting.TaskPluginDisabledFactoryKeysKey, string(encoded)); err != nil { common.ApiError(c, err); return }
	pluginruntime.DefaultRegistry.SetDisabledFactoryKeys(next)
	common.ApiSuccess(c, gin.H{"runtime_enabled": pluginruntime.DefaultRegistry.Enabled(), "plugins": taskPluginViews()})
}
