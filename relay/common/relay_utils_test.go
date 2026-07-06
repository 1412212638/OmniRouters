package common

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/constant"
	"github.com/gin-gonic/gin"
)

func newTaskTestContext(contentType string, body *bytes.Buffer) *gin.Context {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/v1/video/generations", body)
	req.Header.Set("Content-Type", contentType)
	c.Request = req
	return c
}

func newJSONTaskTestContext(body string) *gin.Context {
	return newTaskTestContext("application/json", bytes.NewBufferString(body))
}

func newTaskRelayInfo() *RelayInfo {
	return &RelayInfo{TaskRelayInfo: &TaskRelayInfo{}}
}

func TestValidateBasicTaskRequestRequiresPromptWithoutMediaInput(t *testing.T) {
	c := newJSONTaskTestContext(`{"model":"kling-v1"}`)
	info := newTaskRelayInfo()

	taskErr := ValidateBasicTaskRequest(c, info, constant.TaskActionGenerate)
	if taskErr == nil {
		t.Fatal("expected prompt validation error")
	}
	if taskErr.Message != "prompt is required" {
		t.Fatalf("unexpected error message: %s", taskErr.Message)
	}
}

func TestValidateBasicTaskRequestAllowsSingleImageWithoutPrompt(t *testing.T) {
	c := newJSONTaskTestContext(`{"model":"kling-v1","image":"https://example.com/frame.png"}`)
	info := newTaskRelayInfo()

	if taskErr := ValidateBasicTaskRequest(c, info, constant.TaskActionGenerate); taskErr != nil {
		t.Fatalf("unexpected validation error: %v", taskErr)
	}

	req, err := GetTaskRequest(c)
	if err != nil {
		t.Fatalf("get task request: %v", err)
	}
	if len(req.Images) != 1 || req.Images[0] != "https://example.com/frame.png" {
		t.Fatalf("single image was not normalized into images: %#v", req.Images)
	}
}

func TestValidateBasicTaskRequestAllowsMetadataMediaWithoutPrompt(t *testing.T) {
	c := newJSONTaskTestContext(`{
		"model":"seedance-1-0-pro",
		"metadata":{
			"content":[
				{"type":"video_url","video_url":{"url":"https://example.com/input.mp4"}}
			]
		}
	}`)
	info := newTaskRelayInfo()

	if taskErr := ValidateBasicTaskRequest(c, info, constant.TaskActionGenerate); taskErr != nil {
		t.Fatalf("unexpected validation error: %v", taskErr)
	}
}

func TestValidateMultipartDirectAllowsInputReferenceStringWithoutPrompt(t *testing.T) {
	c := newJSONTaskTestContext(`{"model":"sora-2","input_reference":"https://example.com/frame.png"}`)
	info := newTaskRelayInfo()

	if taskErr := ValidateMultipartDirect(c, info); taskErr != nil {
		t.Fatalf("unexpected validation error: %v", taskErr)
	}
	if info.Action != constant.TaskActionGenerate {
		t.Fatalf("action = %s, want %s", info.Action, constant.TaskActionGenerate)
	}

	req, err := GetTaskRequest(c)
	if err != nil {
		t.Fatalf("get task request: %v", err)
	}
	if len(req.Images) != 1 || req.Images[0] != "https://example.com/frame.png" {
		t.Fatalf("input_reference was not normalized into images: %#v", req.Images)
	}
}

func TestValidateMultipartDirectAllowsInputReferenceFileWithoutPrompt(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("model", "sora-2"); err != nil {
		t.Fatalf("write model field: %v", err)
	}
	fileWriter, err := writer.CreateFormFile("input_reference", "frame.png")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := fileWriter.Write([]byte("fake image bytes")); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	c := newTaskTestContext(writer.FormDataContentType(), &body)
	info := newTaskRelayInfo()

	if taskErr := ValidateMultipartDirect(c, info); taskErr != nil {
		t.Fatalf("unexpected validation error: %v", taskErr)
	}
	if info.Action != constant.TaskActionGenerate {
		t.Fatalf("action = %s, want %s", info.Action, constant.TaskActionGenerate)
	}
}

func TestValidateMultipartDirectNormalizesWanImageField(t *testing.T) {
	c := newJSONTaskTestContext(`{"model":"wan2.7-i2v","prompt":"animate","image":" https://example.com/first.png "}`)
	info := newTaskRelayInfo()

	if taskErr := ValidateMultipartDirect(c, info); taskErr != nil {
		t.Fatalf("unexpected validation error: %v", taskErr)
	}
	if info.Action != constant.TaskActionGenerate {
		t.Fatalf("action = %s, want %s", info.Action, constant.TaskActionGenerate)
	}

	req, err := GetTaskRequest(c)
	if err != nil {
		t.Fatalf("get task request: %v", err)
	}
	if len(req.Images) != 1 || req.Images[0] != "https://example.com/first.png" {
		t.Fatalf("single image was not normalized into images: %#v", req.Images)
	}
}

// TestTaskDurationBounds guards the billing invariant that user-supplied
// video duration (a quota multiplier via OtherRatio "seconds") is bounded, so
// it can never overflow quota calculation into a negative charge.
func TestTaskDurationBounds(t *testing.T) {
	gin.SetMode(gin.TestMode)

	newContext := func(t *testing.T, body string) (*gin.Context, *RelayInfo) {
		request := httptest.NewRequest(http.MethodPost, "/v1/video/generations", strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		context, _ := gin.CreateTestContext(httptest.NewRecorder())
		context.Request = request
		return context, &RelayInfo{TaskRelayInfo: &TaskRelayInfo{}}
	}

	tests := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{
			name:    "huge duration is rejected",
			body:    `{"model":"sora-2","prompt":"a cat","duration":9999999999}`,
			wantErr: true,
		},
		{
			name:    "huge seconds string is rejected",
			body:    `{"model":"sora-2","prompt":"a cat","seconds":"9999999999"}`,
			wantErr: true,
		},
		{
			name:    "negative duration is rejected",
			body:    `{"model":"sora-2","prompt":"a cat","duration":-8}`,
			wantErr: true,
		},
		{
			name: "normal duration is accepted",
			body: `{"model":"sora-2","prompt":"a cat","seconds":"8"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name+" (multipart direct)", func(t *testing.T) {
			context, info := newContext(t, tt.body)
			taskErr := ValidateMultipartDirect(context, info)
			if tt.wantErr {
				require.NotNil(t, taskErr)
				require.Equal(t, "invalid_seconds", taskErr.Code)
			} else {
				require.Nil(t, taskErr)
			}
		})
		t.Run(tt.name+" (basic task request)", func(t *testing.T) {
			context, info := newContext(t, tt.body)
			taskErr := ValidateBasicTaskRequest(context, info, constant.TaskActionGenerate)
			if tt.wantErr {
				require.NotNil(t, taskErr)
				require.Equal(t, "invalid_seconds", taskErr.Code)
			} else {
				require.Nil(t, taskErr)
			}
		})
	}
}
