package rules_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/duh-rpc/duh-cli"
	"github.com/stretchr/testify/assert"
)

// specWithSchemas wraps the given component-schema block in an otherwise compliant
// DUH spec. The schemas block must define CreateRequest and the enum under test;
// ErrorDetails is appended so the error-response rule is satisfied.
func specWithSchemas(schemas string) string {
	return `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /pets.create:
    post:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateRequest'
      responses:
        200:
          description: Success
          content:
            application/json:
              schema:
                type: object
        400:
          description: Bad request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorDetails'
components:
  schemas:
` + schemas + `
    ErrorDetails:
      type: object
      required: [message]
      properties:
        message:
          type: string`
}

// Acceptance #1: a type: string enum without a sentinel is the ticket's repro and is
// compliant — there is no proto enum and no number 0 to protect.
func TestEnumUnspecifiedStringEnumNoSentinelCompliant(t *testing.T) {
	spec := specWithSchemas(`    CreateRequest:
      type: object
      properties:
        offset:
          $ref: '#/components/schemas/OffsetMode'
    OffsetMode:
      type: string
      enum:
        - earliest
        - latest
        - at_offset`)

	filePath := writeYAML(t, spec)
	var stdout bytes.Buffer
	exitCode := duh.RunCmd(context.Background(), &stdout, []string{"lint", filePath})

	assert.Equal(t, 0, exitCode)
	assert.NotContains(t, stdout.String(), "[ENUM_UNSPECIFIED_VARIANT]")
	assert.NotContains(t, stdout.String(), "[ENUM_STRING_SENTINEL_NAMES]")
}

// Acceptance #2: a type: string enum with sentinel-style names raises no
// ENUM_UNSPECIFIED_VARIANT error (string enums are open proto string fields) but does
// raise the ENUM_STRING_SENTINEL_NAMES advisory warning.
func TestEnumUnspecifiedStringEnumWithSentinelWarns(t *testing.T) {
	spec := specWithSchemas(`    CreateRequest:
      type: object
      properties:
        status:
          $ref: '#/components/schemas/Status'
    Status:
      type: string
      enum:
        - STATUS_UNSPECIFIED
        - STATUS_ACTIVE`)

	filePath := writeYAML(t, spec)
	var stdout bytes.Buffer
	exitCode := duh.RunCmd(context.Background(), &stdout, []string{"lint", filePath})

	assert.Equal(t, 0, exitCode)
	assert.NotContains(t, stdout.String(), "[ENUM_UNSPECIFIED_VARIANT]")
	assert.Contains(t, stdout.String(), "[ENUM_STRING_SENTINEL_NAMES]")
}

// Acceptance #3: a bare numeric integer enum is a legitimate proto enum whose first
// value takes number 0; it must not be flagged.
func TestEnumUnspecifiedBareIntegerEnumCompliant(t *testing.T) {
	spec := specWithSchemas(`    CreateRequest:
      type: object
      properties:
        code:
          $ref: '#/components/schemas/Code'
    Code:
      type: integer
      format: int32
      enum:
        - 200
        - 404
        - 500`)

	filePath := writeYAML(t, spec)
	var stdout bytes.Buffer
	exitCode := duh.RunCmd(context.Background(), &stdout, []string{"lint", filePath})

	assert.Equal(t, 0, exitCode)
	assert.NotContains(t, stdout.String(), "[ENUM_UNSPECIFIED_VARIANT]")
}

// Acceptance #4: an integer named enum with the sentinel declared first is the
// wire-safe closed enum and is compliant.
func TestEnumUnspecifiedIntegerSentinelFirstCompliant(t *testing.T) {
	spec := specWithSchemas(`    CreateRequest:
      type: object
      properties:
        code:
          $ref: '#/components/schemas/Code'
    Code:
      type: integer
      format: int32
      enum:
        - CODE_UNSPECIFIED
        - CODE_OK`)

	filePath := writeYAML(t, spec)
	var stdout bytes.Buffer
	exitCode := duh.RunCmd(context.Background(), &stdout, []string{"lint", filePath})

	assert.Equal(t, 0, exitCode)
	assert.NotContains(t, stdout.String(), "[ENUM_UNSPECIFIED_VARIANT]")
}

// Acceptance #5: an integer named enum whose sentinel is not first is flagged — the
// sentinel must own number 0.
func TestEnumUnspecifiedIntegerSentinelNotFirstFlagged(t *testing.T) {
	spec := specWithSchemas(`    CreateRequest:
      type: object
      properties:
        code:
          $ref: '#/components/schemas/Code'
    Code:
      type: integer
      format: int32
      enum:
        - CODE_OK
        - CODE_UNSPECIFIED`)

	filePath := writeYAML(t, spec)
	var stdout bytes.Buffer
	exitCode := duh.RunCmd(context.Background(), &stdout, []string{"lint", filePath})

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, stdout.String(), "[ENUM_UNSPECIFIED_VARIANT]")
	assert.Contains(t, stdout.String(), "CODE_OK")
}

// Acceptance #6: an integer named enum with no sentinel is compliant — the rule
// mirrors fieldmap and does not nudge toward adding one.
func TestEnumUnspecifiedIntegerNoSentinelCompliant(t *testing.T) {
	spec := specWithSchemas(`    CreateRequest:
      type: object
      properties:
        code:
          $ref: '#/components/schemas/Code'
    Code:
      type: integer
      format: int32
      enum:
        - CODE_OK
        - CODE_ERR`)

	filePath := writeYAML(t, spec)
	var stdout bytes.Buffer
	exitCode := duh.RunCmd(context.Background(), &stdout, []string{"lint", filePath})

	assert.Equal(t, 0, exitCode)
	assert.NotContains(t, stdout.String(), "[ENUM_UNSPECIFIED_VARIANT]")
}

// Acceptance #7: the same mis-ordered sentinel on an inline integer property (not a
// component schema) is also flagged.
func TestEnumUnspecifiedInlineIntegerSentinelNotFirstFlagged(t *testing.T) {
	spec := specWithSchemas(`    CreateRequest:
      type: object
      properties:
        code:
          type: integer
          format: int32
          enum:
            - CODE_OK
            - CODE_UNSPECIFIED`)

	filePath := writeYAML(t, spec)
	var stdout bytes.Buffer
	exitCode := duh.RunCmd(context.Background(), &stdout, []string{"lint", filePath})

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, stdout.String(), "[ENUM_UNSPECIFIED_VARIANT]")
}

// Acceptance #8: a free-form type: string with no enum is never flagged by either rule.
func TestEnumUnspecifiedFreeFormStringNotFlagged(t *testing.T) {
	spec := specWithSchemas(`    CreateRequest:
      type: object
      properties:
        name:
          type: string`)

	filePath := writeYAML(t, spec)
	var stdout bytes.Buffer
	exitCode := duh.RunCmd(context.Background(), &stdout, []string{"lint", filePath})

	assert.Equal(t, 0, exitCode)
	assert.NotContains(t, stdout.String(), "[ENUM_UNSPECIFIED_VARIANT]")
	assert.NotContains(t, stdout.String(), "[ENUM_STRING_SENTINEL_NAMES]")
}

// Acceptance #9: the ENUM_STRING_SENTINEL_NAMES warning alone yields exit code 0 —
// warnings do not fail lint.
func TestEnumStringSentinelWarningDoesNotFailLint(t *testing.T) {
	spec := specWithSchemas(`    CreateRequest:
      type: object
      properties:
        status:
          type: string
          enum:
            - STATUS_UNSPECIFIED
            - STATUS_ACTIVE`)

	filePath := writeYAML(t, spec)
	var stdout bytes.Buffer
	exitCode := duh.RunCmd(context.Background(), &stdout, []string{"lint", filePath})

	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "[ENUM_STRING_SENTINEL_NAMES]")
	assert.Contains(t, stdout.String(), "WARNING")
	assert.Contains(t, stdout.String(), "compliant")
}
