package rules_test

import (
	"bytes"
	"testing"

	"github.com/duh-rpc/duh-cli"
	"github.com/stretchr/testify/assert"
)

func TestSchemaExampleValidationRule(t *testing.T) {
	for _, test := range []struct {
		name           string
		spec           string
		expectedExit   int
		expectedOutput string
	}{
		{
			name: "ValidStringExample",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /pets.create:
    post:
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
          type: string
          example: "Fido"
    Error:
      type: object
      required: [message]
      properties:
        message:
          type: string`,
			expectedExit:   0,
			expectedOutput: "compliant",
		},
		{
			name: "InvalidStringPropertyWithIntExample",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /pets.create:
    post:
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
          type: string
          example: 42
    Error:
      type: object
      required: [message]
      properties:
        message:
          type: string`,
			expectedExit:   1,
			expectedOutput: "[SCHEMA_EXAMPLE_VALIDATION]",
		},
		{
			name: "InvalidIntPropertyWithStringExample",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /pets.create:
    post:
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
        count:
          type: integer
          format: int32
          example: "not-a-number"
    Error:
      type: object
      required: [message]
      properties:
        message:
          type: string`,
			expectedExit:   1,
			expectedOutput: "[SCHEMA_EXAMPLE_VALIDATION]",
		},
		{
			name: "ValidEnumExample",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /pets.create:
    post:
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
        status:
          type: string
          enum: [active, inactive, pending]
          example: "active"
    Error:
      type: object
      required: [message]
      properties:
        message:
          type: string`,
			expectedExit:   0,
			expectedOutput: "compliant",
		},
		{
			name: "InvalidEnumExample",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /pets.create:
    post:
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
        status:
          type: string
          enum: [active, inactive, pending]
          example: "unknown"
    Error:
      type: object
      required: [message]
      properties:
        message:
          type: string`,
			expectedExit:   1,
			expectedOutput: "[SCHEMA_EXAMPLE_VALIDATION]",
		},
		{
			name: "NoExampleSkipped",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /pets.create:
    post:
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
          type: string
    Error:
      type: object
      required: [message]
      properties:
        message:
          type: string`,
			expectedExit:   0,
			expectedOutput: "compliant",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			filePath := writeYAML(t, test.spec)

			var stdout bytes.Buffer
			exitCode := duh.RunCmd(&stdout, []string{"lint", filePath})

			assert.Equal(t, test.expectedExit, exitCode)
			assert.Contains(t, stdout.String(), test.expectedOutput)
		})
	}
}
