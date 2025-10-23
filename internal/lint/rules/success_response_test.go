package rules_test

import (
	"bytes"
	"testing"

	lint "github.com/duh-rpc/duhrpc"
	"github.com/stretchr/testify/assert"
)

func TestSuccessResponseRule(t *testing.T) {
	for _, test := range []struct {
		name           string
		spec           string
		expectedExit   int
		expectedOutput string
	}{
		{
			name: "Valid200ResponseWithObjectSchema",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/users.create:
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
                type: object`,
			expectedExit:   0,
			expectedOutput: "✓ spec.yaml is DUH-RPC compliant",
		},
		{
			name: "Valid200ResponseWithArraySchema",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/users.list:
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
                type: array`,
			expectedExit:   0,
			expectedOutput: "✓ spec.yaml is DUH-RPC compliant",
		},
		{
			name: "Valid200ResponseWithStringSchema",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/users.get:
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
                type: string`,
			expectedExit:   0,
			expectedOutput: "✓ spec.yaml is DUH-RPC compliant",
		},
		{
			name: "Missing200Response",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/users.create:
    post:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
      responses:
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
                    type: string`,
			expectedExit: 1,
			expectedOutput: `[success-response] POST /v1/users.create
  Operation is missing a 200 (success) response`,
		},
		{
			name: "200ResponseWithoutContent",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/users.delete:
    post:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
      responses:
        200:
          description: Success`,
			expectedExit: 1,
			expectedOutput: `[success-response] POST /v1/users.delete
  200 response is missing content`,
		},
		{
			name: "200ResponseWithContentButNoSchema",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/users.create:
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
              example:
                foo: bar`,
			expectedExit: 1,
			expectedOutput: `[success-response] POST /v1/users.create
  200 response content is missing a schema`,
		},
		{
			name: "MultipleMediaTypesWithSchema",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/users.create:
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
            application/protobuf:
              schema:
                type: string
                format: binary`,
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
