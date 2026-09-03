package controller

import (
	"net/http"

	"github.com/QuantumNous/new-api/common"
	"github.com/gin-gonic/gin"
)

const etagVersionPublicContent = "public-content:v1"

type publicContentResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

// serveRevalidatedJSON preserves the public API envelope while enabling
// content-based conditional responses across replicas and JSON encoders.
func serveRevalidatedJSON(c *gin.Context, content string) {
	body, err := common.Marshal(publicContentResponse{Success: true, Data: content})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	etag := common.ETagFor(etagVersionPublicContent, content)
	c.Header("ETag", etag)
	c.Header("Cache-Control", "no-cache")
	c.Header("Vary", "Accept-Encoding")
	if common.ETagMatches(c.GetHeader("If-None-Match"), etag) {
		c.Status(http.StatusNotModified)
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", body)
}
