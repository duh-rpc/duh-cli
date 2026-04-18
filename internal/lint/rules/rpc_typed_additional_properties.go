package rules

import (
	"fmt"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

type RPCTypedAdditionalPropertiesRule struct{}

func NewRPCTypedAdditionalPropertiesRule() *RPCTypedAdditionalPropertiesRule {
	return &RPCTypedAdditionalPropertiesRule{}
}

func (r *RPCTypedAdditionalPropertiesRule) Name() string {
	return "TYPED_ADDITIONAL_PROPERTIES"
}

func (r *RPCTypedAdditionalPropertiesRule) Validate(doc *v3.Document) []Violation {
	var violations []Violation

	if doc == nil || doc.Components == nil || doc.Components.Schemas == nil {
		return violations
	}

	for schemaName, schemaProxy := range doc.Components.Schemas.FromOldest() {
		schema := schemaProxy.Schema()
		if schema == nil {
			continue
		}

		if isSchemaIgnored(schema, r.Name()) {
			continue
		}

		if v := r.checkAdditionalProperties(schema, fmt.Sprintf("components/schemas/%s", schemaName)); v != nil {
			violations = append(violations, *v)
		}

		if schema.Properties == nil {
			continue
		}

		for propName, propProxy := range schema.Properties.FromOldest() {
			propSchema := propProxy.Schema()
			if propSchema == nil {
				continue
			}

			if v := r.checkAdditionalProperties(propSchema, fmt.Sprintf("components/schemas/%s/%s", schemaName, propName)); v != nil {
				violations = append(violations, *v)
			}
		}
	}

	return violations
}

func (r *RPCTypedAdditionalPropertiesRule) checkAdditionalProperties(schema *base.Schema, location string) *Violation {
	if schema.AdditionalProperties == nil {
		return nil
	}

	// additionalProperties: true — untyped map
	if schema.AdditionalProperties.IsB() && schema.AdditionalProperties.B {
		return &Violation{
			Suggestion: "Specify a type for additionalProperties (e.g., additionalProperties: { type: string })",
			Message:    "additionalProperties must have an explicit type for proto3 type safety",
			Location:   location,
			RuleName:   r.Name(),
			Severity:   SeverityError,
		}
	}

	// additionalProperties: {} — empty schema with no explicit type
	if schema.AdditionalProperties.IsA() {
		addlSchema := schema.AdditionalProperties.A.Schema()
		if addlSchema == nil || len(addlSchema.Type) == 0 {
			return &Violation{
				Suggestion: "Specify a type for additionalProperties (e.g., additionalProperties: { type: string })",
				Message:    "additionalProperties must have an explicit type for proto3 type safety",
				Location:   location,
				RuleName:   r.Name(),
				Severity:   SeverityError,
			}
		}
	}

	return nil
}
