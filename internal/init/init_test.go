package init_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	init_ "github.com/duh-rpc/duh-cli/internal/init"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "openapi.yaml")

	var stdout bytes.Buffer
	err := init_.Run(&stdout, outputPath)
	require.NoError(t, err)
	require.Contains(t, stdout.String(), "✓ Created DUH-RPC compliant OpenAPI spec")

	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	require.NotEmpty(t, content)
	require.Contains(t, string(content), "DUH-RPC")
	require.Contains(t, string(content), "openapi: 3.0.3")
}

func TestRunWithCustomPath(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "custom", "my-api.yaml")

	var stdout bytes.Buffer
	err := init_.Run(&stdout, outputPath)
	require.NoError(t, err)

	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	require.NotEmpty(t, content)
}

func TestRunErrorWhenFileExists(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "openapi.yaml")

	err := os.WriteFile(outputPath, []byte("existing content"), 0644)
	require.NoError(t, err)

	var stdout bytes.Buffer
	err = init_.Run(&stdout, outputPath)
	require.Error(t, err)
	require.ErrorContains(t, err, "file already exists")
}

func TestRunCreatesParentDirectory(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "nested", "dir", "openapi.yaml")

	var stdout bytes.Buffer
	err := init_.Run(&stdout, outputPath)
	require.NoError(t, err)

	info, err := os.Stat(filepath.Join(tempDir, "nested", "dir"))
	require.NoError(t, err)
	require.True(t, info.IsDir())

	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	require.NotEmpty(t, content)
}

func TestTemplateIsNotEmpty(t *testing.T) {
	require.NotEmpty(t, init_.Template)
	require.Contains(t, string(init_.Template), "openapi:")
	require.Contains(t, string(init_.Template), "paths:")
	require.Contains(t, string(init_.Template), "components:")
}
