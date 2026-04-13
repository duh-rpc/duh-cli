package rules_test

import (
	"bytes"
	"testing"

	"github.com/duh-rpc/duh-cli"
	"github.com/stretchr/testify/assert"
)

func TestIdempotencyKeyDefinitionRule(t *testing.T) {
	for _, test := range []struct {
		name           string
		spec           string
		expectedExit   int
		expectedOutput string
	}{
		{
			name: "ValidIdempotencyKey",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /payments.create:
    post:
      description: Create a payment
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
        idempotencyKey:
          description: Idempotency key for request deduplication
          type: string
          maxLength: 128
    Error:
      type: object
      required: [message]
      properties:
        message:
          description: Error message
          type: string`,
			expectedExit:   0,
			expectedOutput: "compliant",
		},
		{
			name: "ValidNoIdempotencyKey",
			spec: `openapi: 3.0.0
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
      properties:
        name:
          description: Pet name
          type: string
    Error:
      type: object
      required: [message]
      properties:
        message:
          description: Error message
          type: string`,
			expectedExit:   0,
			expectedOutput: "compliant",
		},
		{
			name: "InvalidIdempotencyKeyInteger",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /payments.create:
    post:
      description: Create a payment
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
        idempotencyKey:
          description: Idempotency key
          type: integer
          format: int64
    Error:
      type: object
      required: [message]
      properties:
        message:
          description: Error message
          type: string`,
			expectedExit:   1,
			expectedOutput: "[IDEMPOTENCY_KEY_DEFINITION]",
		},
		{
			name: "InvalidIdempotencyKeyNoMaxLength",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /payments.create:
    post:
      description: Create a payment
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
        idempotencyKey:
          description: Idempotency key
          type: string
    Error:
      type: object
      required: [message]
      properties:
        message:
          description: Error message
          type: string`,
			expectedExit:   1,
			expectedOutput: "[IDEMPOTENCY_KEY_DEFINITION]",
		},
		{
			name: "InvalidIdempotencyKeyWrongMaxLength",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /payments.create:
    post:
      description: Create a payment
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
        idempotencyKey:
          description: Idempotency key
          type: string
          maxLength: 256
    Error:
      type: object
      required: [message]
      properties:
        message:
          description: Error message
          type: string`,
			expectedExit:   1,
			expectedOutput: "[IDEMPOTENCY_KEY_DEFINITION]",
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
