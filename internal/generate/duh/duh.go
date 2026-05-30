package duh

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/duh-rpc/duh-cli/internal/fieldmap"
	"github.com/duh-rpc/duh-cli/internal/lint"
)

// artifact is a rendered file held in memory until every artifact (and the lock
// reconciliation) has succeeded, so a failure writes nothing.
type artifact struct {
	path         string
	name         string
	content      []byte
	skipIfExists bool
}

func Run(config RunConfig) error {
	spec, err := lint.Load(config.SpecPath)
	if err != nil {
		return err
	}

	result := lint.Validate(spec, config.SpecPath, nil)
	if !result.Valid() {
		return fmt.Errorf("OpenAPI validation failed")
	}

	isFullTemplate := IsInitTemplateSpec(spec)

	genConfig, err := NewConfig(config.PackageName, config.OutputDir, config.ProtoPath, config.ProtoImport, config.ProtoPackage)
	if err != nil {
		return err
	}

	// The lock is co-located with the spec by default and never written under
	// --output-dir. A nonexistent parent directory errors, matching --output-dir.
	lockPath := config.LockPath
	if lockPath == "" {
		lockPath = filepath.Join(filepath.Dir(config.SpecPath), "fieldmap.lock")
	}
	if _, err := os.Stat(filepath.Dir(lockPath)); os.IsNotExist(err) {
		return fmt.Errorf("lock directory does not exist: %s", filepath.Dir(lockPath))
	}

	// Reconcile the lock before any conversion. A corrupt lock or a forced
	// reassignment fails loud here, before a single file is written.
	existingLock, err := fieldmap.Load(lockPath)
	if err != nil {
		return err
	}

	nextLock, fieldNumbers, err := fieldmap.Reconcile(spec, existingLock)
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

	var artifacts []artifact

	serverCode, err := generator.RenderServer(data)
	if err != nil {
		return fmt.Errorf("failed to render server.go: %w", err)
	}
	artifacts = append(artifacts, artifact{path: filepath.Join(config.OutputDir, "server.go"), name: "server.go", content: serverCode})

	clientCode, err := generator.RenderClient(data)
	if err != nil {
		return fmt.Errorf("failed to render client.go: %w", err)
	}
	artifacts = append(artifacts, artifact{path: filepath.Join(config.OutputDir, "client.go"), name: "client.go", content: clientCode})

	specContent, err := os.ReadFile(config.SpecPath)
	if err != nil {
		return fmt.Errorf("failed to read OpenAPI spec: %w", err)
	}

	protoCode, err := config.Converter.Convert(specContent, data.ProtoPackage, data.ProtoImport, fieldNumbers)
	if err != nil {
		return fmt.Errorf("failed to convert OpenAPI to proto: %w", err)
	}
	artifacts = append(artifacts, artifact{path: filepath.Join(config.OutputDir, config.ProtoPath), name: config.ProtoPath, content: protoCode})

	bufYamlCode, err := generator.RenderBufYaml(data)
	if err != nil {
		return fmt.Errorf("failed to render buf.yaml: %w", err)
	}
	artifacts = append(artifacts, artifact{path: filepath.Join(config.OutputDir, "buf.yaml"), name: "buf.yaml", content: bufYamlCode, skipIfExists: true})

	bufGenYamlCode, err := generator.RenderBufGenYaml(data)
	if err != nil {
		return fmt.Errorf("failed to render buf.gen.yaml: %w", err)
	}
	artifacts = append(artifacts, artifact{path: filepath.Join(config.OutputDir, "buf.gen.yaml"), name: "buf.gen.yaml", content: bufGenYamlCode, skipIfExists: true})

	if config.FullFlag {
		daemonCode, err := generator.RenderDaemon(data)
		if err != nil {
			return fmt.Errorf("failed to render daemon.go: %w", err)
		}
		artifacts = append(artifacts, artifact{path: filepath.Join(config.OutputDir, "daemon.go"), name: "daemon.go", content: daemonCode})

		serviceCode, err := generator.RenderService(data)
		if err != nil {
			return fmt.Errorf("failed to render service.go: %w", err)
		}
		artifacts = append(artifacts, artifact{path: filepath.Join(config.OutputDir, "service.go"), name: "service.go", content: serviceCode})

		apiTestCode, err := generator.RenderApiTest(data)
		if err != nil {
			return fmt.Errorf("failed to render api_test.go: %w", err)
		}
		artifacts = append(artifacts, artifact{path: filepath.Join(config.OutputDir, "api_test.go"), name: "api_test.go", content: apiTestCode})

		makefileCode, err := generator.RenderMakefile(data)
		if err != nil {
			return fmt.Errorf("failed to render Makefile: %w", err)
		}
		artifacts = append(artifacts, artifact{path: filepath.Join(config.OutputDir, "Makefile"), name: "Makefile", content: makefileCode})
	}

	// Everything rendered and reconciled successfully — now write. The lock is the
	// one checked-in artifact and is reported separately from the generated output.
	var filesGenerated []string
	for _, a := range artifacts {
		if a.skipIfExists {
			if _, err := os.Stat(a.path); err == nil {
				continue
			}
		}
		if err := writeFile(a.path, a.content); err != nil {
			return fmt.Errorf("failed to write %s: %w", a.name, err)
		}
		filesGenerated = append(filesGenerated, a.name)
	}

	if err := nextLock.Save(lockPath); err != nil {
		return fmt.Errorf("failed to write fieldmap.lock: %w", err)
	}

	_, _ = fmt.Fprintf(config.Writer, "✓ Generated %d file(s) in %s\n", len(filesGenerated), config.OutputDir)
	for _, file := range filesGenerated {
		_, _ = fmt.Fprintf(config.Writer, "  - %s\n", file)
	}
	_, _ = fmt.Fprintf(config.Writer, "✓ Wrote fieldmap.lock to %s\n", lockPath)

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
