package middleware

import (
	"net/http"
	"slices"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/gin-gonic/gin"
)

var corsAllowedMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
var corsAllowedHeaders = map[string]string{
	"accept":               "Accept",
	"authorization":        "Authorization",
	"content-type":         "Content-Type",
	"new-api-user":         "New-Api-User",
	"new-api-organization": "New-Api-Organization",
	"origin":               "Origin",
	"x-requested-with":     "X-Requested-With",
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := strings.TrimSpace(c.GetHeader("Origin"))
		if origin == "" {
			c.Next()
			return
		}

		allowed := false
		allowAll := false
		allowCredentials := false
		if isPublicAPIPath(c.Request.URL.Path) {
			settings := system_setting.GetCORSSettings()
			if settings.Mode == system_setting.CORSModeDisabled {
				c.Next()
				return
			}
			switch settings.Mode {
			case system_setting.CORSModeAll:
				allowed = true
				allowAll = true
			case system_setting.CORSModeAllowlist:
				allowed = originAllowed(origin, settings.AllowedOrigins)
			}
		} else {
			allowed = isDashboardCORSOriginAllowed(c.Request, origin)
			allowCredentials = true
		}

		if !allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"code":    "CORS_ORIGIN_FORBIDDEN",
				"message": "request origin is not allowed",
			})
			return
		}

		vary := c.Writer.Header()
		vary.Add("Vary", "Origin")
		if allowAll {
			c.Header("Access-Control-Allow-Origin", "*")
		} else {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		if allowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if c.Request.Method == http.MethodOptions && c.GetHeader("Access-Control-Request-Method") != "" {
			requestedMethod := strings.ToUpper(strings.TrimSpace(c.GetHeader("Access-Control-Request-Method")))
			if !slices.Contains(corsAllowedMethods, requestedMethod) {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
			requestedHeaders, ok := normalizeCORSRequestHeaders(c.GetHeader("Access-Control-Request-Headers"))
			if !ok {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
			vary.Add("Vary", "Access-Control-Request-Method")
			vary.Add("Vary", "Access-Control-Request-Headers")
			c.Header("Access-Control-Allow-Methods", strings.Join(corsAllowedMethods, ","))
			if requestedHeaders != "" {
				c.Header("Access-Control-Allow-Headers", requestedHeaders)
			}
			c.Header("Access-Control-Max-Age", "43200")
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// Dashboard POSTs carry Origin even when the page and API are on the same
// site. A deployment's own origin must not depend on the public relay list.
func isDashboardCORSOriginAllowed(request *http.Request, rawOrigin string) bool {
	origin, err := common.NormalizeOrigin(rawOrigin)
	if err != nil {
		return false
	}
	return isAllowedSessionOrigin(request, origin) ||
		originAllowed(origin, system_setting.DefaultCORSOrigins)
}

// TLS is commonly terminated by the reverse proxy. With the original Host
// preserved, HTTPS for that exact host is still this deployment's origin.
// Do not trust arbitrary forwarded-host/proto headers or relax host/port checks.
func isSameDeploymentOrigin(request *http.Request, origin string) bool {
	scheme := "http"
	if request.TLS != nil {
		scheme = "https"
	}
	requestOrigin, err := common.NormalizeOrigin(scheme + "://" + request.Host)
	if err == nil && origin == requestOrigin {
		return true
	}
	if request.TLS == nil {
		secureOrigin, err := common.NormalizeOrigin("https://" + request.Host)
		return err == nil && origin == secureOrigin
	}
	return false
}

func normalizeCORSRequestHeaders(requested string) (string, bool) {
	if strings.TrimSpace(requested) == "" {
		return "", true
	}
	headers := strings.Split(requested, ",")
	for i, header := range headers {
		canonical, ok := corsAllowedHeaders[strings.ToLower(strings.TrimSpace(header))]
		if !ok {
			return "", false
		}
		headers[i] = canonical
	}
	return strings.Join(headers, ","), true
}

func originAllowed(origin string, allowedOrigins []string) bool {
	origin = strings.TrimSuffix(strings.TrimSpace(origin), "/")
	for _, allowedOrigin := range allowedOrigins {
		if strings.EqualFold(
			origin,
			strings.TrimSuffix(strings.TrimSpace(allowedOrigin), "/"),
		) {
			return true
		}
	}
	return false
}

func isPublicAPIPath(path string) bool {
	if path == "/v2/video_generation" {
		return true
	}
	if taskID, ok := strings.CutPrefix(path, "/v2/query/video_generation/"); ok {
		return taskID != "" && !strings.Contains(taskID, "/")
	}
	return (path == "/v1" || strings.HasPrefix(path, "/v1/")) &&
		path != "/v1/dashboard" && !strings.HasPrefix(path, "/v1/dashboard/")
}

func Version() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-New-Api-Version", common.Version)
		c.Next()
	}
}
