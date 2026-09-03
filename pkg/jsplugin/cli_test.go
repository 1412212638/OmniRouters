package jsplugin

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPluginCLIWarnsOnParseTaskResultInProgressFallback(t *testing.T) {
	tempDir := t.TempDir()
	pluginPath := filepath.Join(tempDir, "fallback.js")
	require.NoError(t, os.WriteFile(pluginPath, []byte(`
export const meta = { apiVersion: 1, key: "fallback", name: "Fallback", version: "1.0.0", author: {name: "Test"}, models: ["m"], fetchMode: "per_task" };
export function buildSubmitRequest(ctx) { return {url: ctx.baseUrl}; }
export function parseSubmitResponse() { return {taskId: "task"}; }
export function buildQueryRequest(ctx) { return {url: ctx.baseUrl}; }
export function parseTaskResult(ctx, body) { return {status: statuses[body.status] || "IN_PROGRESS"}; }
const statuses = { done: "SUCCESS" };
`), 0o600))

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	assert.Equal(t, 0, RunCLI([]string{"lint", pluginPath}, &stdout, &stderr))
	assert.Contains(t, stdout.String(), "plugin fallback@1.0.0 is valid")
	assert.Contains(t, stderr.String(), `|| "IN_PROGRESS"`)
}
