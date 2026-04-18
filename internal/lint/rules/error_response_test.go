package rules_test

import (
	"bytes"
	"testing"

	duh "github.com/duh-rpc/duh-cli"
	"github.com/stretchr/testify/assert"
)

func TestErrorResponseRule(t *testing.T) {
	for _, test := range []struct {
		name           string
		spec           string
		expectedExit   int
		expectedOutput string
	}{
		{
			name: "ValidInlineErrorSchema",
			spec: `
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /tests.action:
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
          description: Bad Request
          content:
            application/json:
              schema:
                type: object
                required: [message]
                properties:
                  message:
                    type: string
                  code:
                    type: string
                  details:
                    type: object
                    additionalProperties:
                      type: string
`,
			expectedExit:   0,
			expectedOutput: "✓ spec.yaml is DUH-RPC compliant",
		},
		{
			name: "ValidWithDetailsField",
			spec: `
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /tests.action:
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
          description: Bad Request
          content:
            application/json:
              schema:
                type: object
                required: [message]
                properties:
                  message:
                    type: string
                  details:
                    type: object
                    additionalProperties:
                      type: string
`,
			expectedExit:   0,
			expectedOutput: "✓ spec.yaml is DUH-RPC compliant",
		},
		{
			name: "MissingRequiredFields",
			spec: `
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /tests.action:
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
          description: Bad Request
          content:
            application/json:
              schema:
                type: object
                properties:
                  error:
                    type: string
`,
			expectedExit: 1,
			expectedOutput: `[ERROR] [ERROR_SCHEMA] POST /tests.action response 400 (application/json)
  error schema must have required field: message`,
		},
		{
			name: "WrongFieldTypes",
			spec: `
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /tests.action:
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
        500:
          description: Server Error
          content:
            application/json:
              schema:
                type: object
                required: [code, message]
                properties:
                  code:
                    type: string
                  message:
                    type: integer
`,
			expectedExit: 1,
			expectedOutput: `[ERROR] [ERROR_SCHEMA] POST /tests.action response 500 (application/json)
  message field must be type string`,
		},
		{
			name: "MissingObjectType",
			spec: `
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /tests.action:
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
          description: Bad Request
          content:
            application/json:
              schema:
                required: [message]
                properties:
                  message:
                    type: string
`,
			expectedExit: 1,
			expectedOutput: `[ERROR] [ERROR_SCHEMA] POST /tests.action response 400 (application/json)
  error schema must have explicit type 'object'`,
		},
		{
			name: "WrongDetailsType",
			spec: `
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /tests.action:
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
          description: Bad Request
          content:
            application/json:
              schema:
                type: object
                required: [message]
                properties:
                  message:
                    type: string
                  details:
                    type: string
`,
			expectedExit: 1,
			expectedOutput: `[ERROR] [ERROR_SCHEMA] POST /tests.action response 400 (application/json)
  details field must be type object`,
		},
		{
			name: "CodeAsInteger",
			spec: `
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /tests.action:
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
          description: Bad Request
          content:
            application/json:
              schema:
                type: object
                required: [message]
                properties:
                  message:
                    type: string
                  code:
                    type: integer
`,
			expectedExit: 1,
			expectedOutput: `[ERROR] [ERROR_SCHEMA] POST /tests.action response 400 (application/json)
  code field must be type string`,
		},
		{
			name: "TypeFieldAsInteger",
			spec: `
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /tests.action:
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
          description: Bad Request
          content:
            application/json:
              schema:
                type: object
                required: [message]
                properties:
                  message:
                    type: string
                  type:
                    type: integer
`,
			expectedExit: 1,
			expectedOutput: `[ERROR] [ERROR_SCHEMA] POST /tests.action response 400 (application/json)
  type field must be type string`,
		},
		{
			name: "CodeAbsentIsValid",
			spec: `
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /tests.action:
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
          description: Bad Request
          content:
            application/json:
              schema:
                type: object
                required: [message]
                properties:
                  message:
                    type: string
`,
			expectedExit:   0,
			expectedOutput: "✓ spec.yaml is DUH-RPC compliant",
		},
		{
			name: "DetailsWithoutAdditionalProperties",
			spec: `
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /tests.action:
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
          description: Bad Request
          content:
            application/json:
              schema:
                type: object
                required: [message]
                properties:
                  message:
                    type: string
                  details:
                    type: object
`,
			expectedExit: 1,
			expectedOutput: `[ERROR] [ERROR_SCHEMA] POST /tests.action response 400 (application/json)
  details field must have additionalProperties with type string`,
		},
		{
			name: "DetailsWithAdditionalPropertiesTrue",
			spec: `
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /tests.action:
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
          description: Bad Request
          content:
            application/json:
              schema:
                type: object
                required: [message]
                properties:
                  message:
                    type: string
                  details:
                    type: object
                    additionalProperties: true
`,
			expectedExit: 1,
			expectedOutput: `[ERROR] [ERROR_SCHEMA] POST /tests.action response 400 (application/json)
  details field must have additionalProperties with type string`,
		},
		{
			name: "DetailsWithValidAdditionalProperties",
			spec: `
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /tests.action:
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
          description: Bad Request
          content:
            application/json:
              schema:
                type: object
                required: [message]
                properties:
                  message:
                    type: string
                  details:
                    type: object
                    additionalProperties:
                      type: string
`,
			expectedExit:   0,
			expectedOutput: "✓ spec.yaml is DUH-RPC compliant",
		},
		{
			name: "MultipleErrorStatusCodes",
			spec: `
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /tests.action:
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
          description: Bad Request
          content:
            application/json:
              schema:
                type: object
                required: [message]
                properties:
                  message:
                    type: string
        500:
          description: Server Error
          content:
            application/json:
              schema:
                type: object
                required: [message]
                properties:
                  message:
                    type: string
`,
			expectedExit:   0,
			expectedOutput: "✓ spec.yaml is DUH-RPC compliant",
		},
		{
			name: "AllErrorStatusCodes",
			spec: `
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /tests.action:
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
          description: Bad Request
          content:
            application/json:
              schema:
                type: object
                required: [message]
                properties:
                  message:
                    type: string
        401:
          description: Unauthorized
          content:
            application/json:
              schema:
                type: object
                required: [message]
                properties:
                  message:
                    type: string
        403:
          description: Forbidden
          content:
            application/json:
              schema:
                type: object
                required: [message]
                properties:
                  message:
                    type: string
        404:
          description: Not Found
          content:
            application/json:
              schema:
                type: object
                required: [message]
                properties:
                  message:
                    type: string
        409:
          description: Conflict
          content:
            application/json:
              schema:
                type: object
                required: [message]
                properties:
                  message:
                    type: string
        429:
          description: Too Many Requests
          content:
            application/json:
              schema:
                type: object
                required: [message]
                properties:
                  message:
                    type: string
        500:
          description: Server Error
          content:
            application/json:
              schema:
                type: object
                required: [message]
                properties:
                  message:
                    type: string
`,
			expectedExit:   0,
			expectedOutput: "✓ spec.yaml is DUH-RPC compliant",
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

func TestErrorResponseRuleWithRef(t *testing.T) {
	spec := `
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /tests.action:
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
          description: Bad Request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorDetails'
        500:
          description: Server Error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorDetails'
components:
  schemas:
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
            type: string
`

	filePath := writeYAML(t, spec)

	var stdout bytes.Buffer
	exitCode := duh.RunCmd(&stdout, []string{"lint", filePath})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "✓ spec.yaml is DUH-RPC compliant")
}

func TestErrorResponseRuleWithAllOf(t *testing.T) {
	spec := `
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /tests.action:
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
          description: Bad Request
          content:
            application/json:
              schema:
                allOf:
                  - $ref: '#/components/schemas/BaseError'
                  - type: object
                    properties:
                      details:
                        type: object
                        additionalProperties:
                          type: string
components:
  schemas:
    BaseError:
      type: object
      required: [message]
      properties:
        message:
          type: string
        code:
          type: string
`

	filePath := writeYAML(t, spec)

	var stdout bytes.Buffer
	exitCode := duh.RunCmd(&stdout, []string{"lint", filePath})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "✓ spec.yaml is DUH-RPC compliant")
}
