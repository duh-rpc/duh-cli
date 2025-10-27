package duh

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/duh-rpc/duh-cli/internal/lint"
)

func Run(config RunConfig) error {
	spec, err := lint.Load(config.SpecPath)
	if err != nil {
		return err
	}

	result := lint.Validate(spec, config.SpecPath)
	if !result.Valid() {
		return fmt.Errorf("OpenAPI validation failed")
	}

	isFullTemplate := IsInitTemplateSpec(spec)

	genConfig, err := NewConfig(config.PackageName, config.OutputDir, config.ProtoPath, config.ProtoImport, config.ProtoPackage)
	if err != nil {
		return err
	}

	parser := NewParser(spec, genConfig, isFullTemplate)
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

	serverPath := filepath.Join(config.OutputDir, "server.go")
	if err := writeFile(serverPath, serverCode); err != nil {
		return fmt.Errorf("failed to write server.go: %w", err)
	}

	filesGenerated := []string{"server.go"}

	if data.HasListOps {
		iteratorCode, err := generator.RenderIterator(data)
		if err != nil {
			return fmt.Errorf("failed to render iterator.go: %w", err)
		}

		iteratorPath := filepath.Join(config.OutputDir, "iterator.go")
		if err := writeFile(iteratorPath, iteratorCode); err != nil {
			return fmt.Errorf("failed to write iterator.go: %w", err)
		}

		filesGenerated = append(filesGenerated, "iterator.go")
	}

	clientCode, err := generator.RenderClient(data)
	if err != nil {
		return fmt.Errorf("failed to render client.go: %w", err)
	}

	clientPath := filepath.Join(config.OutputDir, "client.go")
	if err := writeFile(clientPath, clientCode); err != nil {
		return fmt.Errorf("failed to write client.go: %w", err)
	}

	filesGenerated = append(filesGenerated, "client.go")

	specContent, err := os.ReadFile(config.SpecPath)
	if err != nil {
		return fmt.Errorf("failed to read OpenAPI spec: %w", err)
	}

	protoCode, err := config.Converter.Convert(specContent, data.ProtoPackage, data.ProtoImport)
	if err != nil {
		return fmt.Errorf("failed to convert OpenAPI to proto: %w", err)
	}

	protoFilePath := filepath.Join(config.OutputDir, config.ProtoPath)
	if err := writeFile(protoFilePath, protoCode); err != nil {
		return fmt.Errorf("failed to write proto file: %w", err)
	}

	filesGenerated = append(filesGenerated, config.ProtoPath)

	bufYamlPath := filepath.Join(config.OutputDir, "buf.yaml")
	if _, err := os.Stat(bufYamlPath); os.IsNotExist(err) {
		bufYamlCode, err := generator.RenderBufYaml(data)
		if err != nil {
			return fmt.Errorf("failed to render buf.yaml: %w", err)
		}

		if err := writeFile(bufYamlPath, bufYamlCode); err != nil {
			return fmt.Errorf("failed to write buf.yaml: %w", err)
		}

		filesGenerated = append(filesGenerated, "buf.yaml")
	}

	bufGenYamlPath := filepath.Join(config.OutputDir, "buf.gen.yaml")
	if _, err := os.Stat(bufGenYamlPath); os.IsNotExist(err) {
		bufGenYamlCode, err := generator.RenderBufGenYaml(data)
		if err != nil {
			return fmt.Errorf("failed to render buf.gen.yaml: %w", err)
		}

		if err := writeFile(bufGenYamlPath, bufGenYamlCode); err != nil {
			return fmt.Errorf("failed to write buf.gen.yaml: %w", err)
		}

		filesGenerated = append(filesGenerated, "buf.gen.yaml")
	}

	if config.FullFlag {
		daemonCode, err := generator.RenderDaemon(data)
		if err != nil {
			return fmt.Errorf("failed to render daemon.go: %w", err)
		}

		daemonPath := filepath.Join(config.OutputDir, "daemon.go")
		if err := writeFile(daemonPath, daemonCode); err != nil {
			return fmt.Errorf("failed to write daemon.go: %w", err)
		}

		filesGenerated = append(filesGenerated, "daemon.go")

		serviceCode, err := generator.RenderService(data)
		if err != nil {
			return fmt.Errorf("failed to render service.go: %w", err)
		}

		servicePath := filepath.Join(config.OutputDir, "service.go")
		if err := writeFile(servicePath, serviceCode); err != nil {
			return fmt.Errorf("failed to write service.go: %w", err)
		}

		filesGenerated = append(filesGenerated, "service.go")

		apiTestCode, err := generator.RenderApiTest(data)
		if err != nil {
			return fmt.Errorf("failed to render api_test.go: %w", err)
		}

		apiTestPath := filepath.Join(config.OutputDir, "api_test.go")
		if err := writeFile(apiTestPath, apiTestCode); err != nil {
			return fmt.Errorf("failed to write api_test.go: %w", err)
		}

		filesGenerated = append(filesGenerated, "api_test.go")

		makefileCode, err := generator.RenderMakefile(data)
		if err != nil {
			return fmt.Errorf("failed to render Makefile: %w", err)
		}

		makefilePath := filepath.Join(config.OutputDir, "Makefile")
		if err := writeFile(makefilePath, makefileCode); err != nil {
			return fmt.Errorf("failed to write Makefile: %w", err)
		}

		filesGenerated = append(filesGenerated, "Makefile")
	}

	_, _ = fmt.Fprintf(config.Writer, "✓ Generated %d file(s) in %s\n", len(filesGenerated), config.OutputDir)
	for _, file := range filesGenerated {
		_, _ = fmt.Fprintf(config.Writer, "  - %s\n", file)
	}

	_, _ = fmt.Fprintf(config.Writer, "\nNext steps:\n")
	_, _ = fmt.Fprintf(config.Writer, "  1. Run 'buf generate' to generate Go code from proto files\n")
	_, _ = fmt.Fprintf(config.Writer, "  2. Run 'go mod tidy' to update dependencies\n")

	return nil
}

func writeFile(path string, content []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, content, 0644)
}
