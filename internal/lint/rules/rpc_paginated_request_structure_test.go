package rules_test

import (
	"bytes"
	"testing"

	"github.com/duh-rpc/duh-cli"
	"github.com/stretchr/testify/assert"
)

func TestRPCPaginatedRequestStructureRule(t *testing.T) {
	for _, test := range []struct {
		name           string
		spec           string
		expectedExit   int
		expectedOutput string
	}{
		{
			name: "ValidPageSubObject",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /pets.list:
    post:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                page:
                  type: object
                  properties:
                    first:
                      type: integer
                      format: int32
                      minimum: 1
                      maximum: 100
                    after:
                      type: string
      responses:
        200:
          description: Success
          content:
            application/json:
              schema:
                type: object`,
			expectedExit:   0,
			expectedOutput: "compliant",
		},
		{
			name: "NonPaginatedWithFirst",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /pets.create:
    post:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                first:
                  type: integer
      responses:
        200:
          description: Success
          content:
            application/json:
              schema:
                type: object`,
			expectedExit:   0,
			expectedOutput: "compliant",
		},
		{
			name: "InvalidFirstAtRoot",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /pets.list:
    post:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                first:
                  type: integer
      responses:
        200:
          description: Success
          content:
            application/json:
              schema:
                type: object`,
			expectedExit:   1,
			expectedOutput: "[RPC_PAGINATED_REQUEST_STRUCTURE]",
		},
		{
			name: "InvalidAfterAtRoot",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /pets.list:
    post:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                after:
                  type: string
      responses:
        200:
          description: Success
          content:
            application/json:
              schema:
                type: object`,
			expectedExit:   1,
			expectedOutput: "[RPC_PAGINATED_REQUEST_STRUCTURE]",
		},
		{
			name: "InvalidBothAtRoot",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /pets.list:
    post:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                first:
                  type: integer
                after:
                  type: string
      responses:
        200:
          description: Success
          content:
            application/json:
              schema:
                type: object`,
			expectedExit: 1,
			expectedOutput: `Pagination parameter 'first' must be nested under 'page' sub-object`,
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

	// Additional check: both first and after produce separate violations
	t.Run("BothAtRootProducesTwoViolations", func(t *testing.T) {
		spec := `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /pets.list:
    post:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                first:
                  type: integer
                after:
                  type: string
      responses:
        200:
          description: Success
          content:
            application/json:
              schema:
                type: object`

		filePath := writeYAML(t, spec)

		var stdout bytes.Buffer
		exitCode := duh.RunCmd(&stdout, []string{"lint", filePath})

		assert.Equal(t, 1, exitCode)
		assert.Contains(t, stdout.String(), "Pagination parameter 'first' must be nested under 'page' sub-object")
		assert.Contains(t, stdout.String(), "Pagination parameter 'after' must be nested under 'page' sub-object")
	})
}
