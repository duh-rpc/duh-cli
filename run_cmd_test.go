package lint_test

import (
	"bytes"
	"strings"
	"testing"

	lint "github.com/duh-rpc/duhrpc-lint"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunCmdHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exitCode := lint.RunCmd(nil, &stdout, &stderr, []string{"--help"})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "duhrpc-lint - Validate OpenAPI specs")
	assert.Contains(t, stdout.String(), "Usage:")
	assert.Contains(t, stdout.String(), "Exit Codes:")
	assert.Empty(t, stderr.String())
}

func TestRunCmdVersion(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exitCode := lint.RunCmd(nil, &stdout, &stderr, []string{"--version"})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "duhrpc-lint version")
	assert.Contains(t, stdout.String(), "1.0.0")
	assert.Empty(t, stderr.String())
}

func TestRunCmdValidFile(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exitCode := lint.RunCmd(nil, &stdout, &stderr, []string{"testdata/minimal-valid.yaml"})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "✓")
	assert.Contains(t, stdout.String(), "minimal-valid.yaml")
	assert.Contains(t, stdout.String(), "DUH-RPC compliant")
	assert.Empty(t, stderr.String())
}

func TestRunCmdFileNotFound(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exitCode := lint.RunCmd(nil, &stdout, &stderr, []string{"nonexistent.yaml"})

	assert.Equal(t, 2, exitCode)
	assert.Contains(t, stderr.String(), "Error:")
	assert.Contains(t, stderr.String(), "file not found")
	assert.Empty(t, stdout.String())
}

func TestRunCmdInvalidYAML(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exitCode := lint.RunCmd(nil, &stdout, &stderr, []string{"testdata/invalid-syntax.yaml"})

	assert.Equal(t, 2, exitCode)
	assert.Contains(t, stderr.String(), "Error:")
	assert.Contains(t, stderr.String(), "failed to parse OpenAPI spec")
	assert.Empty(t, stdout.String())
}

func TestRunCmdNoArguments(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exitCode := lint.RunCmd(nil, &stdout, &stderr, []string{})

	assert.Equal(t, 2, exitCode)
	assert.Contains(t, stderr.String(), "Error:")
	assert.Contains(t, stderr.String(), "Exactly one OpenAPI file path is required")
	assert.Empty(t, stdout.String())
}

func TestRunCmdMultipleArguments(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exitCode := lint.RunCmd(nil, &stdout, &stderr, []string{"file1.yaml", "file2.yaml"})

	assert.Equal(t, 2, exitCode)
	assert.Contains(t, stderr.String(), "Error:")
	output := strings.ToLower(stderr.String())
	require.True(t, strings.Contains(output, "exactly one") || strings.Contains(output, "one openapi file"))
	assert.Empty(t, stdout.String())
}
