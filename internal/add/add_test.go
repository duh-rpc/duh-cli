package add_test

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

const minimalOpenAPI = `openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
paths: {}
components:
  schemas:
    Error:
      type: object
      required:
        - code
        - message
      properties:
        code:
          type: integer
        message:
          type: string
`

func TestAddCommandWithDefaultFile(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	err = os.Chdir(tempDir)
	require.NoError(t, err)
	defer func() {
		_ = os.Chdir(originalDir)
	}()

	const defaultOutput = "openapi.yaml"
	err = os.WriteFile(defaultOutput, []byte(minimalOpenAPI), 0644)
	require.NoError(t, err)

	var stdout bytes.Buffer
	exitCode := duh.RunCmd(&stdout, []string{"add", "/v1/users.create", "CreateUser"})

	require.Equal(t, 0, exitCode)
	require.Contains(t, stdout.String(), "✓ Added endpoint /v1/users.create")

	content, err := os.ReadFile(defaultOutput)
	require.NoError(t, err)
	require.Contains(t, string(content), "/v1/users.create")
	require.Contains(t, string(content), "CreateUserRequest")
	require.Contains(t, string(content), "CreateUserResponse")
}

func TestAddCommandWithFFlag(t *testing.T) {
	tempDir := t.TempDir()
	customPath := filepath.Join(tempDir, "custom-api.yaml")

	err := os.WriteFile(customPath, []byte(minimalOpenAPI), 0644)
	require.NoError(t, err)

	var stdout bytes.Buffer
	exitCode := duh.RunCmd(&stdout, []string{"add", "-f", customPath, "/v2/products.list", "ListProducts"})

	require.Equal(t, 0, exitCode)
	require.Contains(t, stdout.String(), "✓ Added endpoint /v2/products.list")
	require.Contains(t, stdout.String(), customPath)

	content, err := os.ReadFile(customPath)
	require.NoError(t, err)
	require.Contains(t, string(content), "/v2/products.list")
	require.Contains(t, string(content), "ListProductsRequest")
	require.Contains(t, string(content), "ListProductsResponse")
	require.Contains(t, string(content), "post:")
	require.Contains(t, string(content), "application/json")
}

func TestAddCommandDuplicatePath(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "openapi.yaml")

	existingAPI := `openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
paths:
  /v1/users.create:
    post:
      summary: Create user
      responses:
        '200':
          description: Success
components:
  schemas: {}
`
	err := os.WriteFile(filePath, []byte(existingAPI), 0644)
	require.NoError(t, err)

	var stdout bytes.Buffer
	exitCode := duh.RunCmd(&stdout, []string{"add", "-f", filePath, "/v1/users.create", "CreateUser"})

	require.Equal(t, 2, exitCode)
	require.Contains(t, stdout.String(), "Error:")
	require.Contains(t, stdout.String(), "path already exists")
}

func TestAddCommandInvalidPath(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "openapi.yaml")

	err := os.WriteFile(filePath, []byte(minimalOpenAPI), 0644)
	require.NoError(t, err)

	for _, test := range []struct {
		path    string
		name    string
		wantErr string
	}{
		{
			path:    "/users/create",
			name:    "CreateUser",
			wantErr: "invalid path format",
		},
		{
			path:    "/v1/users",
			name:    "CreateUser",
			wantErr: "invalid path format",
		},
		{
			path:    "/v1/Users.Create",
			name:    "CreateUser",
			wantErr: "invalid path format",
		},
		{
			path:    "v1/users.create",
			name:    "CreateUser",
			wantErr: "invalid path format",
		},
	} {
		var stdout bytes.Buffer
		exitCode := duh.RunCmd(&stdout, []string{"add", "-f", filePath, test.path, test.name})

		assert.Equal(t, 2, exitCode)
		assert.Contains(t, stdout.String(), "Error:")
		assert.Contains(t, stdout.String(), test.wantErr)
	}
}

func TestAddCommandFileNotFound(t *testing.T) {
	tempDir := t.TempDir()
	nonexistentFile := filepath.Join(tempDir, "nonexistent.yaml")

	var stdout bytes.Buffer
	exitCode := duh.RunCmd(&stdout, []string{"add", "-f", nonexistentFile, "/v1/users.create", "CreateUser"})

	require.Equal(t, 2, exitCode)
	require.Contains(t, stdout.String(), "Error:")
	require.Contains(t, stdout.String(), "file not found")
}

func TestAddCommandNoArguments(t *testing.T) {
	var stdout bytes.Buffer
	exitCode := duh.RunCmd(&stdout, []string{"add"})

	require.Equal(t, 2, exitCode)
	output := strings.ToLower(stdout.String())
	require.Contains(t, output, "error: accepts 2 arg")
}

func TestAddCommandHelp(t *testing.T) {
	var stdout bytes.Buffer
	exitCode := duh.RunCmd(&stdout, []string{"add", "--help"})

	require.Equal(t, 0, exitCode)
	require.Contains(t, stdout.String(), "add <path> <name>")
	require.Contains(t, stdout.String(), "Add a new DUH-RPC endpoint")
	require.Contains(t, stdout.String(), "-f")
	require.Contains(t, stdout.String(), "openapi.yaml")
}

func TestAddGeneratedEndpointPassesLint(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "openapi.yaml")

	err := os.WriteFile(filePath, []byte(minimalOpenAPI), 0644)
	require.NoError(t, err)

	var addStdout bytes.Buffer
	addExitCode := duh.RunCmd(&addStdout, []string{"add", "-f", filePath, "/v1/orders.update", "UpdateOrder"})
	require.Equal(t, 0, addExitCode)

	var lintStdout bytes.Buffer
	lintExitCode := duh.RunCmd(&lintStdout, []string{"lint", filePath})
	require.Equal(t, 0, lintExitCode)
	require.Contains(t, lintStdout.String(), "✓")
	require.Contains(t, lintStdout.String(), "DUH-RPC compliant")
}

func TestAddCommandMultipleEndpoints(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "openapi.yaml")

	err := os.WriteFile(filePath, []byte(minimalOpenAPI), 0644)
	require.NoError(t, err)

	var stdout1 bytes.Buffer
	exitCode1 := duh.RunCmd(&stdout1, []string{"add", "-f", filePath, "/v1/users.create", "CreateUser"})
	require.Equal(t, 0, exitCode1)

	var stdout2 bytes.Buffer
	exitCode2 := duh.RunCmd(&stdout2, []string{"add", "-f", filePath, "/v1/users.get", "GetUser"})
	require.Equal(t, 0, exitCode2)

	content, err := os.ReadFile(filePath)
	require.NoError(t, err)
	require.Contains(t, string(content), "/v1/users.create")
	require.Contains(t, string(content), "/v1/users.get")
	require.Contains(t, string(content), "CreateUserRequest")
	require.Contains(t, string(content), "GetUserRequest")
}

func TestAddCommandVerifyResponseStructure(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "openapi.yaml")

	err := os.WriteFile(filePath, []byte(minimalOpenAPI), 0644)
	require.NoError(t, err)

	var stdout bytes.Buffer
	exitCode := duh.RunCmd(&stdout, []string{"add", "-f", filePath, "/v1/products.create", "CreateProduct"})
	require.Equal(t, 0, exitCode)

	content, err := os.ReadFile(filePath)
	require.NoError(t, err)
	contentStr := string(content)

	require.Contains(t, contentStr, "'200'")
	require.Contains(t, contentStr, "'400'")
	require.Contains(t, contentStr, "'404'")
	require.Contains(t, contentStr, "'500'")

	require.Contains(t, contentStr, "$ref: '#/components/schemas/CreateProductRequest'")
	require.Contains(t, contentStr, "$ref: '#/components/schemas/CreateProductResponse'")
	require.Contains(t, contentStr, "$ref: '#/components/schemas/Error'")

	require.Contains(t, contentStr, "id:")
	require.Contains(t, contentStr, "name:")
	require.Contains(t, contentStr, "example:")
}
