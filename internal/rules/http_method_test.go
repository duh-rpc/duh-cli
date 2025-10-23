package rules_test

import (
	"bytes"
	"testing"

	"github.com/duh-rpc/duhrpc-lint"
	"github.com/stretchr/testify/assert"
)

func TestHTTPMethodRule(t *testing.T) {
	for _, test := range []struct {
		name           string
		spec           string
		expectedExit   int
		expectedOutput string
	}{
		{
			name: "ValidPOSTMethod",
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
			name: "InvalidGETMethod",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/users.list:
    get:
      responses:
        200:
          description: Success
          content:
            application/json:
              schema:
                type: array`,
			expectedExit: 1,
			expectedOutput: `[http-method] GET /v1/users.list
  HTTP method GET is not allowed in DUH-RPC
  Use POST method for all DUH-RPC operations`,
		},
		{
			name: "InvalidPUTMethod",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/users.update:
    put:
      requestBody:
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
			expectedExit: 1,
			expectedOutput: `[http-method] PUT /v1/users.update
  HTTP method PUT is not allowed in DUH-RPC
  Use POST method for all DUH-RPC operations`,
		},
		{
			name: "InvalidDELETEMethod",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/users.delete:
    delete:
      responses:
        200:
          description: Success
          content:
            application/json:
              schema:
                type: object`,
			expectedExit: 1,
			expectedOutput: `[http-method] DELETE /v1/users.delete
  HTTP method DELETE is not allowed in DUH-RPC
  Use POST method for all DUH-RPC operations`,
		},
		{
			name: "InvalidPATCHMethod",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/users.patch:
    patch:
      requestBody:
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
			expectedExit: 1,
			expectedOutput: `[http-method] PATCH /v1/users.patch
  HTTP method PATCH is not allowed in DUH-RPC
  Use POST method for all DUH-RPC operations`,
		},
		{
			name: "MultipleNonPOSTMethods",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/users.manage:
    get:
      responses:
        200:
          description: Success
          content:
            application/json:
              schema:
                type: object
    put:
      requestBody:
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
    delete:
      responses:
        200:
          description: Success
          content:
            application/json:
              schema:
                type: object`,
			expectedExit: 1,
			expectedOutput: `[http-method] GET /v1/users.manage
  HTTP method GET is not allowed in DUH-RPC
  Use POST method for all DUH-RPC operations`,
		},
		{
			name: "MixedPOSTAndOtherMethods",
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
    get:
      responses:
        200:
          description: Success
          content:
            application/json:
              schema:
                type: object`,
			expectedExit: 1,
			expectedOutput: `[http-method] GET /v1/users.create
  HTTP method GET is not allowed in DUH-RPC
  Use POST method for all DUH-RPC operations`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			filePath := writeYAML(t, test.spec)

			var stdout bytes.Buffer
			exitCode := lint.RunCmd(&stdout, []string{filePath})

			assert.Equal(t, test.expectedExit, exitCode)
			assert.Contains(t, stdout.String(), test.expectedOutput)
		})
	}
}
