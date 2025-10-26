package duh

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/duh-rpc/duh-cli/internal/lint"
)

func Run(w io.Writer, specPath, packageName, outputDir, protoPath, protoImport, protoPackage string) error {
	spec, err := lint.Load(specPath)
	if err != nil {
		return err
	}

	result := lint.Validate(spec, specPath)
	if !result.Valid() {
		return fmt.Errorf("OpenAPI validation failed")
	}

	config, err := NewConfig(packageName, outputDir, protoPath, protoImport, protoPackage)
	if err != nil {
		return err
	}

	parser := NewParser(spec, config)
	data, err := parser.Parse()
	if err != nil {
		return err
	}

	generator, err := NewGenerator()
	if err != nil {
		return fmt.Errorf("failed to create generator: %w", err)
	}

	serverCode, err := generator.RenderServer(data)
	if err != nil {
		return fmt.Errorf("failed to render server.go: %w", err)
	}

	serverPath := filepath.Join(outputDir, "server.go")
	if err := writeFile(serverPath, serverCode); err != nil {
		return fmt.Errorf("failed to write server.go: %w", err)
	}

	_, _ = fmt.Fprintf(w, "✓ Generated 1 file in %s\n", outputDir)
	_, _ = fmt.Fprintf(w, "  - server.go\n")

	return nil
}

func writeFile(path string, content []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, content, 0644)
}
