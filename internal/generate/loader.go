package generate

import (
	"fmt"
	"os"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/util"
)

// Load reads and parses an OpenAPI specification file.
func Load(filePath string) (*openapi3.T, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", filePath)
	}

	spec, err := util.LoadSwagger(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OpenAPI spec: %w", err)
	}

	return spec, nil
}
