package add

import (
	"fmt"
	"io"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

var pathFormatRegex = regexp.MustCompile(`^/v(0|[1-9][0-9]*)/[a-z][a-z0-9_-]{0,49}\.[a-z][a-z0-9_-]{0,49}$`)

func Run(w io.Writer, filePath, path, name string) error {
	if !pathFormatRegex.MatchString(path) {
		return fmt.Errorf("invalid path format: %s (must follow /v{N}/{subject}.{method})", path)
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filePath)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	if root.Kind != yaml.DocumentNode || len(root.Content) == 0 {
		return fmt.Errorf("invalid OpenAPI document structure")
	}

	doc := root.Content[0]

	pathsNode, err := findOrCreateNode(doc, "paths")
	if err != nil {
		return fmt.Errorf("failed to find or create paths: %w", err)
	}

	if pathExists(pathsNode, path) {
		return fmt.Errorf("path already exists: %s", path)
	}

	componentsNode, err := findOrCreateNode(doc, "components")
	if err != nil {
		return fmt.Errorf("failed to find or create components: %w", err)
	}

	schemasNode, err := findOrCreateNode(componentsNode, "schemas")
	if err != nil {
		return fmt.Errorf("failed to find or create schemas: %w", err)
	}

	addSchema(schemasNode, name+"Request", generateRequestSchema(name))
	addSchema(schemasNode, name+"Response", generateResponseSchema(name))

	addPath(pathsNode, path, generatePathItem(name))

	output, err := yaml.Marshal(&root)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	if err := os.WriteFile(filePath, output, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	_, _ = fmt.Fprintf(w, "✓ Added endpoint %s to %s\n", path, filePath)
	return nil
}

func findOrCreateNode(parent *yaml.Node, key string) (*yaml.Node, error) {
	if parent.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("parent is not a mapping node")
	}

	for i := 0; i < len(parent.Content); i += 2 {
		if parent.Content[i].Value == key {
			return parent.Content[i+1], nil
		}
	}

	keyNode := &yaml.Node{Kind: yaml.ScalarNode, Value: key}
	valueNode := &yaml.Node{Kind: yaml.MappingNode}
	parent.Content = append(parent.Content, keyNode, valueNode)
	return valueNode, nil
}

func pathExists(pathsNode *yaml.Node, path string) bool {
	if pathsNode.Kind != yaml.MappingNode {
		return false
	}

	for i := 0; i < len(pathsNode.Content); i += 2 {
		if pathsNode.Content[i].Value == path {
			return true
		}
	}
	return false
}

func addPath(pathsNode *yaml.Node, path string, pathItem *yaml.Node) {
	keyNode := &yaml.Node{Kind: yaml.ScalarNode, Value: path}
	pathsNode.Content = append(pathsNode.Content, keyNode, pathItem)
}

func addSchema(schemasNode *yaml.Node, name string, schema *yaml.Node) {
	keyNode := &yaml.Node{Kind: yaml.ScalarNode, Value: name}
	schemasNode.Content = append(schemasNode.Content, keyNode, schema)
}
