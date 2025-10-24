package duh_test

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

func TestGenerateClientCommand(t *testing.T) {
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
}

func TestGenerateClientWithFlags(t *testing.T) {
	tempDir := t.TempDir()
	specPath := filepath.Join(tempDir, "openapi.yaml")
	outputPath := filepath.Join(tempDir, "myclient.go")

	require.NoError(t, os.WriteFile(specPath, []byte(validSpec), 0644))

	var stdout bytes.Buffer
	exitCode := duh.RunCmd(&stdout, []string{"generate", "client", specPath, "-o", outputPath, "-p", "myclient"})

	require.Equal(t, 0, exitCode)

	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "package myclient")
}

func TestGenerateClientHelp(t *testing.T) {
	var stdout bytes.Buffer
	exitCode := duh.RunCmd(&stdout, []string{"generate", "client", "--help"})

	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "Generate HTTP client code")
	assert.Contains(t, stdout.String(), "--output")
	assert.Contains(t, stdout.String(), "--package")
}
