package duh_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	duh "github.com/duh-rpc/duh-cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RS-001: a teaching path declared on a path item that produces no generated
// operation (no post:) passes ValidateDidYouMean but is never emitted as a case
// arm, because resolveDidYouMean only runs for path items that survive
// extractOperations. The blueprint's contract is "for each declared path, the
// switch gains one arm" — this proves a declared path is silently dropped.
const didYouMeanOperationlessSpec = `openapi: 3.0.0
info:
  title: Widgets API
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /widgets.create:
    post:
      summary: Create a widget
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateRequest'
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/CreateResponse'
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorDetails'
  /widgets.get:
    x-duh-did-you-mean:
      - /widgets.fetch
components:
  schemas:
    CreateRequest:
      type: object
      properties:
        id:
          type: string
    CreateResponse:
      type: object
      properties:
        id:
          type: string
    ErrorDetails:
      type: object
      required:
        - message
      properties:
        message:
          type: string
`

// TestDidYouMeanOnOperationlessPathItemIsNotSilentlyDropped proves RS-001: a
// declared teaching path must not vanish with a clean exit 0. Either generate
// emits the teaching arm (the blueprint contract) or it errors naming the path —
// what it must not do is succeed while silently discarding the declared guess.
func TestDidYouMeanOnOperationlessPathItemIsNotSilentlyDropped(t *testing.T) {
	t.Skip("review-suite: RS-001 declared teaching path on operationless path item silently dropped; see review-suite.html")
	specPath, stdout := setupTest(t, didYouMeanOperationlessSpec)
	tempDir := filepath.Dir(specPath)

	exitCode := duh.RunCmd(context.Background(), stdout, []string{"generate", specPath})

	// A declared teaching path was silently dropped when generation succeeds but
	// the arm is absent. That is the defect: neither taught nor reported.
	if exitCode == 0 {
		content, err := os.ReadFile(filepath.Join(tempDir, "server.go"))
		require.NoError(t, err)
		assert.Contains(t, string(content), "/v1/widgets.fetch")
	}
}

// RS-002: an entry that passes the leading-slash check but contains characters
// that cannot appear in a Go string literal ('"', backslash, newline) slips past
// both validators and breaks template rendering with an opaque error. The
// blueprint acceptance criterion requires generate to exit non-zero NAMING the
// offending entry when an entry is malformed.
const didYouMeanUnsafeEntrySpec = `openapi: 3.0.0
info:
  title: Widgets API
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /widgets.create:
    x-duh-did-you-mean:
      - '/widgets".injected'
    post:
      summary: Create a widget
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateRequest'
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/CreateResponse'
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorDetails'
components:
  schemas:
    CreateRequest:
      type: object
      properties:
        id:
          type: string
    CreateResponse:
      type: object
      properties:
        id:
          type: string
    ErrorDetails:
      type: object
      required:
        - message
      properties:
        message:
          type: string
`

// TestDidYouMeanDuplicateWithinListNamesPathOnce proves RS-003: when the same
// teaching path is listed twice inside ONE path item's list, the error names the
// declaring path once — not the confusing "on X and X". Reuses
// duplicateWithinListSpec, which declares /diffs.get twice under /commits.diff.
func TestDidYouMeanDuplicateWithinListNamesPathOnce(t *testing.T) {
	specPath, stdout := setupTest(t, duplicateWithinListSpec)

	exitCode := duh.RunCmd(context.Background(), stdout, []string{"generate", specPath})

	require.NotEqual(t, 0, exitCode)
	output := stdout.String()
	assert.Contains(t, output, "declared more than once")
	assert.NotContains(t, output, `"/commits.diff" and "/commits.diff"`)
}

// TestDidYouMeanUnsafeEntryIsRejectedByName proves RS-002: an entry that would
// break generated Go source must be rejected by the validator with a named
// error, not slip through to an opaque template-render failure.
func TestDidYouMeanUnsafeEntryIsRejectedByName(t *testing.T) {
	t.Skip("review-suite: RS-002 entry breaking Go string literal slips past validation to opaque template error; see review-suite.html")
	specPath, stdout := setupTest(t, didYouMeanUnsafeEntrySpec)

	exitCode := duh.RunCmd(context.Background(), stdout, []string{"generate", specPath})

	require.NotEqual(t, 0, exitCode)
	output := stdout.String()
	assert.Contains(t, strings.ToLower(output), "malformed")
	assert.Contains(t, output, "widgets\".injected")
}
