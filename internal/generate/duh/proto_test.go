package duh_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	duh "github.com/duh-rpc/duh-cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const multiSchemaSpec = `openapi: 3.0.0
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
  /v1/users.get:
    post:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/GetUserRequest'
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
    GetUserRequest:
      type: object
      properties:
        id:
          type: string
    UserResponse:
      type: object
      properties:
        id:
          type: string
        name:
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

func TestGenerateDuhCreatesProtoFile(t *testing.T) {
	specPath, stdout := setupTest(t, simpleValidSpec)
	tempDir := filepath.Dir(specPath)

	exitCode := duh.RunCmd(stdout, []string{"generate", "duh", specPath})

	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "✓")
	assert.Contains(t, stdout.String(), "proto/v1/api.proto")

	protoPath := filepath.Join(tempDir, "proto/v1/api.proto")
	_, err := os.Stat(protoPath)
	require.NoError(t, err)

	protoContent, err := os.ReadFile(protoPath)
	require.NoError(t, err)

	assert.Contains(t, string(protoContent), "syntax = \"proto3\"")
	assert.Contains(t, string(protoContent), "package duh.api.v1")
	assert.Contains(t, string(protoContent), "message CreateUserRequest")
	assert.Contains(t, string(protoContent), "message UserResponse")
	assert.Contains(t, string(protoContent), "message Error")
}

func TestProtoFileStructure(t *testing.T) {
	specPath, stdout := setupTest(t, simpleValidSpec)
	tempDir := filepath.Dir(specPath)

	exitCode := duh.RunCmd(stdout, []string{"generate", "duh", specPath})

	require.Equal(t, 0, exitCode)

	protoPath := filepath.Join(tempDir, "proto/v1/api.proto")
	protoContent, err := os.ReadFile(protoPath)
	require.NoError(t, err)

	content := string(protoContent)

	assert.True(t, strings.HasPrefix(content, "syntax = \"proto3\""))

	assert.Contains(t, content, "package duh.api.v1")

	assert.Contains(t, content, "message CreateUserRequest {}")
	assert.Contains(t, content, "message UserResponse {}")
	assert.Contains(t, content, "message Error {}")

	lines := strings.Split(content, "\n")
	messageCount := 0
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "message ") {
			messageCount++
		}
	}
	assert.Equal(t, 3, messageCount)
}

func TestProtoWithCustomPath(t *testing.T) {
	specPath, stdout := setupTest(t, simpleValidSpec)
	tempDir := filepath.Dir(specPath)

	exitCode := duh.RunCmd(stdout, []string{
		"generate", "duh", specPath,
		"--proto-path", "custom/path/api.proto",
	})

	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "custom/path/api.proto")

	protoPath := filepath.Join(tempDir, "custom/path/api.proto")
	_, err := os.Stat(protoPath)
	require.NoError(t, err)

	protoContent, err := os.ReadFile(protoPath)
	require.NoError(t, err)
	assert.Contains(t, string(protoContent), "syntax = \"proto3\"")
}

func TestProtoSchemaExtraction(t *testing.T) {
	specPath, stdout := setupTest(t, multiSchemaSpec)
	tempDir := filepath.Dir(specPath)

	exitCode := duh.RunCmd(stdout, []string{"generate", "duh", specPath})

	require.Equal(t, 0, exitCode)

	protoPath := filepath.Join(tempDir, "proto/v1/api.proto")
	protoContent, err := os.ReadFile(protoPath)
	require.NoError(t, err)

	content := string(protoContent)

	assert.Contains(t, content, "message CreateUserRequest {}")
	assert.Contains(t, content, "message GetUserRequest {}")
	assert.Contains(t, content, "message UserResponse {}")
	assert.Contains(t, content, "message Error {}")

	seenSchemas := make(map[string]int)
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "message ") {
			schemaName := strings.TrimSuffix(strings.TrimPrefix(trimmed, "message "), " {}")
			seenSchemas[schemaName]++
		}
	}

	for schema, count := range seenSchemas {
		assert.Equal(t, 1, count, "Schema %s should appear exactly once, but appeared %d times", schema, count)
	}
}
