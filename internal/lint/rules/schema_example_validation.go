package rules

import (
	"fmt"
	"slices"

	"github.com/pb33f/libopenapi/datamodel/high/v3"
	"go.yaml.in/yaml/v4"
)

type SchemaExampleValidationRule struct{}

func NewSchemaExampleValidationRule() *SchemaExampleValidationRule {
	return &SchemaExampleValidationRule{}
}

func (r *SchemaExampleValidationRule) Name() string {
	return "SCHEMA_EXAMPLE_VALIDATION"
}

func (r *SchemaExampleValidationRule) Validate(doc *v3.Document) []Violation {
	var violations []Violation

	if doc == nil || doc.Components == nil || doc.Components.Schemas == nil {
		return violations
	}

	for schemaName, schemaProxy := range doc.Components.Schemas.FromOldest() {
		schema := schemaProxy.Schema()
		if schema == nil || schema.Properties == nil {
			continue
		}

		for propName, propProxy := range schema.Properties.FromOldest() {
			propSchema := propProxy.Schema()
			if propSchema == nil || propSchema.Example == nil {
				continue
			}

			location := fmt.Sprintf("components/schemas/%s/%s", schemaName, propName)
			example := propSchema.Example

			if len(propSchema.Type) > 0 {
				v := r.validateExampleType(propSchema.Type[0], example, location)
				if v != nil {
					violations = append(violations, *v)
					continue
				}
			}

			if len(propSchema.Enum) > 0 {
				v := r.validateExampleEnum(propSchema.Enum, example, location)
				if v != nil {
					violations = append(violations, *v)
				}
			}
		}
	}

	return violations
}

func (r *SchemaExampleValidationRule) validateExampleType(schemaType string, example *yaml.Node, location string) *Violation {
	switch schemaType {
	case "string":
		if example.Tag != "!!str" {
			return &Violation{
				Suggestion: "Ensure the example value matches the schema type",
				Message:    "Example value does not match schema type 'string'",
				Location:   location,
				RuleName:   r.Name(),
				Severity:   SeverityError,
			}
		}
	case "integer":
		if example.Tag != "!!int" {
			return &Violation{
				Suggestion: "Ensure the example value matches the schema type",
				Message:    "Example value does not match schema type 'integer'",
				Location:   location,
				RuleName:   r.Name(),
				Severity:   SeverityError,
			}
		}
	case "number":
		if example.Tag != "!!int" && example.Tag != "!!float" {
			return &Violation{
				Suggestion: "Ensure the example value matches the schema type",
				Message:    "Example value does not match schema type 'number'",
				Location:   location,
				RuleName:   r.Name(),
				Severity:   SeverityError,
			}
		}
	case "boolean":
		if example.Tag != "!!bool" {
			return &Violation{
				Suggestion: "Ensure the example value matches the schema type",
				Message:    "Example value does not match schema type 'boolean'",
				Location:   location,
				RuleName:   r.Name(),
				Severity:   SeverityError,
			}
		}
	}
	return nil
}

func (r *SchemaExampleValidationRule) validateExampleEnum(enum []*yaml.Node, example *yaml.Node, location string) *Violation {
	values := make([]string, len(enum))
	for i, node := range enum {
		values[i] = node.Value
	}

	if !slices.Contains(values, example.Value) {
		return &Violation{
			Suggestion: "Ensure the example value is one of the allowed enum values",
			Message:    fmt.Sprintf("Example value '%s' is not one of the allowed enum values", example.Value),
			Location:   location,
			RuleName:   r.Name(),
			Severity:   SeverityError,
		}
	}
	return nil
}
