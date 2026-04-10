package rules_test

import (
	"bytes"
	"testing"

	"github.com/duh-rpc/duh-cli"
	"github.com/stretchr/testify/assert"
)

func TestRPCPaginationParametersRule(t *testing.T) {
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
			name: "NonPaginatedWithLimit",
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
                limit:
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
			name: "ValidListNoLimit",
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
                filter:
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
			name: "InvalidListWithLimit",
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
                limit:
                  type: integer
      responses:
        200:
          description: Success
          content:
            application/json:
              schema:
                type: object`,
			expectedExit:   1,
			expectedOutput: "[RPC_PAGINATION_PARAMETERS]",
		},
		{
			name: "InvalidSearchWithLimit",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /pets.search:
    post:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                limit:
                  type: integer
      responses:
        200:
          description: Success
          content:
            application/json:
              schema:
                type: object`,
			expectedExit:   1,
			expectedOutput: "[RPC_PAGINATION_PARAMETERS]",
		},
		{
			name: "InvalidQueryWithLimit",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /pets.query:
    post:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                limit:
                  type: integer
      responses:
        200:
          description: Success
          content:
            application/json:
              schema:
                type: object`,
			expectedExit:   1,
			expectedOutput: "[RPC_PAGINATION_PARAMETERS]",
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
