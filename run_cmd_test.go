package duh_test

import (
	"bytes"
	"strings"
	"testing"

	lint "github.com/duh-rpc/duhrpc-lint"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunCmdHelp(t *testing.T) {
	var stdout bytes.Buffer
	exitCode := lint.RunCmd(&stdout, []string{"--help"})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "duhrpc-lint - Validate OpenAPI specs")
	assert.Contains(t, stdout.String(), "Usage:")
	assert.Contains(t, stdout.String(), "Exit Codes:")
}

func TestRunCmdVersion(t *testing.T) {
	var stdout bytes.Buffer
	exitCode := lint.RunCmd(&stdout, []string{"--version"})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "duhrpc-lint version")
	assert.Contains(t, stdout.String(), "1.0.0")
}

func TestRunCmdFileNotFound(t *testing.T) {
	var stdout bytes.Buffer
	exitCode := lint.RunCmd(&stdout, []string{"nonexistent.yaml"})

	assert.Equal(t, 2, exitCode)
	assert.Contains(t, stdout.String(), "Error:")
	assert.Contains(t, stdout.String(), "file not found")
}

func TestRunCmdNoArguments(t *testing.T) {
	var stdout bytes.Buffer
	exitCode := lint.RunCmd(&stdout, []string{})

	assert.Equal(t, 2, exitCode)
	assert.Contains(t, stdout.String(), "Error:")
	assert.Contains(t, stdout.String(), "Exactly one OpenAPI file path is required")
}

func TestRunCmdMultipleArguments(t *testing.T) {
	var stdout bytes.Buffer
	exitCode := lint.RunCmd(&stdout, []string{"file1.yaml", "file2.yaml"})

	assert.Equal(t, 2, exitCode)
	assert.Contains(t, stdout.String(), "Error:")
	output := strings.ToLower(stdout.String())
	require.True(t, strings.Contains(output, "exactly one") || strings.Contains(output, "one openapi file"))
}
