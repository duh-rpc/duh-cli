package duh_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/duh-rpc/duh-cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLinterValidSpec(t *testing.T) {
	var stdout bytes.Buffer

	exitCode := duh.RunCmd(&stdout, []string{"lint", "internal/lint/testdata/valid-spec.yaml"})

	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "✓")
	assert.Contains(t, stdout.String(), "compliant")
}

func TestLinterAllRuleViolations(t *testing.T) {
	tests := []struct {
		name              string
		file              string
		expectedViolation string
		expectedExitCode  int
	}{
		{
			name:              "BadPathFormat",
			file:              "internal/lint/testdata/bad-path-format.yaml",
			expectedViolation: "[path-format]",
			expectedExitCode:  1,
		},
		{
			name:              "WrongHTTPMethod",
			file:              "internal/lint/testdata/wrong-http-method.yaml",
			expectedViolation: "[http-method]",
			expectedExitCode:  1,
		},
		{
			name:              "HasQueryParams",
			file:              "internal/lint/testdata/has-query-params.yaml",
			expectedViolation: "[query-parameters]",
			expectedExitCode:  1,
		},
		{
			name:              "MissingRequestBody",
			file:              "internal/lint/testdata/missing-request-body.yaml",
			expectedViolation: "[request-body-required]",
			expectedExitCode:  1,
		},
		{
			name:              "InvalidStatusCode",
			file:              "internal/lint/testdata/invalid-status-code.yaml",
			expectedViolation: "[status-code]",
			expectedExitCode:  1,
		},
		{
			name:              "MissingSuccessResponse",
			file:              "internal/lint/testdata/missing-success-response.yaml",
			expectedViolation: "[success-response]",
			expectedExitCode:  1,
		},
		{
			name:              "InvalidContentType",
			file:              "internal/lint/testdata/invalid-content-type.yaml",
			expectedViolation: "[content-type]",
			expectedExitCode:  1,
		},
		{
			name:              "BadErrorSchema",
			file:              "internal/lint/testdata/bad-error-schema.yaml",
			expectedViolation: "[error-response-schema]",
			expectedExitCode:  1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var stdout bytes.Buffer

			exitCode := duh.RunCmd(&stdout, []string{"lint", test.file})

			assert.Equal(t, test.expectedExitCode, exitCode)
			output := stdout.String()
			assert.Contains(t, output, test.expectedViolation)
			assert.Contains(t, output, "ERRORS FOUND:")
			assert.Contains(t, output, "Summary:")
		})
	}
}

func TestLinterMultipleViolations(t *testing.T) {
	var stdout bytes.Buffer

	exitCode := duh.RunCmd(&stdout, []string{"lint", "internal/lint/testdata/multiple-violations.yaml"})

	require.Equal(t, 1, exitCode)

	output := stdout.String()

	// Verify expected violations are reported
	expectedViolations := []string{
		"[path-format]",
		"[http-method]",
		"[query-parameters]",
		"[request-body-required]",
		"[status-code]",
		"[content-type]",
		"[error-response-schema]",
		"[success-response]",
	}

	for _, violation := range expectedViolations {
		assert.Contains(t, output, violation)
	}

	// Verify output structure
	assert.Contains(t, output, "ERRORS FOUND:")
	assert.Contains(t, output, "Summary:")

	// Count violations - should have at least 8 (one for each rule type)
	violationCount := strings.Count(output, "[")
	assert.GreaterOrEqual(t, violationCount, 8)
}

func TestLinterFileNotFound(t *testing.T) {
	var stdout bytes.Buffer

	exitCode := duh.RunCmd(&stdout, []string{"lint", "nonexistent.yaml"})

	require.Equal(t, 2, exitCode)
	assert.Contains(t, stdout.String(), "file not found")
}

func TestLinterInvalidYAML(t *testing.T) {
	var stdout bytes.Buffer

	exitCode := duh.RunCmd(&stdout, []string{"lint", "internal/lint/testdata/invalid-syntax.yaml"})

	require.Equal(t, 2, exitCode)
	assert.Contains(t, stdout.String(), "failed to parse OpenAPI spec")
}
