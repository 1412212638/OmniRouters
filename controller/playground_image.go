package controller

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/service"

	"github.com/gin-gonic/gin"
)

type playgroundImageSourceRequest struct {
	URL string `json:"url"`
}

// PlaygroundImageSource downloads a generated image server-side so the browser
// does not need CORS access to a provider's signed image URL before editing it.
func PlaygroundImageSource(c *gin.Context) {
	request := &playgroundImageSourceRequest{}
	if err := common.UnmarshalBodyReusable(c, request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "invalid image source request", "type": "invalid_request_error"},
		})
		return
	}

	imageURL := strings.TrimSpace(request.URL)
	if imageURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "image source url is required", "type": "invalid_request_error"},
		})
		return
	}

	mimeType, data, err := service.GetImageFromUrl(imageURL)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{
				"message": fmt.Sprintf("failed to read reference image: %s", err.Error()),
				"type":    "upstream_error",
			},
		})
		return
	}

	if separator := strings.IndexByte(mimeType, ';'); separator >= 0 {
		mimeType = mimeType[:separator]
	}
	if mimeType == "" {
		mimeType = "image/png"
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"data_url": fmt.Sprintf("data:%s;base64,%s", mimeType, data),
		},
	})
}
