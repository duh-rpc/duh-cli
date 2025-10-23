package rules_test

import (
	"testing"

	"github.com/duh-rpc/duhrpc-lint/internal/rules"
	"github.com/pb33f/libopenapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSuccessResponseRule(t *testing.T) {
	for _, test := range []struct {
		name          string
		spec          string
		wantViolation bool
		wantMessage   string
	}{
		{
			name: "Valid200ResponseWithObjectSchema",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/users.create:
    post:
      responses:
        200:
          description: Success
          content:
            application/json:
              schema:
                type: object`,
			wantViolation: false,
		},
		{
			name: "Valid200ResponseWithArraySchema",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/users.list:
    post:
      responses:
        200:
          description: Success
          content:
            application/json:
              schema:
                type: array`,
			wantViolation: false,
		},
		{
			name: "Valid200ResponseWithStringSchema",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/users.get:
    post:
      responses:
        200:
          description: Success
          content:
            text/plain:
              schema:
                type: string`,
			wantViolation: false,
		},
		{
			name: "Missing200Response",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/users.create:
    post:
      responses:
        201:
          description: Created`,
			wantViolation: true,
			wantMessage:   "missing a 200",
		},
		{
			name: "200ResponseWithoutContent",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/users.delete:
    post:
      responses:
        200:
          description: Success`,
			wantViolation: true,
			wantMessage:   "missing content",
		},
		{
			name: "200ResponseWithContentButNoSchema",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/users.create:
    post:
      responses:
        200:
          description: Success
          content:
            application/json:
              example:
                foo: bar`,
			wantViolation: true,
			wantMessage:   "missing a schema",
		},
		{
			name: "MultipleMediaTypesWithSchema",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/users.create:
    post:
      responses:
        200:
          description: Success
          content:
            application/json:
              schema:
                type: object
            application/xml:
              schema:
                type: object`,
			wantViolation: false,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			doc, err := libopenapi.NewDocument([]byte(test.spec))
			require.NoError(t, err)

			model, errs := doc.BuildV3Model()
			require.Empty(t, errs)

			rule := rules.NewSuccessResponseRule()
			violations := rule.Validate(&model.Model)

			if test.wantViolation {
				require.NotEmpty(t, violations)
				assert.Contains(t, violations[0].Message, test.wantMessage)
				assert.Equal(t, "success-response", violations[0].RuleName)
			} else {
				assert.Empty(t, violations)
			}
		})
	}
}
