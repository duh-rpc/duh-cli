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

func TestGenerateServerCommand(t *testing.T) {
	tempDir := t.TempDir()
	specPath := filepath.Join(tempDir, "openapi.yaml")
	outputPath := filepath.Join(tempDir, "server.go")

	require.NoError(t, os.WriteFile(specPath, []byte(validSpec), 0644))

	var stdout bytes.Buffer
	exitCode := duh.RunCmd(&stdout, []string{"generate", "server", specPath, "-o", outputPath})

	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "✓")

	_, err := os.Stat(outputPath)
	require.NoError(t, err)
}

func TestGenerateServerWithFlags(t *testing.T) {
	tempDir := t.TempDir()
	specPath := filepath.Join(tempDir, "openapi.yaml")
	outputPath := filepath.Join(tempDir, "myserver.go")

	require.NoError(t, os.WriteFile(specPath, []byte(validSpec), 0644))

	var stdout bytes.Buffer
	exitCode := duh.RunCmd(&stdout, []string{"generate", "server", specPath, "-o", outputPath, "-p", "myserver"})

	require.Equal(t, 0, exitCode)

	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "package myserver")
}
