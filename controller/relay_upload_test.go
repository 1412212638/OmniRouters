package controller

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestRelayUploadForwardsMultipartAndReturnsUpstreamResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fileContents := []byte("image bytes")
	upstreamResult := make(chan error, 1)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fail := func(err error) {
			upstreamResult <- err
			http.Error(w, "upstream assertion failed", http.StatusInternalServerError)
		}
		if r.Method != http.MethodPost || r.URL.Path != "/v1/upload" || r.Header.Get("Authorization") != "Bearer upstream-key" {
			fail(io.ErrUnexpectedEOF)
			return
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			fail(err)
			return
		}
		if r.FormValue("model") != "seedance-asset-upload" {
			fail(io.ErrUnexpectedEOF)
			return
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			fail(err)
			return
		}
		if header.Filename != "Ling.png" {
			_ = file.Close()
			fail(io.ErrUnexpectedEOF)
			return
		}
		defer file.Close()
		actual, err := io.ReadAll(file)
		if err != nil {
			fail(err)
			return
		}
		if !bytes.Equal(fileContents, actual) {
			fail(io.ErrUnexpectedEOF)
			return
		}
		upstreamResult <- nil
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Upstream-Upload", "accepted")
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"id":"asset_123","url":"https://upstream.example/asset_123"}`)
	}))
	defer upstream.Close()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	require.NoError(t, writer.WriteField("model", "seedance-asset-upload"))
	part, err := writer.CreateFormFile("file", "Ling.png")
	require.NoError(t, err)
	_, err = part.Write(fileContents)
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	request := httptest.NewRequest(http.MethodPost, "/v1/upload", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = request
	common.SetContextKey(c, constant.ContextKeyChannelBaseUrl, upstream.URL)
	common.SetContextKey(c, constant.ContextKeyChannelKey, "upstream-key")
	common.SetContextKey(c, constant.ContextKeyChannelSetting, dto.ChannelSettings{})

	RelayUpload(c)

	require.NoError(t, <-upstreamResult)
	require.Equal(t, http.StatusCreated, recorder.Code)
	require.Equal(t, "accepted", recorder.Header().Get("X-Upstream-Upload"))
	require.JSONEq(t, `{"id":"asset_123","url":"https://upstream.example/asset_123"}`, recorder.Body.String())
}

func TestIsSuccessfulUploadStatus(t *testing.T) {
	tests := []struct {
		status int
		want   bool
	}{
		{status: http.StatusOK, want: true},
		{status: http.StatusCreated, want: true},
		{status: http.StatusNoContent, want: true},
		{status: http.StatusMultipleChoices, want: false},
		{status: http.StatusBadRequest, want: false},
	}
	for _, test := range tests {
		require.Equal(t, test.want, isSuccessfulUploadStatus(test.status))
	}
}

func TestUploadPricingRequiresFixedPrice(t *testing.T) {
	err := validateUploadPrice(false)
	require.ErrorContains(t, err, "fixed per-request price")
	require.NoError(t, validateUploadPrice(true))
}
