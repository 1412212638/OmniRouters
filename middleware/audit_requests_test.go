package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestAuditRequestOutcomes(t *testing.T) {
	for _, tc := range []struct { name string; status int; body string; truncated bool; outcome string }{
		{"business_failure", 200, `{"success":false,"message":"secret must not be logged"}`, false, "failed"},
		{"success", 200, `{"success":true}`, false, "success"},
		{"string_data", 200, `{"success":true,"data":"a secret token"}`, false, "success"},
		{"array_data", 200, `{"success":false,"data":[]}`, false, "failed"},
		{"pending_2fa", 200, `{"success":true,"data":{"require_2fa":true}}`, false, "pending"},
		{"throttled", 429, "", false, "failed"},
		{"oversized", 200, `{"success":true}`, true, "unknown"},
		{"malformed", 200, "not json", false, "unknown"},
		{"empty", 204, "", false, "success"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			outcome, failure := auditOutcome(tc.status, []byte(tc.body), tc.truncated)
			require.Equal(t, tc.outcome, outcome)
			require.NotContains(t, failure, "secret")
		})
	}
}

func TestAuditRequestsCaptureCompressedBusinessFailureAndExcludeAuditReads(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	previous := model.DB
	model.DB = db
	t.Cleanup(func() { model.DB = previous; raw, _ := db.DB(); _ = raw.Close() })
	require.NoError(t, db.AutoMigrate(&model.AuditLog{}))
	r := gin.New()
	r.Use(gzip.Gzip(gzip.DefaultCompression), AuditRequests())
	r.POST("/api/example/:id", func(c *gin.Context) {
		c.Set("id", 7); c.Set("username", "operator")
		c.JSON(200, gin.H{"success": false, "message": "private-secret"})
	})
	r.GET("/api/audit", func(c *gin.Context) { c.Set("id", 7); c.JSON(200, gin.H{"success": true}) })
	r.POST("/api/user/login", func(c *gin.Context) { c.JSON(429, gin.H{"success": false}) })
	for _, path := range []string{"/api/example/8?secret=hidden", "/api/audit", "/api/user/login"} {
		method := http.MethodPost
		if path == "/api/audit" { method = http.MethodGet }
		req := httptest.NewRequest(method, path, nil)
		req.Header.Set("Accept-Encoding", "gzip")
		req.Header.Set("Authorization", "Bearer hidden")
		r.ServeHTTP(httptest.NewRecorder(), req)
	}
	var entries []model.AuditLog
	require.NoError(t, db.Order("id").Find(&entries).Error)
	require.Len(t, entries, 2)
	require.Equal(t, 7, entries[0].UserId)
	require.Equal(t, "failed", entries[0].Outcome)
	require.NotContains(t, entries[0].Other, "private-secret")
	require.NotContains(t, entries[0].Other, "hidden")
	var details model.AuditDetails
	require.NoError(t, common.UnmarshalJsonStr(entries[0].Other, &details))
	require.Equal(t, "/api/example/:id #8", details.Target)
	require.Equal(t, "security", entries[1].Category)
	require.Zero(t, entries[1].UserId)
}
