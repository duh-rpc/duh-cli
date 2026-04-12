package rules_test

import (
	"bytes"
	"testing"

	"github.com/duh-rpc/duh-cli"
	"github.com/stretchr/testify/assert"
)

func TestResponsePaginatedStructureRule(t *testing.T) {
	for _, test := range []struct {
		name           string
		spec           string
		expectedExit   int
		expectedOutput string
	}{
		{
			name: "ValidPaginatedResponse",
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
              $ref: '#/components/schemas/ListRequest'
      responses:
        200:
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ListResponse'
        400:
          description: Bad request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
components:
  schemas:
    ListRequest:
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
    ListResponse:
      type: object
      properties:
        items:
          type: array
          items:
            type: object
            properties:
              name:
                type: string
        page:
          type: object
          properties:
            endCursor:
              type: string
    Error:
      type: object
      required: [message]
      properties:
        message:
          type: string`,
			expectedExit:   0,
			expectedOutput: "compliant",
		},
		{
			name: "NonPaginatedEndpointSkipped",
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
              $ref: '#/components/schemas/CreateRequest'
      responses:
        200:
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/CreateResponse'
        400:
          description: Bad request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
components:
  schemas:
    CreateRequest:
      type: object
      properties:
        name:
          type: string
    CreateResponse:
      type: object
      properties:
        id:
          type: string
    Error:
      type: object
      required: [message]
      properties:
        message:
          type: string`,
			expectedExit:   0,
			expectedOutput: "compliant",
		},
		{
			name: "MissingItems",
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
              $ref: '#/components/schemas/ListRequest'
      responses:
        200:
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ListResponse'
        400:
          description: Bad request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
components:
  schemas:
    ListRequest:
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
    ListResponse:
      type: object
      properties:
        page:
          type: object
          properties:
            endCursor:
              type: string
    Error:
      type: object
      required: [message]
      properties:
        message:
          type: string`,
			expectedExit:   1,
			expectedOutput: "[RESPONSE_PAGINATED_STRUCTURE]",
		},
		{
			name: "ItemsNotArray",
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
              $ref: '#/components/schemas/ListRequest'
      responses:
        200:
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ListResponse'
        400:
          description: Bad request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
components:
  schemas:
    ListRequest:
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
    ListResponse:
      type: object
      properties:
        items:
          type: string
        page:
          type: object
          properties:
            endCursor:
              type: string
    Error:
      type: object
      required: [message]
      properties:
        message:
          type: string`,
			expectedExit:   1,
			expectedOutput: "[RESPONSE_PAGINATED_STRUCTURE]",
		},
		{
			name: "MissingPage",
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
              $ref: '#/components/schemas/ListRequest'
      responses:
        200:
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ListResponse'
        400:
          description: Bad request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
components:
  schemas:
    ListRequest:
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
    ListResponse:
      type: object
      properties:
        items:
          type: array
          items:
            type: object
            properties:
              name:
                type: string
    Error:
      type: object
      required: [message]
      properties:
        message:
          type: string`,
			expectedExit:   1,
			expectedOutput: "[RESPONSE_PAGINATED_STRUCTURE]",
		},
		{
			name: "PageMissingEndCursor",
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
              $ref: '#/components/schemas/ListRequest'
      responses:
        200:
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ListResponse'
        400:
          description: Bad request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
components:
  schemas:
    ListRequest:
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
    ListResponse:
      type: object
      properties:
        items:
          type: array
          items:
            type: object
            properties:
              name:
                type: string
        page:
          type: object
          properties:
            hasMore:
              type: boolean
    Error:
      type: object
      required: [message]
      properties:
        message:
          type: string`,
			expectedExit:   1,
			expectedOutput: "[RESPONSE_PAGINATED_STRUCTURE]",
		},
		{
			name: "SearchEndpointMissingBoth",
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
              $ref: '#/components/schemas/SearchRequest'
      responses:
        200:
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/SearchResponse'
        400:
          description: Bad request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
components:
  schemas:
    SearchRequest:
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
    SearchResponse:
      type: object
      properties:
        total:
          type: integer
          format: int32
    Error:
      type: object
      required: [message]
      properties:
        message:
          type: string`,
			expectedExit:   1,
			expectedOutput: "[RESPONSE_PAGINATED_STRUCTURE]",
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
