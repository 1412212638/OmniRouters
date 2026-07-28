package model

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/ClickHouse/clickhouse-go/v2/lib/proto"
	"github.com/QuantumNous/new-api/common"
	"github.com/glebarez/sqlite"
	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestSanitizeDBErrorStripsDriverMessage(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		want   string
		leaked string
	}{
		{"mysql duplicate entry", &mysql.MySQLError{Number: 1062, Message: "Duplicate entry 'secret-value'"}, "mysql error 1062", "secret-value"},
		{"postgres unique violation", &pgconn.PgError{Code: "23505", Detail: "Key (k)=(secret-value) exists"}, "postgres error SQLSTATE 23505", "secret-value"},
		{"clickhouse exception", &proto.Exception{Code: 241, Message: "secret-value"}, "clickhouse error 241", "secret-value"},
		{"wrapped driver error", fmt.Errorf("exec failed: %w", &mysql.MySQLError{Number: 1064, Message: "secret-value"}), "mysql error 1064", "secret-value"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := sanitizeDBError(tc.err)
			require.Error(t, got)
			assert.Equal(t, tc.want, got.Error())
			assert.NotContains(t, got.Error(), tc.leaked)
		})
	}
}

func TestSanitizeDBErrorSQLiteDriver(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	execErr := db.Exec("INSERT INTO missing_table (k) VALUES (?)", "secret-value").Error
	require.Error(t, execErr)

	got := sanitizeDBError(execErr)
	assert.Regexp(t, `^sqlite error \d+$`, got.Error())
	assert.NotContains(t, got.Error(), "secret-value")
}

func TestSanitizeDBErrorKeepsNonDriverErrors(t *testing.T) {
	err := fmt.Errorf("dial tcp 127.0.0.1:3306: connect: connection refused")
	assert.Equal(t, err, sanitizeDBError(err))
}

func TestGormLoggerEndToEndSanitizedOutput(t *testing.T) {
	previousDebug := common.DebugEnabled
	t.Cleanup(func() { common.DebugEnabled = previousDebug })

	execQuery := func() string {
		var buf bytes.Buffer
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: newGormLogger(&buf)})
		require.NoError(t, err)
		db.Exec("SELECT * FROM missing_table WHERE k = ?", "secret-value")
		return buf.String()
	}

	common.DebugEnabled = false
	out := execQuery()
	assert.Contains(t, out, "k = ?")
	assert.NotContains(t, out, "secret-value")
	assert.Contains(t, out, "sqlite error")
	assert.Contains(t, out, "gorm_logger_test.go")

	common.DebugEnabled = true
	debugOut := execQuery()
	assert.Contains(t, debugOut, "secret-value")
	assert.Contains(t, debugOut, "no such table")
}
