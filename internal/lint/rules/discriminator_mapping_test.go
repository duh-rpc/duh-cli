package rules_test

import (
	"bytes"
	"testing"

	"github.com/duh-rpc/duh-cli"
	"github.com/stretchr/testify/assert"
)

func TestDiscriminatorMappingRule(t *testing.T) {
	for _, test := range []struct {
		name           string
		spec           string
		expectedExit   int
		wantContain    string
		wantNotContain string
	}{
		{
			name: "ValidDiscriminatorWithMapping",
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
			wantNotContain: "[DISCRIMINATOR_MAPPING]",
		},
		{
			name: "InvalidDiscriminatorNoMapping",
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
    Cat:
      type: object
      properties:
        type:
          type: string
          description: The variant type
    Dog:
      type: object
      properties:
        type:
          type: string
          description: The variant type
    Error:
      type: object
      required: [message]
      properties:
        message:
          type: string
          description: Error message`,
			expectedExit: 1,
			wantContain:  "[DISCRIMINATOR_MAPPING]",
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
