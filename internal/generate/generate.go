package generate

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/codegen"
)

// Generate generates Go code from an OpenAPI specification and writes it to a file.
func Generate(spec *openapi3.T, config codegen.Configuration, outputPath string) error {
	code, err := codegen.Generate(spec, config)
	if err != nil {
		return fmt.Errorf("code generation failed: %w", err)
	}

	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	if err := os.WriteFile(outputPath, []byte(code), 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}
