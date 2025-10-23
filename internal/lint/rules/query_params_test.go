package rules_test

import (
	"bytes"
	"testing"

	"github.com/duh-rpc/duhrpc"
	"github.com/stretchr/testify/assert"
)

func TestQueryParamsRule(t *testing.T) {
	for _, test := range []struct {
		name           string
		spec           string
		expectedExit   int
		expectedOutput string
	}{
		{
			name: "NoParameters",
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
			name: "HeaderParametersAllowed",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/users.create:
    post:
      parameters:
        - name: X-API-Key
          in: header
          required: true
          schema:
            type: string
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
			name: "CookieParametersAllowed",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/users.create:
    post:
      parameters:
        - name: session
          in: cookie
          required: true
          schema:
            type: string
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
			name: "QueryParameterNotAllowed",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/users.search:
    post:
      parameters:
        - name: query
          in: query
          schema:
            type: string
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
			expectedExit: 1,
			expectedOutput: `[query-parameters] POST /v1/users.search
  Query parameter 'query' is not allowed in DUH-RPC
  Move 'query' to request body`,
		},
		{
			name: "MultipleQueryParameters",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/users.search:
    post:
      parameters:
        - name: query
          in: query
          schema:
            type: string
        - name: limit
          in: query
          schema:
            type: integer
        - name: offset
          in: query
          schema:
            type: integer
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
			expectedExit: 1,
			expectedOutput: `[query-parameters] POST /v1/users.search
  Query parameter 'query' is not allowed in DUH-RPC
  Move 'query' to request body`,
		},
		{
			name: "MixedParameterTypes",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/users.search:
    post:
      parameters:
        - name: X-API-Key
          in: header
          required: true
          schema:
            type: string
        - name: query
          in: query
          schema:
            type: string
        - name: session
          in: cookie
          schema:
            type: string
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
			expectedExit: 1,
			expectedOutput: `[query-parameters] POST /v1/users.search
  Query parameter 'query' is not allowed in DUH-RPC
  Move 'query' to request body`,
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
