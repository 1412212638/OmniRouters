package router

import (
	"fmt"
	"net/http"

	"github.com/QuantumNous/new-api/controller"
	"github.com/QuantumNous/new-api/middleware"
	pluginruntime "github.com/QuantumNous/new-api/pkg/jsplugin"
	"github.com/QuantumNous/new-api/relaykit/types"
	"github.com/gin-gonic/gin"
)

func SetTaskPluginProtocolRouter(router *gin.Engine) {
	// Resolve the declaration on every request so marketplace installation and
	// activation take effect without restarting the HTTP server.
	router.POST("/v1/systemone",
		middleware.RouteTag("relay"), middleware.TokenAuth(), middleware.SystemPerformanceCheck(),
		pinTypeSafeRoute, middleware.ModelRequestRateLimit(), middleware.PrepareTaskPluginRoute(),
		middleware.Distribute(), controller.RelayTask,
	)
	for _, protocol := range pluginruntime.HostProtocols() {
		for _, operation := range protocol.Operations {
			for _, method := range operation.Methods {
				handlers, err := taskPluginProtocolHandlers(protocol.Name, operation.Name)
				if err != nil {
					panic(err)
				}
				router.Handle(method, operation.Path, handlers...)
			}
		}
	}
	registerDoubaoNativeRoutes(router)
}

// registerDoubaoNativeRoutes keeps the historical /doubao-prefixed Ark task
// API available alongside the official /api/v3 aliases registered in the API
// router. Both paths enter the same task-plugin pipeline and are pinned to the
// currently loaded plugin generation on every request.
func registerDoubaoNativeRoutes(router *gin.Engine) {
	const (
		createPath = "/doubao/api/v3/contents/generations/tasks"
		queryPath  = "/doubao/api/v3/contents/generations/tasks/:task_id"
	)
	router.POST(createPath, doubaoNativeSubmitHandlers(http.MethodPost, createPath)...)
	router.GET(queryPath, doubaoNativeQueryHandlers(http.MethodGet, queryPath)...)
}

func doubaoNativeSubmitHandlers(method, declaredPath string) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		middleware.RouteTag("relay"), middleware.TokenAuth(), middleware.SystemPerformanceCheck(),
		pinDeclaredTaskPluginRoute(method, declaredPath), middleware.ModelRequestRateLimit(),
		middleware.PrepareTaskPluginRoute(), middleware.Distribute(), controller.RelayTask,
	}
}

func doubaoNativeQueryHandlers(method, declaredPath string) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		middleware.RouteTag("relay"), middleware.TokenAuth(), middleware.SystemPerformanceCheck(),
		pinDeclaredTaskPluginRoute(method, declaredPath), middleware.PrepareTaskPluginRoute(),
	}
}

func pinDeclaredTaskPluginRoute(method, declaredPath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		generation := pluginruntime.DefaultRegistry.Generation()
		binding, ok := generation.LookupDeclaredRoute(method, declaredPath)
		if !ok || binding.Plugin == nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": gin.H{
				"message": "Doubao native route is unavailable; install and enable the Doubao plugin",
				"type":    "new_api_error", "code": "task_plugin_route_unavailable",
			}})
			return
		}
		c.Set(pluginruntime.ContextKeyPinnedRoute, pluginruntime.PinnedRoute{
			Generation: generation, Plugin: binding.Plugin, Route: binding.Route,
		})
	}
}

func pinTypeSafeRoute(c *gin.Context) {
	generation := pluginruntime.DefaultRegistry.Generation()
	binding, ok := generation.LookupDeclaredRoute(http.MethodPost, "/typesafe/v1/systemone")
	if !ok || binding.Plugin == nil || binding.Plugin.Meta.Key != "typesafe" {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": gin.H{
			"message": "TypeSafe native route is unavailable; install and enable the TypeSafe plugin",
			"type":    "new_api_error", "code": "task_plugin_route_unavailable",
		}})
		return
	}
	c.Set(pluginruntime.ContextKeyPinnedRoute, pluginruntime.PinnedRoute{
		Generation: generation, Plugin: binding.Plugin, Route: binding.Route,
	})
}

func taskPluginProtocolHandlers(protocol, operation string) ([]gin.HandlerFunc, error) {
	switch protocol + "." + operation {
	case "openai_responses.create":
		return []gin.HandlerFunc{
			middleware.RouteTag("relay"), middleware.SystemPerformanceCheck(), middleware.TokenAuth(),
			middleware.ModelRequestRateLimit(), middleware.RouteTaskPluginEndpoint(), middleware.PrepareTaskPluginEndpoint(), middleware.Distribute(),
			func(c *gin.Context) {
				controller.RelayTaskPluginEndpoint(c, func(c *gin.Context) { controller.Relay(c, types.RelayFormatOpenAIResponses) })
			},
		}, nil
	case "openai_image.generate", "openai_image.edit":
		return []gin.HandlerFunc{
			middleware.RouteTag("relay"), middleware.SystemPerformanceCheck(), middleware.TokenAuth(),
			middleware.ModelRequestRateLimit(), middleware.RouteTaskPluginEndpoint(), middleware.PrepareTaskPluginEndpoint(), middleware.Distribute(),
			func(c *gin.Context) {
				controller.RelayTaskPluginEndpoint(c, func(c *gin.Context) { controller.Relay(c, types.RelayFormatOpenAIImage) })
			},
		}, nil
	case "openai_video.create":
		return []gin.HandlerFunc{
			middleware.RouteTag("relay"), middleware.TokenAuth(), middleware.SystemPerformanceCheck(),
			middleware.RouteTaskPluginEndpoint(), middleware.TaskPluginEndpointOnly(middleware.ModelRequestRateLimit()), middleware.PrepareTaskPluginEndpoint(), middleware.Distribute(),
			func(c *gin.Context) { controller.RelayTaskPluginEndpoint(c, controller.RelayTask) },
		}, nil
	case "openai_responses.retrieve":
		return []gin.HandlerFunc{middleware.RouteTag("relay"), middleware.TokenAuth(), controller.RetrieveTaskPluginResponse}, nil
	case "openai_video.retrieve":
		return []gin.HandlerFunc{middleware.RouteTag("relay"), middleware.TokenAuth(), middleware.Distribute(), controller.RelayTaskFetch}, nil
	case "openai_video.content":
		return []gin.HandlerFunc{middleware.RouteTag("relay"), middleware.TokenAuth(), controller.VideoProxy}, nil
	default:
		return nil, fmt.Errorf("host protocol registry operation %s.%s has no handler", protocol, operation)
	}
}
