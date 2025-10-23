package rules_test

import (
	"testing"

	"github.com/duh-rpc/duhrpc-lint/internal/rules"
	"github.com/pb33f/libopenapi"
	"github.com/stretchr/testify/require"
)

func TestPathFormatRule(t *testing.T) {
	for _, test := range []struct {
		name            string
		spec            string
		expectViolation bool
		violationCount  int
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
			expectViolation: false,
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
			expectViolation: false,
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
			expectViolation: false,
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
			expectViolation: true,
			violationCount:  1,
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
			expectViolation: true,
			violationCount:  1,
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
			expectViolation: true,
			violationCount:  1,
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
			expectViolation: true,
			violationCount:  1,
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
			expectViolation: true,
			violationCount:  1,
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
			expectViolation: true,
			violationCount:  1,
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
			expectViolation: true,
			violationCount:  1,
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
			expectViolation: true,
			violationCount:  1,
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
			expectViolation: true,
			violationCount:  1,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			doc, err := libopenapi.NewDocument([]byte(test.spec))
			require.NoError(t, err)

			model, errs := doc.BuildV3Model()
			require.Empty(t, errs)

			rule := rules.NewPathFormatRule()
			violations := rule.Validate(&model.Model)

			if test.expectViolation {
				require.NotEmpty(t, violations)
				require.Len(t, violations, test.violationCount)
				require.Equal(t, "path-format", violations[0].RuleName)
				require.NotEmpty(t, violations[0].Location)
				require.NotEmpty(t, violations[0].Message)
				require.NotEmpty(t, violations[0].Suggestion)
			} else {
				require.Empty(t, violations)
			}
		})
	}
}
