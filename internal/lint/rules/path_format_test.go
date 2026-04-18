package rules_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/duh-rpc/duh-cli"
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
			name: "ValidPath",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
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
			expectedExit:   0,
			expectedOutput: "✓ spec.yaml is DUH-RPC compliant",
		},
		{
			name: "ValidPathWithHyphensAndUnderscores",
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
			name: "ValidPathBeta",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /betas.test:
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
			name: "UppercaseResource",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /Users.create:
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
			expectedOutput: `[ERROR] [PATH_FORMAT] /Users.create
  Resource/Method must start with a lowercase letter`,
		},
		{
			name: "MissingDot",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /users:
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
			expectedOutput: `[ERROR] [PATH_FORMAT] /users
  Path must have format /{resource}.{method} with a dot separator
  Use format /{resource}.{method} (e.g., /users.create)`,
		},
		{
			name: "PathParametersInURL",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /users/{id}.get:
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
			expectedOutput: `[ERROR] [PATH_FORMAT] /users/{id}.get
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
  /user$accounts.create:
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
			expectedOutput: `[ERROR] [PATH_FORMAT] /user$accounts.create
  Resource/Method must contain only lowercase letters, numbers, hyphens, and underscores`,
		},
		{
			name: "PathParameterDefined",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /users.get:
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
			expectedOutput: `[ERROR] [PATH_FORMAT] /users.get
  Path parameter 'id' is not allowed in DUH-RPC
  Move path parameters to request body fields`,
		},
		{
			name: "MultiplePaths",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
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
			expectedOutput: `[ERROR] [PATH_FORMAT] /invalid-path
  Path must have format /{resource}.{method} with a dot separator
  Use format /{resource}.{method} (e.g., /users.create)`,
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
