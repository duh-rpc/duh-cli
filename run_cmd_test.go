package duh_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	lint "github.com/duh-rpc/duh-cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunCmdHelp(t *testing.T) {
	var stdout bytes.Buffer
	exitCode := lint.RunCmd(&stdout, []string{"--help"})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "duh is a command-line tool")
	assert.Contains(t, stdout.String(), "Usage:")
	assert.Contains(t, stdout.String(), "Available Commands:")
}

func TestRunCmdVersion(t *testing.T) {
	var stdout bytes.Buffer
	exitCode := lint.RunCmd(&stdout, []string{"--version"})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "duh version")
	assert.Contains(t, stdout.String(), "1.0.0")
}

func TestRunCmdFileNotFound(t *testing.T) {
	var stdout bytes.Buffer
	exitCode := lint.RunCmd(&stdout, []string{"lint", "nonexistent.yaml"})

	assert.Equal(t, 2, exitCode)
	assert.Contains(t, stdout.String(), "Error:")
	assert.Contains(t, stdout.String(), "file not found")
}

func TestRunCmdNoArguments(t *testing.T) {
	var stdout bytes.Buffer
	exitCode := lint.RunCmd(&stdout, []string{"lint"})

	assert.Equal(t, 2, exitCode)
	output := strings.ToLower(stdout.String())
	fmt.Println(output)
	require.Contains(t, output, "error: accepts 1 arg")
}

func TestRunCmdMultipleArguments(t *testing.T) {
	var stdout bytes.Buffer
	exitCode := lint.RunCmd(&stdout, []string{"lint", "file1.yaml", "file2.yaml"})

	assert.Equal(t, 2, exitCode)
	output := strings.ToLower(stdout.String())
	require.Contains(t, output, "error: accepts 1 arg")
}

func TestInitCommandWithDefaultPath(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	err = os.Chdir(tempDir)
	require.NoError(t, err)
	defer func() {
		_ = os.Chdir(originalDir)
	}()

	var stdout bytes.Buffer
	const defaultOutput = "openapi.yaml"
	exitCode := lint.RunCmd(&stdout, []string{"init"})

	require.Equal(t, 0, exitCode)
	require.Contains(t, stdout.String(), "✓ Created DUH-RPC compliant OpenAPI spec")
	require.Contains(t, stdout.String(), defaultOutput)

	content, err := os.ReadFile(defaultOutput)
	require.NoError(t, err)
	require.NotEmpty(t, content)
	require.Contains(t, string(content), "openapi: 3.0.3")
}

func TestInitCommandWithCustomPath(t *testing.T) {
	tempDir := t.TempDir()
	customPath := filepath.Join(tempDir, "custom-api.yaml")

	var stdout bytes.Buffer
	exitCode := lint.RunCmd(&stdout, []string{"init", customPath})

	require.Equal(t, 0, exitCode)
	require.Contains(t, stdout.String(), "✓ Created DUH-RPC compliant OpenAPI spec")
	require.Contains(t, stdout.String(), customPath)

	content, err := os.ReadFile(customPath)
	require.NoError(t, err)
	require.NotEmpty(t, content)
	require.Contains(t, string(content), "DUH-RPC")
}

func TestInitCommandFileAlreadyExists(t *testing.T) {
	tempDir := t.TempDir()
	existingFile := filepath.Join(tempDir, "existing.yaml")

	err := os.WriteFile(existingFile, []byte("existing"), 0644)
	require.NoError(t, err)

	var stdout bytes.Buffer
	exitCode := lint.RunCmd(&stdout, []string{"init", existingFile})

	require.Equal(t, 2, exitCode)
	require.Contains(t, stdout.String(), "Error:")
	require.Contains(t, stdout.String(), "file already exists")
}

func TestInitCommandCreatesNestedDirectories(t *testing.T) {
	tempDir := t.TempDir()
	nestedPath := filepath.Join(tempDir, "api", "v1", "openapi.yaml")

	var stdout bytes.Buffer
	exitCode := lint.RunCmd(&stdout, []string{"init", nestedPath})

	require.Equal(t, 0, exitCode)

	content, err := os.ReadFile(nestedPath)
	require.NoError(t, err)
	require.NotEmpty(t, content)
}

func TestInitGeneratedFilePassesLint(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "openapi.yaml")

	var initStdout bytes.Buffer
	initExitCode := lint.RunCmd(&initStdout, []string{"init", outputPath})
	require.Equal(t, 0, initExitCode)

	var lintStdout bytes.Buffer
	lintExitCode := lint.RunCmd(&lintStdout, []string{"lint", outputPath})
	require.Equal(t, 0, lintExitCode)
	require.Contains(t, lintStdout.String(), "✓")
	require.Contains(t, lintStdout.String(), "DUH-RPC compliant")
}

func TestInitCommandHelp(t *testing.T) {
	var stdout bytes.Buffer
	exitCode := lint.RunCmd(&stdout, []string{"init", "--help"})

	require.Equal(t, 0, exitCode)
	require.Contains(t, stdout.String(), "init")
	require.Contains(t, stdout.String(), "DUH-RPC compliant OpenAPI specification template")
	require.Contains(t, stdout.String(), "openapi.yaml")
}
