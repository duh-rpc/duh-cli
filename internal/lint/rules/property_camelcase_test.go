package rules_test

import (
	"bytes"
	"testing"

	"github.com/duh-rpc/duh-cli"
	"github.com/stretchr/testify/assert"
)

func TestPropertyCamelCaseRule(t *testing.T) {
	for _, test := range []struct {
		name           string
		spec           string
		expectedExit   int
		expectedOutput string
	}{
		{
			name: "ValidCamelCase",
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
                $ref: '#/components/schemas/ErrorDetails'
components:
  schemas:
    CreateRequest:
      type: object
      properties:
        petId:
          type: string
        firstName:
          type: string
        name:
          type: string
    ErrorDetails:
      type: object
      required: [message]
      properties:
        message:
          type: string
        code:
          type: string
        type:
          type: string
        details:
          type: object
          additionalProperties:
            type: string`,
			expectedExit:   0,
			expectedOutput: "compliant",
		},
		{
			name: "InvalidSnakeCase",
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
                $ref: '#/components/schemas/ErrorDetails'
components:
  schemas:
    CreateRequest:
      type: object
      properties:
        first_name:
          type: string
    ErrorDetails:
      type: object
      required: [message]
      properties:
        message:
          type: string
        code:
          type: string
        type:
          type: string
        details:
          type: object
          additionalProperties:
            type: string`,
			expectedExit:   1,
			expectedOutput: "[PROPERTY_CAMELCASE]",
		},
		{
			name: "InvalidPascalCase",
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
                $ref: '#/components/schemas/ErrorDetails'
components:
  schemas:
    CreateRequest:
      type: object
      properties:
        FirstName:
          type: string
    ErrorDetails:
      type: object
      required: [message]
      properties:
        message:
          type: string
        code:
          type: string
        type:
          type: string
        details:
          type: object
          additionalProperties:
            type: string`,
			expectedExit:   1,
			expectedOutput: "[PROPERTY_CAMELCASE]",
		},
		{
			name: "InvalidScreamingSnakeCase",
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
                $ref: '#/components/schemas/ErrorDetails'
components:
  schemas:
    CreateRequest:
      type: object
      properties:
        FIRST_NAME:
          type: string
    ErrorDetails:
      type: object
      required: [message]
      properties:
        message:
          type: string
        code:
          type: string
        type:
          type: string
        details:
          type: object
          additionalProperties:
            type: string`,
			expectedExit:   1,
			expectedOutput: "[PROPERTY_CAMELCASE]",
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
