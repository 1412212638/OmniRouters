package common

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewReplayableBodyReaderKeepsStorageLifecycleWithCaller(t *testing.T) {
	payload := []byte(`{"model":"test-model","input":"hello"}`)
	storage, err := CreateBodyStorage(payload)
	require.NoError(t, err)
	defer storage.Close()

	body := NewReplayableBodyReader(storage)
	assert.EqualValues(t, len(payload), body.Size())
	_, exposesCloser := any(body).(io.Closer)
	assert.False(t, exposesCloser, "the request body must not expose the storage closer")

	req, err := http.NewRequest(http.MethodPost, "https://example.com", body)
	require.NoError(t, err)
	require.NoError(t, req.Body.Close())

	replayBody, err := body.NewReader()
	require.NoError(t, err, "closing the HTTP request body must not close the storage")
	replay, err := io.ReadAll(replayBody)
	require.NoError(t, err)
	require.NoError(t, replayBody.Close())
	assert.Equal(t, payload, replay)

	require.NoError(t, storage.Close())
	_, err = body.NewReader()
	require.ErrorIs(t, err, ErrStorageClosed)
}

func TestGetRequestBodyReturnsOriginalReadErrorToLaterReaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", io.NopCloser(&errorReader{err: io.ErrUnexpectedEOF}))

	_, firstErr := GetRequestBody(ctx)
	_, secondErr := GetRequestBody(ctx)

	require.ErrorIs(t, firstErr, io.ErrUnexpectedEOF)
	require.ErrorIs(t, secondErr, io.ErrUnexpectedEOF)
	assert.EqualError(t, secondErr, firstErr.Error())
}

func TestUpstreamAddressRecordingKeepsOnlyDeduplicatedOrigins(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	RecordUpstreamAddress(ctx, "https://user:secret@private.example/v1/chat?key=secret")
	RecordUpstreamAddress(ctx, "https://private.example/another/path")

	assert.Equal(t, []string{"https://private.example"}, GetUpstreamAddresses(ctx))
	assert.True(t, SameUpstreamOrigin("https://api.example/v1", "https://api.example/other"))
	assert.False(t, SameUpstreamOrigin("https://api.example:8443", "https://api.example"))
}

type errorReader struct {
	err error
}

func (r *errorReader) Read([]byte) (int, error) { return 0, r.err }
