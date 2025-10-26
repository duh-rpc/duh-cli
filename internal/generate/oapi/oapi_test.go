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

func TestGenerateOapiWithDefaults(t *testing.T) {
	tempDir := t.TempDir()
	specPath := filepath.Join(tempDir, "openapi.yaml")

	require.NoError(t, os.WriteFile(specPath, []byte(validSpec), 0644))

	var stdout bytes.Buffer
	exitCode := duh.RunCmd(&stdout, []string{"generate", "oapi", specPath, "--output-dir", tempDir})

	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "✓")

	clientPath := filepath.Join(tempDir, "client.go")
	_, err := os.Stat(clientPath)
	require.NoError(t, err)

	serverPath := filepath.Join(tempDir, "server.go")
	_, err = os.Stat(serverPath)
	require.NoError(t, err)

	modelsPath := filepath.Join(tempDir, "models.go")
	_, err = os.Stat(modelsPath)
	require.NoError(t, err)

	clientContent, err := os.ReadFile(clientPath)
	require.NoError(t, err)
	assert.Contains(t, string(clientContent), "package api")

	serverContent, err := os.ReadFile(serverPath)
	require.NoError(t, err)
	assert.Contains(t, string(serverContent), "package api")

	modelsContent, err := os.ReadFile(modelsPath)
	require.NoError(t, err)
	assert.Contains(t, string(modelsContent), "package api")
}

func TestGenerateOapiWithCustomOutputDir(t *testing.T) {
	tempDir := t.TempDir()
	specPath := filepath.Join(tempDir, "openapi.yaml")
	outputDir := filepath.Join(tempDir, "api")

	require.NoError(t, os.WriteFile(specPath, []byte(validSpec), 0644))

	var stdout bytes.Buffer
	exitCode := duh.RunCmd(&stdout, []string{"generate", "oapi", specPath, "--output-dir", outputDir})

	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), outputDir)

	_, err := os.Stat(filepath.Join(outputDir, "client.go"))
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(outputDir, "server.go"))
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(outputDir, "models.go"))
	require.NoError(t, err)
}

func TestGenerateOapiWithCustomPackage(t *testing.T) {
	tempDir := t.TempDir()
	specPath := filepath.Join(tempDir, "openapi.yaml")

	require.NoError(t, os.WriteFile(specPath, []byte(validSpec), 0644))

	var stdout bytes.Buffer
	exitCode := duh.RunCmd(&stdout, []string{"generate", "oapi", specPath, "--output-dir", tempDir, "-p", "myapi"})

	require.Equal(t, 0, exitCode)

	clientContent, err := os.ReadFile(filepath.Join(tempDir, "client.go"))
	require.NoError(t, err)
	assert.Contains(t, string(clientContent), "package myapi")

	serverContent, err := os.ReadFile(filepath.Join(tempDir, "server.go"))
	require.NoError(t, err)
	assert.Contains(t, string(serverContent), "package myapi")

	modelsContent, err := os.ReadFile(filepath.Join(tempDir, "models.go"))
	require.NoError(t, err)
	assert.Contains(t, string(modelsContent), "package myapi")
}

func TestGenerateOapiCreatesDirectory(t *testing.T) {
	tempDir := t.TempDir()
	specPath := filepath.Join(tempDir, "openapi.yaml")
	outputDir := filepath.Join(tempDir, "nested", "api")

	require.NoError(t, os.WriteFile(specPath, []byte(validSpec), 0644))

	var stdout bytes.Buffer
	exitCode := duh.RunCmd(&stdout, []string{"generate", "oapi", specPath, "--output-dir", outputDir})

	require.Equal(t, 0, exitCode)

	_, err := os.Stat(outputDir)
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(outputDir, "client.go"))
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(outputDir, "server.go"))
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(outputDir, "models.go"))
	require.NoError(t, err)
}
