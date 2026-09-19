package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestGeminiCountTokensRejectedBeforeAuthentication(t *testing.T) {
	router := gin.New()
	SetRelayRouter(router)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/v1beta/models/gemini-3-pro:countTokens", nil))
	require.Equal(t, http.StatusNotFound, response.Code)
}
