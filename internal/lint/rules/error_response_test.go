package rules_test

import (
	"bytes"
	"testing"

	lint "github.com/duh-rpc/duh-cli"
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
paths:
  /v1/test.action:
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
                required: [code, message]
                properties:
                  code:
                    type: integer
                  message:
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
paths:
  /v1/test.action:
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
                required: [code, message]
                properties:
                  code:
                    type: integer
                  message:
                    type: string
                  details:
                    type: object
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
  /v1/test.action:
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
			expectedOutput: `[error-response-schema] POST /v1/test.action response 400 (application/json)
  error schema must have required fields: code and message`,
		},
		{
			name: "WrongFieldTypes",
			spec: `
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/test.action:
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
			expectedOutput: `[error-response-schema] POST /v1/test.action response 500 (application/json)
  code field must be type integer`,
		},
		{
			name: "MissingObjectType",
			spec: `
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/test.action:
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
                required: [code, message]
                properties:
                  code:
                    type: integer
                  message:
                    type: string
`,
			expectedExit: 1,
			expectedOutput: `[error-response-schema] POST /v1/test.action response 400 (application/json)
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
  /v1/test.action:
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
                required: [code, message]
                properties:
                  code:
                    type: integer
                  message:
                    type: string
                  details:
                    type: string
`,
			expectedExit: 1,
			expectedOutput: `[error-response-schema] POST /v1/test.action response 400 (application/json)
  details field must be type object`,
		},
		{
			name: "MultipleErrorStatusCodes",
			spec: `
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/test.action:
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
                required: [code, message]
                properties:
                  code:
                    type: integer
                  message:
                    type: string
        500:
          description: Server Error
          content:
            application/json:
              schema:
                type: object
                required: [code, message]
                properties:
                  code:
                    type: integer
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
paths:
  /v1/test.action:
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
                required: [code, message]
                properties:
                  code:
                    type: integer
                  message:
                    type: string
        401:
          description: Unauthorized
          content:
            application/json:
              schema:
                type: object
                required: [code, message]
                properties:
                  code:
                    type: integer
                  message:
                    type: string
        403:
          description: Forbidden
          content:
            application/json:
              schema:
                type: object
                required: [code, message]
                properties:
                  code:
                    type: integer
                  message:
                    type: string
        404:
          description: Not Found
          content:
            application/json:
              schema:
                type: object
                required: [code, message]
                properties:
                  code:
                    type: integer
                  message:
                    type: string
        429:
          description: Too Many Requests
          content:
            application/json:
              schema:
                type: object
                required: [code, message]
                properties:
                  code:
                    type: integer
                  message:
                    type: string
        452:
          description: Custom Error
          content:
            application/json:
              schema:
                type: object
                required: [code, message]
                properties:
                  code:
                    type: integer
                  message:
                    type: string
        500:
          description: Server Error
          content:
            application/json:
              schema:
                type: object
                required: [code, message]
                properties:
                  code:
                    type: integer
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
			exitCode := lint.RunCmd(&stdout, []string{"lint", filePath})

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
paths:
  /v1/test.action:
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
                $ref: '#/components/schemas/Error'
        500:
          description: Server Error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
components:
  schemas:
    Error:
      type: object
      required: [code, message]
      properties:
        code:
          type: integer
        message:
          type: string
        details:
          type: object
`

	filePath := writeYAML(t, spec)

	var stdout bytes.Buffer
	exitCode := lint.RunCmd(&stdout, []string{"lint", filePath})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "✓ spec.yaml is DUH-RPC compliant")
}

func TestErrorResponseRuleWithAllOf(t *testing.T) {
	spec := `
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/test.action:
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
components:
  schemas:
    BaseError:
      type: object
      required: [code, message]
      properties:
        code:
          type: integer
        message:
          type: string
`

	filePath := writeYAML(t, spec)

	var stdout bytes.Buffer
	exitCode := lint.RunCmd(&stdout, []string{"lint", filePath})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "✓ spec.yaml is DUH-RPC compliant")
}
