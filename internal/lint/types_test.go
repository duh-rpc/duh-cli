package lint_test

import (
	"testing"

	"github.com/duh-rpc/duh-cli/internal/lint"
	"github.com/stretchr/testify/assert"
)

func TestViolationString(t *testing.T) {
	for _, test := range []struct {
		name       string
		violation  lint.Violation
		wantString string
	}{
		{
			name: "AllFields",
			violation: lint.Violation{
				Suggestion: "Use POST instead of GET",
				RuleName:   "http-method",
				Location:   "GET /v1/users.list",
				Message:    "Only POST method is allowed",
			},
			wantString: "[http-method] GET /v1/users.list\n  Only POST method is allowed\n  Use POST instead of GET",
		},
		{
			name: "MinimalFields",
			violation: lint.Violation{
				RuleName: "test-rule",
				Location: "/v1/test.endpoint",
				Message:  "Test message",
			},
			wantString: "[test-rule] /v1/test.endpoint\n  Test message\n  ",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result := test.violation.String()
			assert.Equal(t, test.wantString, result)
		})
	}
}

func TestValidationResultValid(t *testing.T) {
	for _, test := range []struct {
		name      string
		result    lint.ValidationResult
		wantValid bool
	}{
		{
			name: "NoViolations",
			result: lint.ValidationResult{
				FilePath: "test.yaml",
			},
			wantValid: true,
		},
		{
			name: "WithViolations",
			result: lint.ValidationResult{
				FilePath: "test.yaml",
				Violations: []lint.Violation{
					{
						RuleName: "test-rule",
						Location: "/v1/test",
						Message:  "Test violation",
					},
				},
			},
			wantValid: false,
		},
		{
			name: "MultipleViolations",
			result: lint.ValidationResult{
				FilePath: "test.yaml",
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
				},
			},
			wantValid: false,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result := test.result.Valid()
			assert.Equal(t, test.wantValid, result)
		})
	}
}
