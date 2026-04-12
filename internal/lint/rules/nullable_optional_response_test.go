package rules_test

import (
	"bytes"
	"testing"

	"github.com/duh-rpc/duh-cli"
	"github.com/stretchr/testify/assert"
)

func TestNullableOptionalResponseRule(t *testing.T) {
	for _, test := range []struct {
		name           string
		spec           string
		expectedExit   int
		expectedOutput string
		absentOutput   string
	}{
		{
			name: "ValidNoNullableProperties",
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
                $ref: '#/components/schemas/CreateResponse'
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
        name:
          type: string
    CreateResponse:
      type: object
      properties:
        petId:
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
			name: "ValidNullableInRequired",
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
                $ref: '#/components/schemas/CreateResponse'
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
        name:
          type: string
    CreateResponse:
      type: object
      required: [petId, nickname]
      properties:
        petId:
          type: string
        nickname:
          type: string
          nullable: true
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
			expectedOutput: "[RPC_NO_NULLABLE]",
			absentOutput:   "[NULLABLE_OPTIONAL_RESPONSE]",
		},
		{
			name: "InvalidNullableOptional",
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
                $ref: '#/components/schemas/CreateResponse'
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
        name:
          type: string
    CreateResponse:
      type: object
      required: [petId]
      properties:
        petId:
          type: string
        nickname:
          type: string
          nullable: true
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
			expectedOutput: "[NULLABLE_OPTIONAL_RESPONSE]",
		},
		{
			name: "ValidRequestOnlyNullableOptional",
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
                $ref: '#/components/schemas/CreateResponse'
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
        name:
          type: string
        nickname:
          type: string
          nullable: true
    CreateResponse:
      type: object
      properties:
        petId:
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
			expectedOutput: "[RPC_NO_NULLABLE]",
			absentOutput:   "[NULLABLE_OPTIONAL_RESPONSE]",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			filePath := writeYAML(t, test.spec)

			var stdout bytes.Buffer
			exitCode := duh.RunCmd(&stdout, []string{"lint", filePath})

			assert.Equal(t, test.expectedExit, exitCode)
			assert.Contains(t, stdout.String(), test.expectedOutput)
			if test.absentOutput != "" {
				assert.NotContains(t, stdout.String(), test.absentOutput)
			}
		})
	}
}
