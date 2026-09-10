package model

// AuditDetails contains explicitly selected metadata, never request/response
// bodies. Keeping it in Other preserves compatibility with existing databases.
type AuditDetails struct {
	Version int `json:"version"`
	Method string `json:"method,omitempty"`
	Route string `json:"route,omitempty"`
	Target string `json:"target,omitempty"`
	AuthMethod string `json:"auth_method,omitempty"`
	HTTPStatus int `json:"http_status,omitempty"`
	Outcome string `json:"outcome,omitempty"`
	Failure string `json:"failure,omitempty"`
	Changes map[string]AuditChange `json:"changes,omitempty"`
	Params map[string]interface{} `json:"params,omitempty"`
}

type AuditChange struct {
	Before interface{} `json:"before"`
	After interface{} `json:"after"`
}
