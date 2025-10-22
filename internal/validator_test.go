package internal_test

import (
	"testing"

	"github.com/duh-rpc/duhrpc-lint/internal"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateEmptyRules(t *testing.T) {
	doc := &v3.Document{
		Version: "3.0.0",
		Info: &base.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Paths: &v3.Paths{},
	}

	result := internal.Validate(doc, "test.yaml")

	require.NotNil(t, result)
	assert.Equal(t, "test.yaml", result.FilePath)
	assert.Empty(t, result.Violations)
	assert.True(t, result.Valid())
}
