package duh

import (
	"bufio"
	"fmt"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Config struct {
	PackageName  string
	OutputDir    string
	ProtoPath    string
	ProtoImport  string
	ProtoPackage string
}

func NewConfig(packageName, outputDir, protoPath, protoImport, protoPackage string) (*Config, error) {
	if packageName == "" {
		packageName = "api"
	}
	if outputDir == "" {
		outputDir = "."
	}
	if protoPath == "" {
		protoPath = "proto/v1/api.proto"
	}

	config := &Config{
		PackageName:  packageName,
		OutputDir:    outputDir,
		ProtoPath:    protoPath,
		ProtoImport:  protoImport,
		ProtoPackage: protoPackage,
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

func (c *Config) Validate() error {
	if !token.IsIdentifier(c.PackageName) {
		return fmt.Errorf("invalid package name: %s", c.PackageName)
	}

	if c.PackageName == "main" {
		return fmt.Errorf("package name cannot be 'main'")
	}

	if _, err := os.Stat(c.OutputDir); os.IsNotExist(err) {
		return fmt.Errorf("output directory does not exist: %s", c.OutputDir)
	}

	return nil
}

func (c *Config) DetectModulePath() (string, error) {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return "", fmt.Errorf("failed to read go.mod: %w", err)
	}

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	moduleRegex := regexp.MustCompile(`^module\s+(.+)$`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if matches := moduleRegex.FindStringSubmatch(line); len(matches) > 1 {
			modulePath := strings.TrimSpace(matches[1])
			if idx := strings.Index(modulePath, "//"); idx != -1 {
				modulePath = strings.TrimSpace(modulePath[:idx])
			}
			if !strings.Contains(modulePath, "/") {
				return "", fmt.Errorf("invalid module path: must contain '/': %s", modulePath)
			}
			return modulePath, nil
		}
	}

	return "", fmt.Errorf("module declaration not found in go.mod")
}

func (c *Config) ConstructProtoImport(modulePath string) string {
	if c.ProtoImport != "" {
		return c.ProtoImport
	}
	return filepath.Join(modulePath, filepath.Dir(c.ProtoPath))
}

func (c *Config) DeriveProtoPackage() string {
	if c.ProtoPackage != "" {
		return c.ProtoPackage
	}

	protoDir := filepath.Dir(c.ProtoPath)
	parts := strings.Split(protoDir, string(filepath.Separator))

	for _, part := range parts {
		if strings.HasPrefix(part, "v") && len(part) > 1 {
			return fmt.Sprintf("duh.api.%s", part)
		}
	}

	return "duh.api.v1"
}
