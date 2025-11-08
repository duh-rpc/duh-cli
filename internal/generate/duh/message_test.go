package duh_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/duh-rpc/duh-cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReminderMessageDisplayed(t *testing.T) {
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
	exitCode := duh.RunCmd(&stdout, []string{"generate", "openapi.yaml"})

	require.Equal(t, 0, exitCode)

	output := stdout.String()
	assert.Contains(t, output, "Next steps:")
	assert.Contains(t, output, "buf generate")
	assert.Contains(t, output, "go mod tidy")
}

func TestReminderMessageWithFullFlag(t *testing.T) {
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
	exitCode := duh.RunCmd(&stdout, []string{"generate", "openapi.yaml", "--full"})

	require.Equal(t, 0, exitCode)

	output := stdout.String()
	assert.Contains(t, output, "Next steps:")
	assert.Contains(t, output, "buf generate")
	assert.Contains(t, output, "go mod tidy")
}

func TestReminderMessageFormat(t *testing.T) {
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
	exitCode := duh.RunCmd(&stdout, []string{"generate", "openapi.yaml"})

	require.Equal(t, 0, exitCode)

	output := stdout.String()
	lines := strings.Split(output, "\n")

	foundNextSteps := false
	foundStep1 := false
	foundStep2 := false
	blankLineBeforeNextSteps := false

	for i, line := range lines {
		if strings.Contains(line, "Next steps:") {
			foundNextSteps = true
			if i > 0 && strings.TrimSpace(lines[i-1]) == "" {
				blankLineBeforeNextSteps = true
			}
		}
		if strings.Contains(line, "1. Run 'buf generate'") {
			foundStep1 = true
		}
		if strings.Contains(line, "2. Run 'go mod tidy'") {
			foundStep2 = true
		}
	}

	assert.True(t, foundNextSteps)
	assert.True(t, foundStep1)
	assert.True(t, foundStep2)
	assert.True(t, blankLineBeforeNextSteps)
}
