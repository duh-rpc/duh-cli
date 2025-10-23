package rules_test

import (
	"testing"

	"github.com/duh-rpc/duhrpc-lint/internal/rules"
	"github.com/pb33f/libopenapi"
	"github.com/stretchr/testify/require"
)

func TestContentTypeRule(t *testing.T) {
	for _, test := range []struct {
		name              string
		spec              string
		expectedViolations int
	}{
		{
			name: "ValidJSONContentType",
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
`,
			expectedViolations: 0,
		},
		{
			name: "ValidProtobufContentType",
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
          application/protobuf:
            schema:
              type: string
              format: binary
      responses:
        200:
          description: Success
          content:
            application/protobuf:
              schema:
                type: string
`,
			expectedViolations: 0,
		},
		{
			name: "ValidOctetStreamContentType",
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
          application/octet-stream:
            schema:
              type: string
              format: binary
      responses:
        200:
          description: Success
          content:
            application/octet-stream:
              schema:
                type: string
`,
			expectedViolations: 0,
		},
		{
			name: "InvalidXMLContentType",
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
          application/xml:
            schema:
              type: object
      responses:
        200:
          description: Success
          content:
            application/json:
              schema:
                type: object
`,
			expectedViolations: 1,
		},
		{
			name: "InvalidTextHTMLContentType",
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
            text/html:
              schema:
                type: string
`,
			expectedViolations: 1,
		},
		{
			name: "InvalidTextPlainContentType",
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
            text/plain:
              schema:
                type: string
`,
			expectedViolations: 1,
		},
		{
			name: "MIMEParametersNotAllowed",
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
          application/json; charset=utf-8:
            schema:
              type: object
      responses:
        200:
          description: Success
          content:
            application/json:
              schema:
                type: object
`,
			expectedViolations: 1,
		},
		{
			name: "MissingApplicationJSON",
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
          application/protobuf:
            schema:
              type: string
      responses:
        200:
          description: Success
          content:
            application/protobuf:
              schema:
                type: string
`,
			expectedViolations: 1,
		},
		{
			name: "MultipleInvalidContentTypes",
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
          application/xml:
            schema:
              type: object
      responses:
        200:
          description: Success
          content:
            text/html:
              schema:
                type: string
`,
			expectedViolations: 2,
		},
		{
			name: "CaseInsensitiveContentType",
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
          Application/JSON:
            schema:
              type: object
      responses:
        200:
          description: Success
          content:
            APPLICATION/JSON:
              schema:
                type: object
`,
			expectedViolations: 0,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			doc, err := libopenapi.NewDocument([]byte(test.spec))
			require.NoError(t, err)

			model, errs := doc.BuildV3Model()
			require.Empty(t, errs)

			rule := rules.NewContentTypeRule()
			violations := rule.Validate(&model.Model)

			require.Len(t, violations, test.expectedViolations)

			if test.expectedViolations > 0 {
				for _, v := range violations {
					require.Equal(t, "content-type", v.RuleName)
					require.NotEmpty(t, v.Message)
					require.NotEmpty(t, v.Suggestion)
					require.NotEmpty(t, v.Location)
				}
			}
		})
	}
}
