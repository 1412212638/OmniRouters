package router

import (
	"github.com/QuantumNous/new-api/controller"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/gin-gonic/gin"
)

func SetTaskRouter(router *gin.Engine) {
	router.Group("/v1/tasks").Use(middleware.RouteTag("relay"), middleware.TokenAuth()).POST("/:key", middleware.PrepareTaskPluginSubmit(), middleware.Distribute(), controller.RelayTask)
	read := router.Group("/v1/tasks").Use(middleware.RouteTag("relay"), middleware.TokenAuth())
	read.GET("/:key", controller.GetTask)
	read.GET("/:key/artifacts", controller.GetTaskArtifacts)
	content := router.Group("/v1/tasks").Use(middleware.RouteTag("relay"), middleware.RedactTaskArtifactAccessQuery(), middleware.TokenOrTaskArtifactAccessAuth("key", "artifact_key"))
	content.GET("/:key/artifacts/:artifact_key/content", controller.TaskArtifactContent)
	content.HEAD("/:key/artifacts/:artifact_key/content", controller.TaskArtifactContent)
}
