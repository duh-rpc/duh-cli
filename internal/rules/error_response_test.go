package rules_test

import (
	"testing"

	"github.com/duh-rpc/duhrpc-lint/internal/rules"
	"github.com/pb33f/libopenapi"
	"github.com/stretchr/testify/require"
)

func TestErrorResponseRule(t *testing.T) {
	for _, test := range []struct {
		name               string
		spec               string
		expectedViolations int
	}{
		{
			name: "ValidInlineErrorSchema",
			spec: `
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/test.action:
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
                    type: string
`,
			expectedViolations: 0,
		},
		{
			name: "ValidWithDetailsField",
			spec: `
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/test.action:
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
                    type: string
                  details:
                    type: object
`,
			expectedViolations: 0,
		},
		{
			name: "MissingRequiredFields",
			spec: `
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/test.action:
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
        400:
          description: Bad Request
          content:
            application/json:
              schema:
                type: object
                properties:
                  error:
                    type: string
`,
			expectedViolations: 1,
		},
		{
			name: "WrongFieldTypes",
			spec: `
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/test.action:
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
        500:
          description: Server Error
          content:
            application/json:
              schema:
                type: object
                required: [code, message]
                properties:
                  code:
                    type: string
                  message:
                    type: integer
`,
			expectedViolations: 1,
		},
		{
			name: "MissingObjectType",
			spec: `
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/test.action:
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
        400:
          description: Bad Request
          content:
            application/json:
              schema:
                required: [code, message]
                properties:
                  code:
                    type: integer
                  message:
                    type: string
`,
			expectedViolations: 1,
		},
		{
			name: "WrongDetailsType",
			spec: `
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/test.action:
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
                    type: string
                  details:
                    type: string
`,
			expectedViolations: 1,
		},
		{
			name: "MultipleErrorStatusCodes",
			spec: `
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/test.action:
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
                    type: string
        500:
          description: Server Error
          content:
            application/json:
              schema:
                type: object
                required: [code, message]
                properties:
                  code:
                    type: integer
                  message:
                    type: string
`,
			expectedViolations: 0,
		},
		{
			name: "AllErrorStatusCodes",
			spec: `
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/test.action:
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
                    type: string
        401:
          description: Unauthorized
          content:
            application/json:
              schema:
                type: object
                required: [code, message]
                properties:
                  code:
                    type: integer
                  message:
                    type: string
        403:
          description: Forbidden
          content:
            application/json:
              schema:
                type: object
                required: [code, message]
                properties:
                  code:
                    type: integer
                  message:
                    type: string
        404:
          description: Not Found
          content:
            application/json:
              schema:
                type: object
                required: [code, message]
                properties:
                  code:
                    type: integer
                  message:
                    type: string
        429:
          description: Too Many Requests
          content:
            application/json:
              schema:
                type: object
                required: [code, message]
                properties:
                  code:
                    type: integer
                  message:
                    type: string
        452:
          description: Custom Error
          content:
            application/json:
              schema:
                type: object
                required: [code, message]
                properties:
                  code:
                    type: integer
                  message:
                    type: string
        500:
          description: Server Error
          content:
            application/json:
              schema:
                type: object
                required: [code, message]
                properties:
                  code:
                    type: integer
                  message:
                    type: string
`,
			expectedViolations: 0,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			doc, err := libopenapi.NewDocument([]byte(test.spec))
			require.NoError(t, err)

			model, errs := doc.BuildV3Model()
			require.Empty(t, errs)

			rule := rules.NewErrorResponseRule()
			violations := rule.Validate(&model.Model)

			require.Len(t, violations, test.expectedViolations)

			if test.expectedViolations > 0 {
				for _, v := range violations {
					require.Equal(t, "error-response-schema", v.RuleName)
					require.NotEmpty(t, v.Message)
					require.NotEmpty(t, v.Suggestion)
					require.NotEmpty(t, v.Location)
				}
			}
		})
	}
}

func TestErrorResponseRuleWithRef(t *testing.T) {
	spec := `
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/test.action:
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
        400:
          description: Bad Request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
        500:
          description: Server Error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
components:
  schemas:
    Error:
      type: object
      required: [code, message]
      properties:
        code:
          type: integer
        message:
          type: string
        details:
          type: object
`

	doc, err := libopenapi.NewDocument([]byte(spec))
	require.NoError(t, err)

	model, errs := doc.BuildV3Model()
	require.Empty(t, errs)

	rule := rules.NewErrorResponseRule()
	violations := rule.Validate(&model.Model)

	require.Len(t, violations, 0)
}

func TestErrorResponseRuleWithAllOf(t *testing.T) {
	spec := `
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/test.action:
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
        400:
          description: Bad Request
          content:
            application/json:
              schema:
                allOf:
                  - $ref: '#/components/schemas/BaseError'
                  - type: object
                    properties:
                      details:
                        type: object
components:
  schemas:
    BaseError:
      type: object
      required: [code, message]
      properties:
        code:
          type: integer
        message:
          type: string
`

	doc, err := libopenapi.NewDocument([]byte(spec))
	require.NoError(t, err)

	model, errs := doc.BuildV3Model()
	require.Empty(t, errs)

	rule := rules.NewErrorResponseRule()
	violations := rule.Validate(&model.Model)

	require.Len(t, violations, 0)
}
