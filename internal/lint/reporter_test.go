package lint_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/duh-rpc/duhrpc-lint/internal/lint"
	"github.com/stretchr/testify/assert"
)

func TestPrintValidResult(t *testing.T) {
	var buf bytes.Buffer
	result := lint.ValidationResult{
		FilePath: "/path/to/test.yaml",
	}

	lint.Print(&buf, result)

	output := buf.String()
	assert.Contains(t, output, "✓")
	assert.Contains(t, output, "test.yaml")
	assert.Contains(t, output, "DUH-RPC compliant")
}

func TestPrintWithViolations(t *testing.T) {
	var buf bytes.Buffer
	result := lint.ValidationResult{
		FilePath: "/path/to/test.yaml",
		Violations: []lint.Violation{
			{
				Suggestion: "Use POST instead",
				RuleName:   "http-method",
				Location:   "GET /v1/users.list",
				Message:    "Only POST is allowed",
			},
		},
	}

	lint.Print(&buf, result)

	output := buf.String()
	assert.Contains(t, output, "Validating test.yaml")
	assert.Contains(t, output, "ERRORS FOUND:")
	assert.Contains(t, output, "[http-method]")
	assert.Contains(t, output, "GET /v1/users.list")
	assert.Contains(t, output, "Only POST is allowed")
	assert.Contains(t, output, "Use POST instead")
	assert.Contains(t, output, "Summary: 1 violations found in test.yaml")
}

func TestPrintMultipleViolations(t *testing.T) {
	var buf bytes.Buffer
	result := lint.ValidationResult{
		FilePath: "/path/to/test.yaml",
		Violations: []lint.Violation{
			{
				RuleName: "rule-1",
				Location: "/v1/test1",
				Message:  "Violation 1",
			},
			{
				RuleName: "rule-2",
				Location: "/v1/test2",
				Message:  "Violation 2",
			},
			{
				RuleName: "rule-3",
				Location: "/v1/test3",
				Message:  "Violation 3",
			},
		},
	}

	lint.Print(&buf, result)

	output := buf.String()
	assert.Contains(t, output, "ERRORS FOUND:")
	assert.Contains(t, output, "[rule-1]")
	assert.Contains(t, output, "[rule-2]")
	assert.Contains(t, output, "[rule-3]")
	assert.Contains(t, output, "Summary: 3 violations found in test.yaml")

	// Verify violations appear in order
	rule1Pos := strings.Index(output, "[rule-1]")
	rule2Pos := strings.Index(output, "[rule-2]")
	rule3Pos := strings.Index(output, "[rule-3]")
	assert.Less(t, rule1Pos, rule2Pos)
	assert.Less(t, rule2Pos, rule3Pos)
}
