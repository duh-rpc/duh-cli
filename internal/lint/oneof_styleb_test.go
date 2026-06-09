package lint_test

import (
	"bytes"
	"context"
	"testing"

	duh "github.com/duh-rpc/duh-cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLintStyleBOneOfPasses verifies a valid wire-compatible style-B union (one
// optional $ref property per variant, a oneOf of single-required branches, no
// discriminator) passes lint with exit code 0 and produces no PROHIBITED_ONEOF or
// DISCRIMINATOR_* violation. (Blueprint AC 1 and AC 5.)
func TestLintStyleBOneOfPasses(t *testing.T) {
	var stdout bytes.Buffer

	exitCode := duh.RunCmd(context.Background(), &stdout, []string{"lint", "testdata/oneof-style-b.yaml"})

	require.Equal(t, 0, exitCode)
	output := stdout.String()
	assert.Contains(t, output, "compliant")
	assert.NotContains(t, output, "[PROHIBITED_ONEOF]")
	assert.NotContains(t, output, "[DISCRIMINATOR_")
}

// TestLintFlatDiscriminatedOneOfFails verifies the flat/discriminated oneOf (a
// top-level oneOf of $refs with a discriminator) still fails lint with
// [PROHIBITED_ONEOF] and a non-zero exit. (Blueprint AC 3.)
func TestLintFlatDiscriminatedOneOfFails(t *testing.T) {
	var stdout bytes.Buffer

	exitCode := duh.RunCmd(context.Background(), &stdout, []string{"lint", "testdata/oneof-flat-discriminated.yaml"})

	require.Equal(t, 1, exitCode)
	output := stdout.String()
	assert.Contains(t, output, "[PROHIBITED_ONEOF]")
	assert.NotContains(t, output, "[DISCRIMINATOR_")
}

// TestLintMalformedStyleBBranchMultipleRequired verifies a style-B branch naming
// more than one required property is rejected with a clear message and non-zero exit.
// (Blueprint AC 4.)
func TestLintMalformedStyleBBranchMultipleRequired(t *testing.T) {
	var stdout bytes.Buffer

	exitCode := duh.RunCmd(context.Background(), &stdout, []string{"lint", "testdata/oneof-style-b-multi-required.yaml"})

	require.Equal(t, 1, exitCode)
	output := stdout.String()
	assert.Contains(t, output, "[PROHIBITED_ONEOF]")
	assert.Contains(t, output, "exactly one required property")
}

// TestLintMalformedStyleBUndeclaredProperty verifies a style-B branch naming a
// property absent from properties is rejected with a clear message and non-zero exit.
// (Blueprint AC 4.)
func TestLintMalformedStyleBUndeclaredProperty(t *testing.T) {
	var stdout bytes.Buffer

	exitCode := duh.RunCmd(context.Background(), &stdout, []string{"lint", "testdata/oneof-style-b-undeclared.yaml"})

	require.Equal(t, 1, exitCode)
	output := stdout.String()
	assert.Contains(t, output, "[PROHIBITED_ONEOF]")
	assert.Contains(t, output, "bird_event")
	assert.Contains(t, output, "not declared in properties")
}

// TestLintMalformedStyleBArrayVariant verifies a style-B variant property whose
// schema is an array is rejected with a clear message and non-zero exit, since proto3
// forbids a repeated field inside a oneof. (Blueprint AC 4.)
func TestLintMalformedStyleBArrayVariant(t *testing.T) {
	var stdout bytes.Buffer

	exitCode := duh.RunCmd(context.Background(), &stdout, []string{"lint", "testdata/oneof-style-b-array-variant.yaml"})

	require.Equal(t, 1, exitCode)
	output := stdout.String()
	assert.Contains(t, output, "[PROHIBITED_ONEOF]")
	assert.Contains(t, output, "cat_event")
	assert.Contains(t, output, "array")
}
