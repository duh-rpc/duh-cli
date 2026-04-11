package lint_test

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

	exitCode := duh.RunCmd(&stdout, []string{"lint", "testdata/valid-spec.yaml"})

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
			file:              "testdata/bad-path-format.yaml",
			expectedViolation: "[PATH_FORMAT]",
			expectedExitCode:  1,
		},
		{
			name:              "WrongHTTPMethod",
			file:              "testdata/wrong-http-method.yaml",
			expectedViolation: "[HTTP_METHOD_ALLOWED]",
			expectedExitCode:  1,
		},
		{
			name:              "HasQueryParams",
			file:              "testdata/has-query-params.yaml",
			expectedViolation: "[RPC_POST_NO_QUERY_PARAMS]",
			expectedExitCode:  1,
		},
		{
			name:              "MissingRequestBody",
			file:              "testdata/missing-request-body.yaml",
			expectedViolation: "[REQUEST_BODY_REQUIRED]",
			expectedExitCode:  1,
		},
		{
			name:              "InvalidStatusCode",
			file:              "testdata/invalid-status-code.yaml",
			expectedViolation: "[STATUS_CODE_ALLOWED]",
			expectedExitCode:  1,
		},
		{
			name:              "MissingSuccessResponse",
			file:              "testdata/missing-success-response.yaml",
			expectedViolation: "[SUCCESS_RESPONSE]",
			expectedExitCode:  1,
		},
		{
			name:              "InvalidContentType",
			file:              "testdata/invalid-content-type.yaml",
			expectedViolation: "[CONTENT_TYPE]",
			expectedExitCode:  1,
		},
		{
			name:              "BadErrorSchema",
			file:              "testdata/bad-error-schema.yaml",
			expectedViolation: "[ERROR_SCHEMA]",
			expectedExitCode:  1,
		},
		{
			name:              "VersionPrefixInPath",
			file:              "testdata/version-prefix-in-path.yaml",
			expectedViolation: "[PATH_NO_VERSION_PREFIX]",
			expectedExitCode:  1,
		},
		{
			name:              "MissingServerVersion",
			file:              "testdata/missing-server-version.yaml",
			expectedViolation: "[SERVER_URL_VERSIONING]",
			expectedExitCode:  1,
		},
		{
			name:              "PaginationUsesLimit",
			file:              "testdata/pagination-uses-limit.yaml",
			expectedViolation: "[RPC_PAGINATION_PARAMETERS]",
			expectedExitCode:  1,
		},
		{
			name:              "PaginationNotNested",
			file:              "testdata/pagination-not-nested.yaml",
			expectedViolation: "[RPC_PAGINATED_REQUEST_STRUCTURE]",
			expectedExitCode:  1,
		},
		{
			name:              "BadRequestName",
			file:              "testdata/bad-request-name.yaml",
			expectedViolation: "[RPC_REQUEST_STANDARD_NAME]",
			expectedExitCode:  1,
		},
		{
			name:              "BadResponseName",
			file:              "testdata/bad-response-name.yaml",
			expectedViolation: "[RPC_RESPONSE_STANDARD_NAME]",
			expectedExitCode:  1,
		},
		{
			name:              "SharedSchemas",
			file:              "testdata/shared-schemas.yaml",
			expectedViolation: "[RPC_REQUEST_RESPONSE_UNIQUE]",
			expectedExitCode:  1,
		},
		{
			name:              "MissingIntegerFormat",
			file:              "testdata/missing-integer-format.yaml",
			expectedViolation: "[RPC_INTEGER_FORMAT_REQUIRED]",
			expectedExitCode:  1,
		},
		{
			name:              "NullableProperty",
			file:              "testdata/nullable-property.yaml",
			expectedViolation: "[RPC_NO_NULLABLE]",
			expectedExitCode:  1,
		},
		{
			name:              "NestedArray",
			file:              "testdata/nested-array.yaml",
			expectedViolation: "[RPC_NO_NESTED_ARRAYS]",
			expectedExitCode:  1,
		},
		{
			name:              "UntypedAdditionalProperties",
			file:              "testdata/untyped-additional-properties.yaml",
			expectedViolation: "[RPC_TYPED_ADDITIONAL_PROPERTIES]",
			expectedExitCode:  1,
		},
		{
			name:              "UsesOneOf",
			file:              "testdata/uses-oneof.yaml",
			expectedViolation: "[RPC_PROHIBITED_ONEOF_AND_ALLOF]",
			expectedExitCode:  1,
		},
		{
			name:              "BadPaginatedResponse",
			file:              "testdata/bad-paginated-response.yaml",
			expectedViolation: "[RESPONSE_PAGINATED_STRUCTURE]",
			expectedExitCode:  1,
		},
		{
			name:              "BadPaginationParams",
			file:              "testdata/bad-pagination-params.yaml",
			expectedViolation: "[PAGINATION_PARAMETERS]",
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
			assert.Contains(t, output, "errors")
		})
	}
}

func TestLinterMultipleViolations(t *testing.T) {
	var stdout bytes.Buffer

	exitCode := duh.RunCmd(&stdout, []string{"lint", "testdata/multiple-violations.yaml"})

	require.Equal(t, 1, exitCode)

	output := stdout.String()

	// Verify expected violations are reported
	expectedViolations := []string{
		"[PATH_FORMAT]",
		"[PATH_NO_VERSION_PREFIX]",
		"[HTTP_METHOD_ALLOWED]",
		"[RPC_POST_NO_QUERY_PARAMS]",
		"[REQUEST_BODY_REQUIRED]",
		"[STATUS_CODE_ALLOWED]",
		"[CONTENT_TYPE]",
		"[ERROR_SCHEMA]",
		"[SUCCESS_RESPONSE]",
		"[RPC_REQUEST_STANDARD_NAME]",
		"[RPC_RESPONSE_STANDARD_NAME]",
		"[RPC_REQUEST_RESPONSE_UNIQUE]",
		"[RPC_INTEGER_FORMAT_REQUIRED]",
		"[RPC_NO_NULLABLE]",
		"[RPC_NO_NESTED_ARRAYS]",
		"[RPC_TYPED_ADDITIONAL_PROPERTIES]",
		"[RPC_PROHIBITED_ONEOF_AND_ALLOF]",
		"[RESPONSE_PAGINATED_STRUCTURE]",
		"[PAGINATION_PARAMETERS]",
	}

	for _, violation := range expectedViolations {
		assert.Contains(t, output, violation)
	}

	// Verify output structure
	assert.Contains(t, output, "errors")

	// Count violations - should have at least 19 (one for each rule type)
	violationCount := strings.Count(output, "[ERROR]")
	assert.GreaterOrEqual(t, violationCount, 19)
}

func TestLinterFileNotFound(t *testing.T) {
	var stdout bytes.Buffer

	exitCode := duh.RunCmd(&stdout, []string{"lint", "nonexistent.yaml"})

	require.Equal(t, 2, exitCode)
	assert.Contains(t, stdout.String(), "file not found")
}

func TestLinterInvalidYAML(t *testing.T) {
	var stdout bytes.Buffer

	exitCode := duh.RunCmd(&stdout, []string{"lint", "testdata/invalid-syntax.yaml"})

	require.Equal(t, 2, exitCode)
	assert.Contains(t, stdout.String(), "failed to parse OpenAPI spec")
}
