package duhrpc_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	lint "github.com/duh-rpc/duhrpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunCmdHelp(t *testing.T) {
	var stdout bytes.Buffer
	exitCode := lint.RunCmd(&stdout, []string{"--help"})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "duhrpc is a command-line tool")
	assert.Contains(t, stdout.String(), "Usage:")
	assert.Contains(t, stdout.String(), "Available Commands:")
}

func TestRunCmdVersion(t *testing.T) {
	var stdout bytes.Buffer
	exitCode := lint.RunCmd(&stdout, []string{"--version"})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "duhrpc version")
	assert.Contains(t, stdout.String(), "1.0.0")
}

func TestRunCmdFileNotFound(t *testing.T) {
	var stdout bytes.Buffer
	exitCode := lint.RunCmd(&stdout, []string{"lint", "nonexistent.yaml"})

	assert.Equal(t, 2, exitCode)
	assert.Contains(t, stdout.String(), "Error:")
	assert.Contains(t, stdout.String(), "file not found")
}

func TestRunCmdNoArguments(t *testing.T) {
	var stdout bytes.Buffer
	exitCode := lint.RunCmd(&stdout, []string{"lint"})

	assert.Equal(t, 2, exitCode)
	output := strings.ToLower(stdout.String())
	fmt.Println(output)
	require.Contains(t, output, "error: accepts 1 arg")
}

func TestRunCmdMultipleArguments(t *testing.T) {
	var stdout bytes.Buffer
	exitCode := lint.RunCmd(&stdout, []string{"lint", "file1.yaml", "file2.yaml"})

	assert.Equal(t, 2, exitCode)
	output := strings.ToLower(stdout.String())
	require.Contains(t, output, "error: accepts 1 arg")
}
