package middleware

import (
	"net/http"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
)

// RealNameBusinessGate blocks authenticated business actions when the
// administrator has enabled mandatory real-name verification. Profile and
// verification routes intentionally do not use this middleware.
func RealNameBusinessGate() gin.HandlerFunc {
	return func(c *gin.Context) {
		required, err := service.RealNameRequiredForUser(c.GetInt("id"))
		if err != nil {
			common.ApiError(c, err)
			c.Abort()
			return
		}
		if required {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "message": "请先完成实名认证后再使用平台服务"})
			c.Abort()
			return
		}
		c.Next()
	}
}
