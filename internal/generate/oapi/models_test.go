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

func TestGenerateModelsWithDefaults(t *testing.T) {
	tempDir := t.TempDir()
	specPath := filepath.Join(tempDir, "openapi.yaml")
	outputPath := filepath.Join(tempDir, "models.go")

	require.NoError(t, os.WriteFile(specPath, []byte(validSpec), 0644))

	var stdout bytes.Buffer
	exitCode := duh.RunCmd(&stdout, []string{"generate", "models", specPath, "-o", outputPath})

	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "✓")

	_, err := os.Stat(outputPath)
	require.NoError(t, err)

	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "package api")
}

func TestGenerateModelsWithCustomOutput(t *testing.T) {
	tempDir := t.TempDir()
	specPath := filepath.Join(tempDir, "openapi.yaml")
	customOutput := filepath.Join(tempDir, "custom", "mymodels.go")

	require.NoError(t, os.WriteFile(specPath, []byte(validSpec), 0644))

	var stdout bytes.Buffer
	exitCode := duh.RunCmd(&stdout, []string{"generate", "models", specPath, "-o", customOutput})

	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), customOutput)

	_, err := os.Stat(customOutput)
	require.NoError(t, err)
}

func TestGenerateModelsWithCustomPackage(t *testing.T) {
	tempDir := t.TempDir()
	specPath := filepath.Join(tempDir, "openapi.yaml")
	outputPath := filepath.Join(tempDir, "models.go")

	require.NoError(t, os.WriteFile(specPath, []byte(validSpec), 0644))

	var stdout bytes.Buffer
	exitCode := duh.RunCmd(&stdout, []string{"generate", "models", specPath, "-o", outputPath, "-p", "mymodels"})

	require.Equal(t, 0, exitCode)

	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "package mymodels")
}
