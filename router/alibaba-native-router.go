package router

import (
	"net/http"

	"github.com/QuantumNous/new-api/controller"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/gin-gonic/gin"
)

// Host-owned aliases keep the reserved /api namespace closed to plugin-defined
// routes. Both aliases pin the existing declaration and reuse its billing and
// ownership checks; historical /ali routes remain available.
func registerAlibabaNativeAPIRoutes(apiRouter *gin.RouterGroup) {
	for _, service := range []string{"video-generation", "image2video"} {
		path := "/v1/services/aigc/" + service + "/video-synthesis"
		apiRouter.POST(path,
			middleware.RouteTag("relay"), middleware.TokenAuth(), middleware.SystemPerformanceCheck(),
			pinDeclaredTaskPluginRoute(http.MethodPost, "/ali/api"+path), middleware.ModelRequestRateLimit(),
			middleware.PrepareTaskPluginRoute(), middleware.Distribute(), controller.RelayTask,
		)
	}
	apiRouter.GET("/v1/tasks/:task_id",
		middleware.RouteTag("relay"), middleware.TokenAuth(), middleware.SystemPerformanceCheck(),
		pinDeclaredTaskPluginRoute(http.MethodGet, "/ali/api/v1/tasks/:task_id"), middleware.PrepareTaskPluginRoute(),
	)
}
