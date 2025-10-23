package rules_test

import (
	"bytes"
	"testing"

	lint "github.com/duh-rpc/duhrpc-lint"
	"github.com/stretchr/testify/assert"
)

func TestContentTypeRule(t *testing.T) {
	for _, test := range []struct {
		name           string
		spec           string
		expectedExit   int
		expectedOutput string
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
			expectedExit:   0,
			expectedOutput: "✓ spec.yaml is DUH-RPC compliant",
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
			expectedExit:   0,
			expectedOutput: "✓ spec.yaml is DUH-RPC compliant",
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
			expectedExit:   0,
			expectedOutput: "✓ spec.yaml is DUH-RPC compliant",
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
			expectedExit: 1,
			expectedOutput: `[content-type] POST /v1/test.action
  Invalid request body content type: application/xml`,
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
			expectedExit: 1,
			expectedOutput: `[content-type] POST /v1/test.action response 200
  Invalid content type: text/html`,
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
			expectedExit: 1,
			expectedOutput: `[content-type] POST /v1/test.action response 200
  Invalid content type: text/plain`,
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
			expectedExit: 1,
			expectedOutput: `[content-type] POST /v1/test.action
  MIME parameters not allowed in request body content type`,
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
			expectedExit: 1,
			expectedOutput: `[content-type] POST /v1/test.action
  Request body must include application/json content type`,
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
			expectedExit: 1,
			expectedOutput: `[content-type]`,
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
			expectedExit:   0,
			expectedOutput: "✓ spec.yaml is DUH-RPC compliant",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			filePath := writeYAML(t, test.spec)

			var stdout bytes.Buffer
			exitCode := lint.RunCmd(&stdout, []string{filePath})

			assert.Equal(t, test.expectedExit, exitCode)
			assert.Contains(t, stdout.String(), test.expectedOutput)
		})
	}
}
