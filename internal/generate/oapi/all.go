package oapi

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/oapi-codegen/oapi-codegen/v2/pkg/codegen"
)

// RunAll generates client, server, and models code from an OpenAPI specification.
func RunAll(w io.Writer, filePath, outputDir, packageName string) error {
	spec, err := Load(filePath)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	clientConfig, err := NewConfig(packageName, codegen.GenerateOptions{
		Client: true,
	})
	if err != nil {
		return err
	}

	if err := Generate(spec, clientConfig, filepath.Join(outputDir, "client.go")); err != nil {
		return err
	}

	serverConfig, err := NewConfig(packageName, codegen.GenerateOptions{
		StdHTTPServer: true,
	})
	if err != nil {
		return err
	}

	if err := Generate(spec, serverConfig, filepath.Join(outputDir, "server.go")); err != nil {
		return err
	}

	modelsConfig, err := NewConfig(packageName, codegen.GenerateOptions{
		Models: true,
	})
	if err != nil {
		return err
	}

	if err := Generate(spec, modelsConfig, filepath.Join(outputDir, "models.go")); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(w, "✓ Generated client, server, and models in %s\n", outputDir)
	return nil
}
