package dto

import relaydto "github.com/QuantumNous/new-api/relaykit/dto"

type OpenAIImageTaskSubmitResponse struct {
	TaskID string `json:"task_id"`
	Status string `json:"status"`
}

type OpenAIImageTaskError struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
}

type OpenAIImageTaskFetchResponse struct {
	TaskID string                 `json:"task_id"`
	Status string                 `json:"status"`
	Data   []relaydto.ImageData   `json:"data,omitempty"`
	Error  *OpenAIImageTaskError  `json:"error,omitempty"`
}
