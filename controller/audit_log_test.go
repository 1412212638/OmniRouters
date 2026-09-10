package controller

import (
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestAuditFiltersAndSelfIsolation(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	previous := model.DB; model.DB = db
	t.Cleanup(func() { model.DB = previous; raw, _ := db.DB(); _ = raw.Close() })
	require.NoError(t, db.AutoMigrate(&model.AuditLog{}))
	for _, item := range []model.AuditLog{
		{UserId: 1, Category: "operation", Action: "update", Outcome: "failed", CreatedAt: 10},
		{UserId: 2, Category: "operation", Action: "update", Outcome: "success", CreatedAt: 20},
		{UserId: 1, Category: "access", Action: "request", Outcome: "success", CreatedAt: 30},
	} { require.NoError(t, model.RecordAuditLog(&item)) }
	r := gin.New()
	r.GET("/api/audit/self", func(c *gin.Context) { c.Set("id", 1) }, GetSecurityAuditLogs)
	for _, tc := range []struct { query string; count int; status int }{
		{"?view=important", 1, 200},
		{"?view=all&user_id=2", 0, 200},
		{"?view=access", 1, 200},
		{"?start=11&end=25", 0, 200},
		{"?outcome=failed", 1, 200},
		{"?view=invalid", 0, 400},
		{"?page_size=-1", 0, 400},
	} {
		res := httptest.NewRecorder(); r.ServeHTTP(res, httptest.NewRequest("GET", "/api/audit/self"+tc.query, nil))
		require.Equal(t, tc.status, res.Code)
		if tc.status == 200 {
			var result struct { Data struct { Total int `json:"total"` } `json:"data"` }
			require.NoError(t, common.Unmarshal(res.Body.Bytes(), &result))
			require.Equal(t, tc.count, result.Data.Total, tc.query)
		}
	}
}
