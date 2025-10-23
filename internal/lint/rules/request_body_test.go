package rules_test

import (
	"bytes"
	"testing"

	"github.com/duh-rpc/duhrpc"
	"github.com/stretchr/testify/assert"
)

func TestRequestBodyRule(t *testing.T) {
	for _, test := range []struct {
		name           string
		spec           string
		expectedExit   int
		expectedOutput string
	}{
		{
			name: "ValidRequestBody",
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
			name: "MissingRequestBody",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/users.create:
    post:
      responses:
        200:
          description: Success
          content:
            application/json:
              schema:
                type: object`,
			expectedExit: 1,
			expectedOutput: `[request-body-required] POST /v1/users.create
  Operation is missing a request body
  Add a required request body to this operation`,
		},
		{
			name: "RequestBodyNotRequired",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/users.create:
    post:
      requestBody:
        required: false
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
			expectedOutput: `[request-body-required] POST /v1/users.create
  Request body must be marked as required
  Set requestBody.required to true`,
		},
		{
			name: "RequestBodyRequiredOmitted",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/users.create:
    post:
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
			expectedOutput: `[request-body-required] POST /v1/users.create
  Request body must be marked as required
  Set requestBody.required to true`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			filePath := writeYAML(t, test.spec)

			var stdout bytes.Buffer
			exitCode := duhrpc.RunCmd(&stdout, []string{"lint", filePath})

			assert.Equal(t, test.expectedExit, exitCode)
			assert.Contains(t, stdout.String(), test.expectedOutput)
		})
	}
}
