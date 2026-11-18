package controller

import (
	"io"
	"net/http"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
)

// DoubaoTaskNotification never applies client-provided status or billing data.
func DoubaoTaskNotification(c *gin.Context) {
	var body struct {
		ID string `json:"id"`
	}
	if err := common.DecodeJson(io.LimitReader(c.Request.Body, 1<<20), &body); err != nil || body.ID == "" {
		c.Status(http.StatusBadRequest)
		return
	}
	if err := service.ReceiveDoubaoCallback(c.Request.Context(), c.Param("task_id"), c.Query("token"), body.ID); err != nil {
		// A notification may beat task persistence; returning non-2xx permits
		// provider retries while scheduled polling still provides a fallback.
		c.Status(http.StatusServiceUnavailable)
		return
	}
	c.Status(http.StatusNoContent)
}
