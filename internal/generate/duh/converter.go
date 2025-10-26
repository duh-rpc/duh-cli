package duh

import (
	"bytes"
	"fmt"
	"sort"

	"gopkg.in/yaml.v3"
)

type ProtoConverter interface {
	Convert(openapi []byte, packageName string) ([]byte, error)
}

type MockProtoConverter struct{}

func NewMockProtoConverter() *MockProtoConverter {
	return &MockProtoConverter{}
}

func (m *MockProtoConverter) Convert(openapi []byte, packageName string) ([]byte, error) {
	schemas, err := extractSchemaNames(openapi)
	if err != nil {
		return nil, fmt.Errorf("failed to extract schema names: %w", err)
	}

	return generateProtoFile(packageName, schemas), nil
}

func extractSchemaNames(openapi []byte) ([]string, error) {
	var doc map[string]interface{}
	if err := yaml.Unmarshal(openapi, &doc); err != nil {
		return nil, err
	}

	components, ok := doc["components"].(map[string]interface{})
	if !ok {
		return []string{}, nil
	}

	schemas, ok := components["schemas"].(map[string]interface{})
	if !ok {
		return []string{}, nil
	}

	names := make([]string, 0, len(schemas))
	for name := range schemas {
		names = append(names, name)
	}

	sort.Strings(names)
	return names, nil
}

func generateProtoFile(packageName string, schemas []string) []byte {
	var buf bytes.Buffer

	buf.WriteString("syntax = \"proto3\";\n\n")
	buf.WriteString(fmt.Sprintf("package %s;\n\n", packageName))

	for _, schema := range schemas {
		buf.WriteString(fmt.Sprintf("message %s {}\n", schema))
	}

	return buf.Bytes()
}
