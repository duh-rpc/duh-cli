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
servers:
  - url: https://api.example.com/v1
paths:
  /users.create:
    post:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateRequest'
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/CreateResponse'
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorDetails'
  /users.get:
    post:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/GetRequest'
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/GetResponse'
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorDetails'
components:
  schemas:
    CreateRequest:
      type: object
      properties:
        name:
          type: string
    GetRequest:
      type: object
      properties:
        id:
          type: string
    CreateResponse:
      type: object
      properties:
        id:
          type: string
        name:
          type: string
    GetResponse:
      type: object
      properties:
        id:
          type: string
    ErrorDetails:
      type: object
      required:
        - message
      properties:
        message:
          type: string
`

func TestGenerateDuhCreatesProtoFile(t *testing.T) {
	specPath, stdout := setupTest(t, simpleValidSpec)
	tempDir := filepath.Dir(specPath)

	exitCode := duh.RunCmd(stdout, []string{"generate", specPath})

	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "✓")
	assert.Contains(t, stdout.String(), "proto/v1/api.proto")

	protoPath := filepath.Join(tempDir, "proto/v1/api.proto")
	_, err := os.Stat(protoPath)
	require.NoError(t, err)

	protoContent, err := os.ReadFile(protoPath)
	require.NoError(t, err)

	content := string(protoContent)

	assert.True(t, strings.HasPrefix(content, "syntax = \"proto3\""))
	assert.Contains(t, content, "option go_package = \"github.com/example/test/proto/v1\";")
	assert.Contains(t, content, "package duh.api.v1")
	assert.Contains(t, content, "message CreateRequest")
	assert.Contains(t, content, "message CreateResponse")
	assert.Contains(t, content, "message ErrorDetails")

	assert.NotContains(t, content, "message CreateRequest {}")
	assert.NotContains(t, content, "message CreateResponse {}")
	assert.NotContains(t, content, "message ErrorDetails {}")

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
		"generate", specPath,
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

	exitCode := duh.RunCmd(stdout, []string{"generate", specPath})

	require.Equal(t, 0, exitCode)

	protoPath := filepath.Join(tempDir, "proto/v1/api.proto")
	protoContent, err := os.ReadFile(protoPath)
	require.NoError(t, err)

	content := string(protoContent)

	assert.Contains(t, content, "message CreateRequest")
	assert.Contains(t, content, "message GetRequest")
	assert.Contains(t, content, "message CreateResponse")
	assert.Contains(t, content, "message GetResponse")
	assert.Contains(t, content, "message ErrorDetails")

	seenSchemas := make(map[string]int)
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "message ") {
			parts := strings.Fields(trimmed)
			if len(parts) >= 2 {
				schemaName := parts[1]
				seenSchemas[schemaName]++
			}
		}
	}

	for schema, count := range seenSchemas {
		assert.Equal(t, 1, count, "Schema %s should appear exactly once, but appeared %d times", schema, count)
	}
}

func TestProtoGenerationErrorCases(t *testing.T) {
	invalidFieldSpec := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /v1/test:
    post:
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/InvalidSchema'
      responses:
        '200':
          description: Success
components:
  schemas:
    InvalidSchema:
      type: object
      properties:
        1invalid:
          type: string
`

	specPath, stdout := setupTest(t, invalidFieldSpec)

	exitCode := duh.RunCmd(stdout, []string{"generate", specPath})

	require.Equal(t, 2, exitCode)
	output := stdout.String()
	require.NotEmpty(t, output)
}
