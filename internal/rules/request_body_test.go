package rules_test

import (
	"testing"

	"github.com/duh-rpc/duhrpc-lint/internal/rules"
	"github.com/pb33f/libopenapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestBodyRule(t *testing.T) {
	for _, test := range []struct {
		name          string
		spec          string
		wantViolation bool
		wantMessage   string
	}{
		{
			name: "ValidRequestBody",
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
          description: Success`,
			wantViolation: false,
		},
		{
			name: "MissingRequestBody",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/users.create:
    post:
      responses:
        200:
          description: Success`,
			wantViolation: true,
			wantMessage:   "missing a request body",
		},
		{
			name: "RequestBodyNotRequired",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/users.create:
    post:
      requestBody:
        required: false
        content:
          application/json:
            schema:
              type: object
      responses:
        200:
          description: Success`,
			wantViolation: true,
			wantMessage:   "must be marked as required",
		},
		{
			name: "RequestBodyRequiredOmitted",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/users.create:
    post:
      requestBody:
        content:
          application/json:
            schema:
              type: object
      responses:
        200:
          description: Success`,
			wantViolation: true,
			wantMessage:   "must be marked as required",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			doc, err := libopenapi.NewDocument([]byte(test.spec))
			require.NoError(t, err)

			model, errs := doc.BuildV3Model()
			require.Empty(t, errs)

			rule := rules.NewRequestBodyRule()
			violations := rule.Validate(&model.Model)

			if test.wantViolation {
				require.NotEmpty(t, violations)
				assert.Contains(t, violations[0].Message, test.wantMessage)
				assert.Equal(t, "request-body-required", violations[0].RuleName)
			} else {
				assert.Empty(t, violations)
			}
		})
	}
}
