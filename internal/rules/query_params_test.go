package rules_test

import (
	"testing"

	"github.com/duh-rpc/duhrpc-lint/internal/rules"
	"github.com/pb33f/libopenapi"
	"github.com/stretchr/testify/require"
)

func TestQueryParamsRule(t *testing.T) {
	for _, test := range []struct {
		name            string
		spec            string
		expectViolation bool
		violationCount  int
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
			expectViolation: false,
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
			expectViolation: false,
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
			expectViolation: false,
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
			expectViolation: true,
			violationCount:  1,
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
			expectViolation: true,
			violationCount:  3,
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
			expectViolation: true,
			violationCount:  1,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			doc, err := libopenapi.NewDocument([]byte(test.spec))
			require.NoError(t, err)

			model, errs := doc.BuildV3Model()
			require.Empty(t, errs)

			rule := rules.NewQueryParamsRule()
			violations := rule.Validate(&model.Model)

			if test.expectViolation {
				require.NotEmpty(t, violations)
				require.Len(t, violations, test.violationCount)
				require.Equal(t, "query-parameters", violations[0].RuleName)
				require.NotEmpty(t, violations[0].Location)
				require.NotEmpty(t, violations[0].Message)
				require.NotEmpty(t, violations[0].Suggestion)
			} else {
				require.Empty(t, violations)
			}
		})
	}
}
