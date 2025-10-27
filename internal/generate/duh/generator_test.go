package duh_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/duh-rpc/duh-cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderBufGenYaml(t *testing.T) {
	tempDir := t.TempDir()
	specPath := filepath.Join(tempDir, "openapi.yaml")

	require.NoError(t, os.WriteFile(specPath, []byte(initTemplateSpec), 0644))
	require.NoError(t, os.WriteFile(
		filepath.Join(tempDir, "go.mod"),
		[]byte("module github.com/test/example\n\ngo 1.24\n"),
		0644,
	))

	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()
	require.NoError(t, os.Chdir(tempDir))

	var stdout bytes.Buffer
	exitCode := duh.RunCmd(&stdout, []string{"generate", "duh", "openapi.yaml"})

	require.Equal(t, 0, exitCode)

	bufGenContent, err := os.ReadFile("buf.gen.yaml")
	require.NoError(t, err)

	content := string(bufGenContent)
	assert.Contains(t, content, "version: v2")
	assert.Contains(t, content, "buf.build/protocolbuffers/go")
	assert.Contains(t, content, "buf.build/grpc/go")
	assert.Contains(t, content, "paths=source_relative")
}

func TestBufFilesGeneratedWithoutFullFlag(t *testing.T) {
	tempDir := t.TempDir()
	specPath := filepath.Join(tempDir, "openapi.yaml")

	require.NoError(t, os.WriteFile(specPath, []byte(initTemplateSpec), 0644))
	require.NoError(t, os.WriteFile(
		filepath.Join(tempDir, "go.mod"),
		[]byte("module github.com/test/example\n\ngo 1.24\n"),
		0644,
	))

	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()
	require.NoError(t, os.Chdir(tempDir))

	var stdout bytes.Buffer
	exitCode := duh.RunCmd(&stdout, []string{"generate", "duh", "openapi.yaml"})

	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "Generated 6 file(s)")

	_, err = os.Stat("buf.yaml")
	require.NoError(t, err)

	_, err = os.Stat("buf.gen.yaml")
	require.NoError(t, err)

	bufGenContent, err := os.ReadFile("buf.gen.yaml")
	require.NoError(t, err)
	assert.Contains(t, string(bufGenContent), "buf.build/protocolbuffers/go")
}

func TestBufFilesGeneratedWithFullFlag(t *testing.T) {
	tempDir := t.TempDir()
	specPath := filepath.Join(tempDir, "openapi.yaml")

	require.NoError(t, os.WriteFile(specPath, []byte(initTemplateSpec), 0644))
	require.NoError(t, os.WriteFile(
		filepath.Join(tempDir, "go.mod"),
		[]byte("module github.com/test/example\n\ngo 1.24\n"),
		0644,
	))

	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()
	require.NoError(t, os.Chdir(tempDir))

	var stdout bytes.Buffer
	exitCode := duh.RunCmd(&stdout, []string{"generate", "duh", "openapi.yaml", "--full"})

	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "Generated 10 file(s)")

	_, err = os.Stat("buf.yaml")
	require.NoError(t, err)

	_, err = os.Stat("buf.gen.yaml")
	require.NoError(t, err)

	bufGenContent, err := os.ReadFile("buf.gen.yaml")
	require.NoError(t, err)
	assert.Contains(t, string(bufGenContent), "buf.build/protocolbuffers/go")
	assert.Contains(t, string(bufGenContent), "buf.build/grpc/go")
}

func TestMakefileWrittenToOutputDir(t *testing.T) {
	tempDir := t.TempDir()
	specPath := filepath.Join(tempDir, "openapi.yaml")
	outputDir := filepath.Join(tempDir, "api")

	require.NoError(t, os.MkdirAll(outputDir, 0755))
	require.NoError(t, os.WriteFile(specPath, []byte(initTemplateSpec), 0644))
	require.NoError(t, os.WriteFile(
		filepath.Join(tempDir, "go.mod"),
		[]byte("module github.com/test/example\n\ngo 1.24\n"),
		0644,
	))

	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()
	require.NoError(t, os.Chdir(tempDir))

	var stdout bytes.Buffer
	exitCode := duh.RunCmd(&stdout, []string{"generate", "duh", "openapi.yaml", "--output-dir", "api", "--full"})
	require.Equal(t, 0, exitCode)

	_, err = os.Stat(filepath.Join("api", "Makefile"))
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join("api", "buf.yaml"))
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join("api", "buf.gen.yaml"))
	require.NoError(t, err)

	_, err = os.Stat("Makefile")
	require.Error(t, err)
}
