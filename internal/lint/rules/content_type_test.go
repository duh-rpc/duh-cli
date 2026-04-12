package rules_test

import (
	"bytes"
	"testing"

	lint "github.com/duh-rpc/duh-cli"
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
servers:
  - url: https://api.example.com/v1
paths:
  /tests.action:
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
servers:
  - url: https://api.example.com/v1
paths:
  /tests.action:
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
			name: "InvalidOctetStreamContentType",
			spec: `
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /tests.action:
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
			expectedExit: 1,
			expectedOutput: `[ERROR] [CONTENT_TYPE]`,
		},
		{
			name: "InvalidXMLContentType",
			spec: `
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /tests.action:
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
			expectedOutput: `[ERROR] [CONTENT_TYPE] POST /tests.action
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
  /tests.action:
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
			expectedOutput: `[ERROR] [CONTENT_TYPE] POST /tests.action response 200
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
  /tests.action:
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
			expectedOutput: `[ERROR] [CONTENT_TYPE] POST /tests.action response 200
  Invalid content type: text/plain`,
		},
		{
			name: "InvalidMultipartFormData",
			spec: `
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /tests.action:
    post:
      requestBody:
        required: true
        content:
          multipart/form-data:
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
			expectedOutput: `[ERROR] [CONTENT_TYPE] POST /tests.action
  Multipart and form-encoded content types are not allowed
  Use application/json or application/protobuf`,
		},
		{
			name: "InvalidFormURLEncoded",
			spec: `
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /tests.action:
    post:
      requestBody:
        required: true
        content:
          application/x-www-form-urlencoded:
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
			expectedOutput: `[ERROR] [CONTENT_TYPE] POST /tests.action
  Multipart and form-encoded content types are not allowed
  Use application/json or application/protobuf`,
		},
		{
			name: "MIMEParametersNotAllowed",
			spec: `
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /tests.action:
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
			expectedOutput: `[ERROR] [CONTENT_TYPE] POST /tests.action
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
  /tests.action:
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
			expectedOutput: `[ERROR] [CONTENT_TYPE] POST /tests.action
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
  /tests.action:
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
			expectedExit:   1,
			expectedOutput: `[ERROR] [CONTENT_TYPE]`,
		},
		{
			name: "CaseInsensitiveContentType",
			spec: `
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /tests.action:
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
			exitCode := lint.RunCmd(&stdout, []string{"lint", filePath})

			assert.Equal(t, test.expectedExit, exitCode)
			assert.Contains(t, stdout.String(), test.expectedOutput)
		})
	}
}
