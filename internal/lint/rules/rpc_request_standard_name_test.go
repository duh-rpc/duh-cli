package rules_test

import (
	"bytes"
	"testing"

	"github.com/duh-rpc/duh-cli"
	"github.com/stretchr/testify/assert"
)

func TestRPCRequestStandardNameRule(t *testing.T) {
	for _, test := range []struct {
		name           string
		spec           string
		expectedExit   int
		expectedOutput string
	}{
		{
			name: "ValidMethodRequest",
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
          type: string`,
			expectedExit:   0,
			expectedOutput: "compliant",
		},
		{
			name: "ValidServiceMethodRequest",
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
              $ref: '#/components/schemas/PetsCreateRequest'
      responses:
        200:
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/PetsCreateResponse'
        400:
          description: Bad request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorDetails'
components:
  schemas:
    PetsCreateRequest:
      type: object
      properties:
        name:
          type: string
    PetsCreateResponse:
      type: object
      properties:
        petId:
          type: string
    ErrorDetails:
      type: object
      required: [message]
      properties:
        message:
          type: string`,
			expectedExit:   0,
			expectedOutput: "compliant",
		},
		{
			name: "ValidHyphenatedPath",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /user-accounts.get-by-id:
    post:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/GetByIdRequest'
      responses:
        200:
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/GetByIdResponse'
        400:
          description: Bad request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorDetails'
components:
  schemas:
    GetByIdRequest:
      type: object
      properties:
        userId:
          type: string
    GetByIdResponse:
      type: object
      properties:
        name:
          type: string
    ErrorDetails:
      type: object
      required: [message]
      properties:
        message:
          type: string`,
			expectedExit:   0,
			expectedOutput: "compliant",
		},
		{
			name: "ValidHyphenatedServiceMethodRequest",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /user-accounts.get-by-id:
    post:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/UserAccountsGetByIdRequest'
      responses:
        200:
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/UserAccountsGetByIdResponse'
        400:
          description: Bad request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorDetails'
components:
  schemas:
    UserAccountsGetByIdRequest:
      type: object
      properties:
        userId:
          type: string
    UserAccountsGetByIdResponse:
      type: object
      properties:
        name:
          type: string
    ErrorDetails:
      type: object
      required: [message]
      properties:
        message:
          type: string`,
			expectedExit:   0,
			expectedOutput: "compliant",
		},
		{
			name: "InvalidRequestName",
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
              $ref: '#/components/schemas/NewPetPayload'
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
    NewPetPayload:
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
          type: string`,
			expectedExit:   1,
			expectedOutput: "[RPC_REQUEST_STANDARD_NAME]",
		},
		{
			name: "InlineSchemaSkipped",
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
              type: object
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
