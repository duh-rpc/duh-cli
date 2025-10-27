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

func TestGenerateDuhWithFullFlagAndInitSpec(t *testing.T) {
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
	assert.Contains(t, stdout.String(), "Generated 8 file(s)")

	_, err = os.Stat("daemon.go")
	require.NoError(t, err)

	_, err = os.Stat("service.go")
	require.NoError(t, err)

	_, err = os.Stat("api_test.go")
	require.NoError(t, err)

	_, err = os.Stat("Makefile")
	require.NoError(t, err)

	serviceContent, err := os.ReadFile("service.go")
	require.NoError(t, err)
	assert.Contains(t, string(serviceContent), "map[string]*pb.UserResponse")
	assert.Contains(t, string(serviceContent), "func (s *Service) CreateUser")
	assert.Contains(t, string(serviceContent), "func (s *Service) GetUserById")
	assert.Contains(t, string(serviceContent), "func (s *Service) ListUsers")
	assert.Contains(t, string(serviceContent), "func (s *Service) UpdateUser")

	apiTestContent, err := os.ReadFile("api_test.go")
	require.NoError(t, err)
	assert.Contains(t, string(apiTestContent), "func TestCreateUser(t *testing.T)")
	assert.Contains(t, string(apiTestContent), "func TestGetUserById(t *testing.T)")
	assert.Contains(t, string(apiTestContent), "func TestListUsers(t *testing.T)")
	assert.Contains(t, string(apiTestContent), "func TestUpdateUser(t *testing.T)")

	daemonContent, err := os.ReadFile("daemon.go")
	require.NoError(t, err)
	assert.Contains(t, string(daemonContent), "YOU CAN EDIT")

	makefileContent, err := os.ReadFile("Makefile")
	require.NoError(t, err)
	assert.Contains(t, string(makefileContent), "YOU CAN EDIT")
	assert.Contains(t, string(makefileContent), "proto:")
	assert.Contains(t, string(makefileContent), "test:")
}

func TestGenerateDuhWithFullFlagAndCustomSpec(t *testing.T) {
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
	assert.Contains(t, stdout.String(), "Generated 7 file(s)")

	serviceContent, err := os.ReadFile("service.go")
	require.NoError(t, err)
	assert.Contains(t, string(serviceContent), "TODO")
	assert.Contains(t, string(serviceContent), "CodeNotImplemented")

	apiTestContent, err := os.ReadFile("api_test.go")
	require.NoError(t, err)
	assert.Contains(t, string(apiTestContent), "TODO")
	assert.Contains(t, string(apiTestContent), "func TestProductsCreate(t *testing.T)")
}

func TestGenerateDuhWithoutFullFlag(t *testing.T) {
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
	args := []string{"generate", "duh", "openapi.yaml"}
	exitCode := duh.RunCmd(&stdout, args)

	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "Generated 4 file(s)")

	_, err = os.Stat("daemon.go")
	require.Error(t, err)

	_, err = os.Stat("service.go")
	require.Error(t, err)

	_, err = os.Stat("api_test.go")
	require.Error(t, err)

	_, err = os.Stat("Makefile")
	require.Error(t, err)
}

func TestRegenerateWithFullFlagOverwrites(t *testing.T) {
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

	const customContent = "// MY CUSTOM EDIT"
	err = os.WriteFile("service.go", []byte(customContent), 0644)
	require.NoError(t, err)

	var stdout2 bytes.Buffer
	exitCode = duh.RunCmd(&stdout2, args)
	require.Equal(t, 0, exitCode)

	serviceContent, err := os.ReadFile("service.go")
	require.NoError(t, err)
	assert.NotContains(t, string(serviceContent), customContent)
	assert.Contains(t, string(serviceContent), "func (s *Service) CreateUser")
}

func TestMakefileGoesToProjectRoot(t *testing.T) {
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
	args := []string{"generate", "duh", "openapi.yaml", "--output-dir", "api", "--full"}
	exitCode := duh.RunCmd(&stdout, args)
	require.Equal(t, 0, exitCode)

	_, err = os.Stat("Makefile")
	require.NoError(t, err)

	apiFiles := []string{"daemon.go", "service.go", "api_test.go", "server.go", "client.go"}
	for _, file := range apiFiles {
		filePath := filepath.Join("api", file)
		_, err = os.Stat(filePath)
		require.NoError(t, err, "file %s should exist in api directory", file)
	}
}

func TestFullGeneratedCodeFormat(t *testing.T) {
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

	goFiles := []string{"daemon.go", "service.go", "api_test.go", "server.go", "client.go"}
	for _, file := range goFiles {
		content, err := os.ReadFile(file)
		require.NoError(t, err)

		lines := strings.Split(string(content), "\n")
		require.NotEmpty(t, lines)

		firstLine := lines[0]
		assert.True(t, strings.HasPrefix(firstLine, "// Code generated"), "file %s should have generated header", file)

		if file == "daemon.go" || file == "service.go" || file == "api_test.go" {
			assert.Contains(t, firstLine, "YOU CAN EDIT", "file %s should have YOU CAN EDIT marker", file)
		} else {
			assert.Contains(t, firstLine, "DO NOT EDIT", "file %s should have DO NOT EDIT marker", file)
		}
	}
}
