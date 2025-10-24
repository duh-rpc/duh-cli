package oapi

import (
	"fmt"
	"io"

	"github.com/oapi-codegen/oapi-codegen/v2/pkg/codegen"
)

// RunServer generates HTTP server stub code from an OpenAPI specification.
func RunServer(w io.Writer, filePath, outputPath, packageName string) error {
	spec, err := Load(filePath)
	if err != nil {
		return err
	}

	config, err := NewConfig(packageName, codegen.GenerateOptions{
		StdHTTPServer: true,
	})
	if err != nil {
		return err
	}

	if err := Generate(spec, config, outputPath); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(w, "✓ Generated server code at %s\n", outputPath)
	return nil
}
