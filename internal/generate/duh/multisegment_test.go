package duh_test

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A two-segment namespace path (/admin/jobs.types) is permitted by the PATH_FORMAT
// lint rule. Before the fix, ToCamelCase did not split on '/', so the '/' survived
// into the generated route const (RPCAdmin/jobsTypes), which go/format rejected as a
// division expression. The spec mirrors steve's real /admin/jobs.* contract.
const multiSegmentSpec = `openapi: 3.0.0
info:
  title: Admin Jobs API
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /admin/jobs.types:
    post:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/TypesRequest'
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/TypesResponse'
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorDetails'
components:
  schemas:
    TypesRequest:
      type: object
      properties:
        filter:
          type: string
    TypesResponse:
      type: object
      properties:
        names:
          type: array
          items:
            type: string
    ErrorDetails:
      type: object
      required:
        - message
      properties:
        message:
          type: string
`

// assertValidGo fails if the generated file at path is not syntactically valid Go.
// The multi-segment-path defect produced an invalid identifier that surfaces only as
// a render/parse error, so a syntax check is what catches it — a source-text Contains
// assertion would not.
func assertValidGo(t *testing.T, path string) {
	t.Helper()
	_, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.AllErrors)
	require.NoError(t, err)
}

// A two-segment path generates valid Go: the operation name and route const fold the
// namespace segment into a single identifier (AdminJobsTypes / RPCAdminJobsTypes) with
// no surviving '/', and both server.go and client.go parse cleanly.
func TestGenerateMultiSegmentPath(t *testing.T) {
	dir := writeProject(t, multiSegmentSpec)

	exitCode, out := generate(t, "--output-dir", "out")
	require.Equal(t, 0, exitCode, out)

	server := readFile(t, filepath.Join(dir, "out", "server.go"))
	assert.Contains(t, server, "AdminJobsTypes")
	assert.Contains(t, server, "RPCAdminJobsTypes")
	assert.NotContains(t, server, "Admin/jobs")

	client := readFile(t, filepath.Join(dir, "out", "client.go"))
	assert.Contains(t, client, "AdminJobsTypes")

	assertValidGo(t, filepath.Join(dir, "out", "server.go"))
	assertValidGo(t, filepath.Join(dir, "out", "client.go"))
}
