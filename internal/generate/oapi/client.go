package oapi

import (
	"fmt"
	"io"

	"github.com/oapi-codegen/oapi-codegen/v2/pkg/codegen"
)

// RunClient generates HTTP client code from an OpenAPI specification.
func RunClient(w io.Writer, filePath, outputPath, packageName string) error {
	spec, err := Load(filePath)
	if err != nil {
		return err
	}

	config, err := NewConfig(packageName, codegen.GenerateOptions{
		Client: true,
	})
	if err != nil {
		return err
	}

	if err := Generate(spec, config, outputPath); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(w, "✓ Generated client code at %s\n", outputPath)
	return nil
}
