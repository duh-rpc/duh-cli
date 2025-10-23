package rules_test

import (
	"testing"

	"github.com/duh-rpc/duhrpc-lint/internal/rules"
	"github.com/pb33f/libopenapi"
	"github.com/stretchr/testify/require"
)

func TestHTTPMethodRule(t *testing.T) {
	for _, test := range []struct {
		name            string
		spec            string
		expectViolation bool
		violationCount  int
	}{
		{
			name: "ValidPOSTMethod",
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
			name: "InvalidGETMethod",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/users.list:
    get:
      responses:
        200:
          description: Success
          content:
            application/json:
              schema:
                type: array`,
			expectViolation: true,
			violationCount:  1,
		},
		{
			name: "InvalidPUTMethod",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/users.update:
    put:
      requestBody:
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
			name: "InvalidDELETEMethod",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/users.delete:
    delete:
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
			name: "InvalidPATCHMethod",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/users.patch:
    patch:
      requestBody:
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
			name: "MultipleNonPOSTMethods",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/users.manage:
    get:
      responses:
        200:
          description: Success
          content:
            application/json:
              schema:
                type: object
    put:
      requestBody:
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
    delete:
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
			name: "MixedPOSTAndOtherMethods",
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
    get:
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

			rule := rules.NewHTTPMethodRule()
			violations := rule.Validate(&model.Model)

			if test.expectViolation {
				require.NotEmpty(t, violations)
				require.Len(t, violations, test.violationCount)
				require.Equal(t, "http-method", violations[0].RuleName)
				require.NotEmpty(t, violations[0].Location)
				require.NotEmpty(t, violations[0].Message)
				require.NotEmpty(t, violations[0].Suggestion)
			} else {
				require.Empty(t, violations)
			}
		})
	}
}
