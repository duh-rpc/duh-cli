package lint_test

import (
	"testing"

	"github.com/duh-rpc/duhrpc-lint/internal/lint"
	"github.com/pb33f/libopenapi"
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

	result := lint.Validate(doc, "test.yaml")

	require.NotNil(t, result)
	assert.Equal(t, "test.yaml", result.FilePath)
	assert.Empty(t, result.Violations)
	assert.True(t, result.Valid())
}

func TestValidateWithRules(t *testing.T) {
	spec := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /invalid-path:
    get:
      parameters:
        - name: query
          in: query
          schema:
            type: string
      responses:
        200:
          description: Success
          content:
            application/json:
              schema:
                type: object`

	doc, err := libopenapi.NewDocument([]byte(spec))
	require.NoError(t, err)

	model, errs := doc.BuildV3Model()
	require.Empty(t, errs)

	result := lint.Validate(&model.Model, "test.yaml")

	require.NotNil(t, result)
	assert.Equal(t, "test.yaml", result.FilePath)
	assert.NotEmpty(t, result.Violations)
	assert.False(t, result.Valid())
	assert.Len(t, result.Violations, 3)
}
