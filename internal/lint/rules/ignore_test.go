package rules_test

import (
	"bytes"
	"testing"

	"github.com/duh-rpc/duh-cli"
	"github.com/stretchr/testify/assert"
)

func TestIgnoreAtOperationLevel(t *testing.T) {
	// Spec with x-duh-lint-ignore on an operation to suppress DESCRIPTION_REQUIRED
	spec := `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /pets.create:
    post:
      x-duh-lint-ignore:
        - DESCRIPTION_REQUIRED
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateRequest'
      responses:
        200:
          description: Success
          content:
            application/json:
              schema:
                type: object
        400:
          description: Bad request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
components:
  schemas:
    CreateRequest:
      type: object
      properties:
        name:
          description: The name
          type: string
    Error:
      type: object
      required: [message]
      properties:
        message:
          description: Error message
          type: string`

	filePath := writeYAML(t, spec)

	var stdout bytes.Buffer
	exitCode := duh.RunCmd(&stdout, []string{"lint", filePath})

	// DESCRIPTION_REQUIRED should not appear for that operation
	assert.Equal(t, 0, exitCode)
	assert.NotContains(t, stdout.String(), "DESCRIPTION_REQUIRED")
}

func TestIgnoreAtSchemaLevel(t *testing.T) {
	// Spec with x-duh-lint-ignore on a schema to suppress RPC_NO_NULLABLE
	spec := `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /pets.create:
    post:
      description: Create a pet
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateRequest'
      responses:
        200:
          description: Success
          content:
            application/json:
              schema:
                type: object
        400:
          description: Bad request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
components:
  schemas:
    CreateRequest:
      type: object
      x-duh-lint-ignore:
        - RPC_NO_NULLABLE
      properties:
        name:
          description: The name
          type: string
          nullable: true
    Error:
      type: object
      required: [message]
      properties:
        message:
          description: Error message
          type: string`

	filePath := writeYAML(t, spec)

	var stdout bytes.Buffer
	exitCode := duh.RunCmd(&stdout, []string{"lint", filePath})

	// RPC_NO_NULLABLE should not appear since schema has x-duh-lint-ignore
	assert.Equal(t, 0, exitCode)
	assert.NotContains(t, stdout.String(), "RPC_NO_NULLABLE")
}
