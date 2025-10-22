package internal_test

import (
	"testing"

	"github.com/duh-rpc/duhrpc-lint/internal"
	"github.com/stretchr/testify/require"
)

func TestLoadValidFile(t *testing.T) {
	doc, err := internal.Load("../testdata/minimal-valid.yaml")
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.NotNil(t, doc.Paths)
	require.NotNil(t, doc.Paths.PathItems)
}

func TestLoadFileNotFound(t *testing.T) {
	doc, err := internal.Load("nonexistent.yaml")
	require.Error(t, err)
	require.Nil(t, doc)
	require.ErrorContains(t, err, "file not found")
}

func TestLoadInvalidYAML(t *testing.T) {
	doc, err := internal.Load("../testdata/invalid-syntax.yaml")
	require.Error(t, err)
	require.Nil(t, doc)
	require.ErrorContains(t, err, "failed to parse OpenAPI spec")
}
