package oapi

import (
	"fmt"
	"io"

	"github.com/oapi-codegen/oapi-codegen/v2/pkg/codegen"
)

func RunModels(w io.Writer, filePath, outputPath, packageName string) error {
	spec, err := Load(filePath)
	if err != nil {
		return err
	}

	config, err := NewConfig(packageName, codegen.GenerateOptions{
		Models: true,
	})
	if err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}

	if err := Generate(spec, config, outputPath); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(w, "✓ Generated models code at %s\n", outputPath)
	return nil
}
