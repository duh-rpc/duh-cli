package rules_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/duh-rpc/duhrpc"
	"github.com/stretchr/testify/assert"
)

func writeYAML(t *testing.T, yaml string) string {
	t.Helper()
	dir := t.TempDir()
	filePath := filepath.Join(dir, "spec.yaml")
	err := os.WriteFile(filePath, []byte(yaml), 0644)
	if err != nil {
		t.Fatalf("Failed to write test YAML: %v", err)
	}
	return filePath
}

func TestPathFormatRule(t *testing.T) {
	for _, test := range []struct {
		name           string
		spec           string
		expectedExit   int
		expectedOutput string
	}{
		{
			name: "ValidPathV1",
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
			name: "ValidPathV0",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v0/beta.test:
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
			name: "ValidPathV10WithHyphensAndUnderscores",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v10/user-accounts.get-by-id:
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
			name: "MissingVersion",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /users.create:
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
			expectedExit: 1,
			expectedOutput: `[path-format] /users.create
  Path must start with version prefix (e.g., /v1/)
  Add a version prefix like /v1/`,
		},
		{
			name: "InvalidVersionFormat",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1.2/users.create:
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
			expectedExit: 1,
			expectedOutput: `[path-format] /v1.2/users.create
  Version must be /v{N}/ where N is a non-negative integer (e.g., /v1/, /v2/)
  Ensure path follows format /v{N}/subject.method with lowercase letters, numbers, hyphens, and underscores only`,
		},
		{
			name: "InvalidVersionBeta",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /vbeta/users.create:
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
			expectedExit: 1,
			expectedOutput: `[path-format] /vbeta/users.create
  Version must be /v{N}/ where N is a non-negative integer (e.g., /v1/, /v2/)
  Ensure path follows format /v{N}/subject.method with lowercase letters, numbers, hyphens, and underscores only`,
		},
		{
			name: "UppercaseSubject",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/Users.create:
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
			expectedExit: 1,
			expectedOutput: `[path-format] /v1/Users.create
  Subject must start with a lowercase letter
  Ensure path follows format /v{N}/subject.method with lowercase letters, numbers, hyphens, and underscores only`,
		},
		{
			name: "MissingDot",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/users:
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
			expectedExit: 1,
			expectedOutput: `[path-format] /v1/users
  Path must have format /v{N}/subject.method with a dot separator
  Use format /v1/subject.method (e.g., /v1/users.create)`,
		},
		{
			name: "PathParametersInURL",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/users/{id}.get:
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
			expectedExit: 1,
			expectedOutput: `[path-format] /v1/users/{id}.get
  Path contains path parameters, which are not allowed in DUH-RPC
  Remove path parameters and use request body fields instead`,
		},
		{
			name: "InvalidCharacters",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/user$accounts.create:
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
			expectedExit: 1,
			expectedOutput: `[path-format] /v1/user$accounts.create
  Subject must contain only lowercase letters, numbers, hyphens, and underscores
  Ensure path follows format /v{N}/subject.method with lowercase letters, numbers, hyphens, and underscores only`,
		},
		{
			name: "PathParameterDefined",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/users.get:
    parameters:
      - name: id
        in: path
        required: true
        schema:
          type: string
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
			expectedExit: 1,
			expectedOutput: `[path-format] /v1/users.get
  Path parameter 'id' is not allowed in DUH-RPC
  Move path parameters to request body fields`,
		},
		{
			name: "MultiplePaths",
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
  /invalid-path:
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
			expectedExit: 1,
			expectedOutput: `[path-format] /invalid-path
  Path must start with version prefix (e.g., /v1/)
  Add a version prefix like /v1/`,
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
