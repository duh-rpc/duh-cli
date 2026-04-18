package rules_test

import (
	"bytes"
	"testing"

	"github.com/duh-rpc/duh-cli"
	"github.com/stretchr/testify/assert"
)

func TestPaginationParametersRule(t *testing.T) {
	for _, test := range []struct {
		name           string
		spec           string
		expectedExit   int
		expectedOutput string
	}{
		{
			name: "ValidPaginationParameters",
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
        pagination:
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
        pagination:
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
			name: "MissingPaginationSubObject",
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
        filter:
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
        pagination:
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
			expectedOutput: "[PAGINATION_PARAMETERS]",
		},
		{
			name: "MissingFirst",
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
        pagination:
          type: object
          properties:
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
        pagination:
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
			expectedOutput: "[PAGINATION_PARAMETERS]",
		},
		{
			name: "FirstWrongType",
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
        pagination:
          type: object
          properties:
            first:
              type: string
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
        pagination:
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
			expectedOutput: "[PAGINATION_PARAMETERS]",
		},
		{
			name: "FirstMissingMinimum",
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
        pagination:
          type: object
          properties:
            first:
              type: integer
              format: int32
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
        pagination:
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
			expectedOutput: "[PAGINATION_PARAMETERS]",
		},
		{
			name: "FirstMissingMaximum",
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
        pagination:
          type: object
          properties:
            first:
              type: integer
              format: int32
              minimum: 1
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
        pagination:
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
			expectedOutput: "[PAGINATION_PARAMETERS]",
		},
		{
			name: "MissingAfter",
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
        pagination:
          type: object
          properties:
            first:
              type: integer
              format: int32
              minimum: 1
              maximum: 100
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
        pagination:
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
			expectedOutput: "[PAGINATION_PARAMETERS]",
		},
		{
			name: "AfterWrongType",
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
        pagination:
          type: object
          properties:
            first:
              type: integer
              format: int32
              minimum: 1
              maximum: 100
            after:
              type: integer
              format: int32
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
        pagination:
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
			expectedOutput: "[PAGINATION_PARAMETERS]",
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
