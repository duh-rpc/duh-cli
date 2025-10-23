package rules_test

import (
	"testing"

	"github.com/duh-rpc/duhrpc-lint/internal/rules"
	"github.com/pb33f/libopenapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatusCodeRule(t *testing.T) {
	for _, test := range []struct {
		name          string
		spec          string
		wantViolation bool
		wantCode      string
	}{
		{
			name: "AllowedStatusCode200",
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
			wantViolation: false,
		},
		{
			name: "AllowedStatusCode400",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/users.create:
    post:
      responses:
        400:
          description: Bad Request`,
			wantViolation: false,
		},
		{
			name: "AllowedStatusCode500",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/users.create:
    post:
      responses:
        500:
          description: Server Error`,
			wantViolation: false,
		},
		{
			name: "AllowedStatusCode452",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/users.create:
    post:
      responses:
        452:
          description: Custom Error`,
			wantViolation: false,
		},
		{
			name: "DisallowedStatusCode201",
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
			wantCode:      "201",
		},
		{
			name: "DisallowedStatusCode204",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/users.delete:
    post:
      responses:
        204:
          description: No Content`,
			wantViolation: true,
			wantCode:      "204",
		},
		{
			name: "DisallowedStatusCode503",
			spec: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /v1/users.create:
    post:
      responses:
        503:
          description: Service Unavailable`,
			wantViolation: true,
			wantCode:      "503",
		},
		{
			name: "MultipleAllowedCodes",
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
        400:
          description: Bad Request
        500:
          description: Server Error`,
			wantViolation: false,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			doc, err := libopenapi.NewDocument([]byte(test.spec))
			require.NoError(t, err)

			model, errs := doc.BuildV3Model()
			require.Empty(t, errs)

			rule := rules.NewStatusCodeRule()
			violations := rule.Validate(&model.Model)

			if test.wantViolation {
				require.NotEmpty(t, violations)
				assert.Contains(t, violations[0].Message, test.wantCode)
				assert.Equal(t, "status-code", violations[0].RuleName)
			} else {
				assert.Empty(t, violations)
			}
		})
	}
}
