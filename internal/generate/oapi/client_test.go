package oapi_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	duh "github.com/duh-rpc/duh-cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const validSpec = `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      operationId: getTest
      responses:
        200:
          description: Success
`

func TestGenerateClientWithDefaults(t *testing.T) {
	tempDir := t.TempDir()
	specPath := filepath.Join(tempDir, "openapi.yaml")
	outputPath := filepath.Join(tempDir, "client.go")

	require.NoError(t, os.WriteFile(specPath, []byte(validSpec), 0644))

	var stdout bytes.Buffer
	exitCode := duh.RunCmd(&stdout, []string{"generate", "client", specPath, "-o", outputPath})

	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "✓")

	_, err := os.Stat(outputPath)
	require.NoError(t, err)

	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "package api")
}

func TestGenerateClientWithCustomOutput(t *testing.T) {
	tempDir := t.TempDir()
	specPath := filepath.Join(tempDir, "openapi.yaml")
	customOutput := filepath.Join(tempDir, "custom", "myclient.go")

	require.NoError(t, os.WriteFile(specPath, []byte(validSpec), 0644))

	var stdout bytes.Buffer
	exitCode := duh.RunCmd(&stdout, []string{"generate", "client", specPath, "-o", customOutput})

	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), customOutput)

	_, err := os.Stat(customOutput)
	require.NoError(t, err)
}

func TestGenerateClientWithCustomPackage(t *testing.T) {
	tempDir := t.TempDir()
	specPath := filepath.Join(tempDir, "openapi.yaml")
	outputPath := filepath.Join(tempDir, "client.go")

	require.NoError(t, os.WriteFile(specPath, []byte(validSpec), 0644))

	var stdout bytes.Buffer
	exitCode := duh.RunCmd(&stdout, []string{"generate", "client", specPath, "-o", outputPath, "-p", "myclient"})

	require.Equal(t, 0, exitCode)

	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "package myclient")
}

func TestGenerateClientFileNotFound(t *testing.T) {
	tempDir := t.TempDir()
	nonexistentPath := filepath.Join(tempDir, "nonexistent.yaml")
	outputPath := filepath.Join(tempDir, "client.go")

	var stdout bytes.Buffer
	exitCode := duh.RunCmd(&stdout, []string{"generate", "client", nonexistentPath, "-o", outputPath})

	require.Equal(t, 2, exitCode)
	assert.Contains(t, stdout.String(), "file not found")
}

func TestGenerateClientInvalidSpec(t *testing.T) {
	tempDir := t.TempDir()
	specPath := filepath.Join(tempDir, "invalid.yaml")
	outputPath := filepath.Join(tempDir, "client.go")

	invalidSpec := `this is not valid yaml: [[[`
	require.NoError(t, os.WriteFile(specPath, []byte(invalidSpec), 0644))

	var stdout bytes.Buffer
	exitCode := duh.RunCmd(&stdout, []string{"generate", "client", specPath, "-o", outputPath})

	require.Equal(t, 2, exitCode)
	assert.Contains(t, stdout.String(), "failed to parse OpenAPI spec")
}
