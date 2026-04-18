package rules_test

import (
	"bytes"
	"testing"

	"github.com/duh-rpc/duh-cli"
	"github.com/stretchr/testify/assert"
)

func TestDiscriminatorRequiredRule(t *testing.T) {
	for _, test := range []struct {
		name           string
		spec           string
		expectedExit   int
		wantContain    string
		wantNotContain string
	}{
		{
			name: "ValidOneOfWithDiscriminator",
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
                $ref: '#/components/schemas/Error'
components:
  schemas:
    PetsCreateRequest:
      type: object
      properties:
        name:
          type: string
          description: The pet name
    PetsCreateResponse:
      oneOf:
        - $ref: '#/components/schemas/Cat'
        - $ref: '#/components/schemas/Dog'
      discriminator:
        propertyName: type
        mapping:
          cat: '#/components/schemas/Cat'
          dog: '#/components/schemas/Dog'
    Cat:
      type: object
      properties:
        type:
          type: string
          description: The variant type
        whiskers:
          type: integer
          format: int32
          description: Number of whiskers
    Dog:
      type: object
      properties:
        type:
          type: string
          description: The variant type
        breed:
          type: string
          description: The breed
    Error:
      type: object
      required: [message]
      properties:
        message:
          type: string
          description: Error message`,
			expectedExit:   1,
			wantNotContain: "[DISCRIMINATOR_REQUIRED]",
		},
		{
			name: "InvalidOneOfNoDiscriminator",
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
                $ref: '#/components/schemas/Error'
components:
  schemas:
    PetsCreateRequest:
      type: object
      properties:
        name:
          type: string
          description: The pet name
    PetsCreateResponse:
      oneOf:
        - type: object
          properties:
            name:
              type: string
        - type: object
          properties:
            id:
              type: string
    Error:
      type: object
      required: [message]
      properties:
        message:
          type: string
          description: Error message`,
			expectedExit: 1,
			wantContain:  "[DISCRIMINATOR_REQUIRED]",
		},
		{
			name: "ValidNoOneOf",
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
                $ref: '#/components/schemas/Error'
components:
  schemas:
    PetsCreateRequest:
      type: object
      properties:
        name:
          type: string
          description: The pet name
    PetsCreateResponse:
      type: object
      properties:
        id:
          type: string
          description: The ID
    Error:
      type: object
      required: [message]
      properties:
        message:
          type: string
          description: Error message`,
			expectedExit:   0,
			wantNotContain: "[DISCRIMINATOR_REQUIRED]",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			filePath := writeYAML(t, test.spec)

			var stdout bytes.Buffer
			exitCode := duh.RunCmd(&stdout, []string{"lint", filePath})

			assert.Equal(t, test.expectedExit, exitCode)
			if test.wantContain != "" {
				assert.Contains(t, stdout.String(), test.wantContain)
			}
			if test.wantNotContain != "" {
				assert.NotContains(t, stdout.String(), test.wantNotContain)
			}
		})
	}
}
