package init_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/duh-rpc/duh-cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitDefaultPath(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(originalDir) })

	require.NoError(t, os.Chdir(tempDir))

	var stdout bytes.Buffer

	exitCode := duh.RunCmd(&stdout, []string{"init"})

	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "✓")
	assert.Contains(t, stdout.String(), "Created DUH-RPC compliant OpenAPI spec")

	content, err := os.ReadFile("openapi.yaml")
	require.NoError(t, err)
	require.NotEmpty(t, content)
	assert.Contains(t, string(content), "DUH-RPC")
	assert.Contains(t, string(content), "openapi: 3.0.3")
}

func TestInitCustomPath(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "custom", "my-api.yaml")

	var stdout bytes.Buffer

	exitCode := duh.RunCmd(&stdout, []string{"init", outputPath})

	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "✓")

	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	require.NotEmpty(t, content)
}

func TestInitFileAlreadyExists(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "openapi.yaml")

	require.NoError(t, os.WriteFile(outputPath, []byte("existing content"), 0644))

	var stdout bytes.Buffer

	exitCode := duh.RunCmd(&stdout, []string{"init", outputPath})

	require.Equal(t, 2, exitCode)
	assert.Contains(t, stdout.String(), "file already exists")
}

func TestInitCreatesParentDirectory(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "nested", "dir", "openapi.yaml")

	var stdout bytes.Buffer

	exitCode := duh.RunCmd(&stdout, []string{"init", outputPath})

	require.Equal(t, 0, exitCode)

	info, err := os.Stat(filepath.Join(tempDir, "nested", "dir"))
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	require.NotEmpty(t, content)
}
