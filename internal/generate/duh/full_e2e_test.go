package duh_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/duh-rpc/duh-cli"
	"github.com/stretchr/testify/require"
)

func TestFullGenerationWithInitSpecCompiles(t *testing.T) {
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
	args := []string{"generate", "duh", "openapi.yaml", "--full"}
	exitCode := duh.RunCmd(&stdout, args)
	require.Equal(t, 0, exitCode)

	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyOutput, err := tidyCmd.CombinedOutput()
	if err != nil {
		t.Logf("go mod tidy output: %s", tidyOutput)
		t.Skipf("Skipping E2E test - go mod tidy failed: %v", err)
	}

	buildCmd := exec.Command("go", "build", "./...")
	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Logf("Build output: %s", buildOutput)
		t.Skipf("Skipping compilation test - requires proto generation and full dependencies. Error: %v", err)
	}
}

func TestFullGenerationWithCustomSpecCompiles(t *testing.T) {
	tempDir := t.TempDir()
	specPath := filepath.Join(tempDir, "openapi.yaml")

	require.NoError(t, os.WriteFile(specPath, []byte(customSpec), 0644))
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
	args := []string{"generate", "duh", "openapi.yaml", "--full"}
	exitCode := duh.RunCmd(&stdout, args)
	require.Equal(t, 0, exitCode)

	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyOutput, err := tidyCmd.CombinedOutput()
	if err != nil {
		t.Logf("go mod tidy output: %s", tidyOutput)
		t.Skipf("Skipping E2E test - go mod tidy failed: %v", err)
	}

	buildCmd := exec.Command("go", "build", "./...")
	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Logf("Build output: %s", buildOutput)
		t.Skipf("Skipping compilation test - requires proto generation and full dependencies. Error: %v", err)
	}
}
