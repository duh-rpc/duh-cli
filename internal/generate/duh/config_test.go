package duh_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	duh "github.com/duh-rpc/duh-cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const simpleValidSpec = `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /v1/users.create:
    post:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateUserRequest'
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/UserResponse'
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
components:
  schemas:
    CreateUserRequest:
      type: object
      properties:
        name:
          type: string
    UserResponse:
      type: object
      properties:
        id:
          type: string
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

const specWithListOp = `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /v1/users.list:
    post:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ListUsersRequest'
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ListUsersResponse'
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
components:
  schemas:
    ListUsersRequest:
      type: object
      properties:
        offset:
          type: integer
        limit:
          type: integer
    ListUsersResponse:
      type: object
      properties:
        users:
          type: array
          items:
            $ref: '#/components/schemas/UserResponse'
        total:
          type: integer
    UserResponse:
      type: object
      properties:
        id:
          type: string
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

func setupTest(t *testing.T, spec string) (string, *bytes.Buffer) {
	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))
	require.NoError(t, os.WriteFile("go.mod", []byte("module github.com/example/test\n"), 0644))

	specPath := filepath.Join(tempDir, "openapi.yaml")
	require.NoError(t, os.WriteFile(specPath, []byte(spec), 0644))

	var stdout bytes.Buffer
	return specPath, &stdout
}

func TestGenerateDuhParsesSimpleSpec(t *testing.T) {
	specPath, stdout := setupTest(t, simpleValidSpec)

	exitCode := duh.RunCmd(stdout, []string{"generate", "duh", specPath})

	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "✓")
	assert.Contains(t, stdout.String(), "Operations: 1")
	assert.Contains(t, stdout.String(), "List operations: 0")
}

func TestGenerateDuhParsesListOperation(t *testing.T) {
	specPath, stdout := setupTest(t, specWithListOp)

	exitCode := duh.RunCmd(stdout, []string{"generate", "duh", specPath})

	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "Operations: 1")
	assert.Contains(t, stdout.String(), "List operations: 1")
}

func TestGenerateDuhWithCustomPackage(t *testing.T) {
	specPath, stdout := setupTest(t, simpleValidSpec)

	exitCode := duh.RunCmd(stdout, []string{"generate", "duh", specPath, "-p", "myapi"})

	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "Package: myapi")
}

func TestGenerateDuhRejectsMainPackage(t *testing.T) {
	specPath, stdout := setupTest(t, simpleValidSpec)

	exitCode := duh.RunCmd(stdout, []string{"generate", "duh", specPath, "-p", "main"})

	require.Equal(t, 2, exitCode)
	assert.Contains(t, stdout.String(), "package name cannot be 'main'")
}

func TestGenerateDuhRejectsInvalidPackage(t *testing.T) {
	specPath, stdout := setupTest(t, simpleValidSpec)

	exitCode := duh.RunCmd(stdout, []string{"generate", "duh", specPath, "-p", "my-api"})

	require.Equal(t, 2, exitCode)
	assert.Contains(t, stdout.String(), "invalid package name")
}

func TestGenerateDuhDetectsModulePath(t *testing.T) {
	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))
	require.NoError(t, os.WriteFile("go.mod", []byte("module github.com/example/project\n"), 0644))

	specPath := filepath.Join(tempDir, "openapi.yaml")
	require.NoError(t, os.WriteFile(specPath, []byte(simpleValidSpec), 0644))

	var stdout bytes.Buffer
	exitCode := duh.RunCmd(&stdout, []string{"generate", "duh", specPath})

	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "github.com/example/project")
}

func TestGenerateDuhMissingGoMod(t *testing.T) {
	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))

	specPath := filepath.Join(tempDir, "openapi.yaml")
	require.NoError(t, os.WriteFile(specPath, []byte(simpleValidSpec), 0644))

	var stdout bytes.Buffer
	exitCode := duh.RunCmd(&stdout, []string{"generate", "duh", specPath})

	require.Equal(t, 2, exitCode)
	assert.Contains(t, stdout.String(), "failed to read go.mod")
}
