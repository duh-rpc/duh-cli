package rules

import (
	"fmt"
	"slices"

	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

type RPCNoNullableRule struct{}

func NewRPCNoNullableRule() *RPCNoNullableRule {
	return &RPCNoNullableRule{}
}

func (r *RPCNoNullableRule) Name() string {
	return "RPC_NO_NULLABLE"
}

func (r *RPCNoNullableRule) Validate(doc *v3.Document) []Violation {
	var violations []Violation

	if doc == nil || doc.Components == nil || doc.Components.Schemas == nil {
		return violations
	}

	for schemaName, schemaProxy := range doc.Components.Schemas.FromOldest() {
		schema := schemaProxy.Schema()
		if schema == nil || schema.Properties == nil {
			continue
		}

		if isSchemaIgnored(schema, r.Name()) {
			continue
		}

		for propName, propProxy := range schema.Properties.FromOldest() {
			propSchema := propProxy.Schema()
			if propSchema == nil {
				continue
			}

			if propSchema.Nullable != nil && *propSchema.Nullable {
				violations = append(violations, Violation{
					Suggestion: "Remove nullable: true from the property definition",
					Message:    "Property must not use nullable: true; proto3 uses zero values and field absence instead",
					Location:   fmt.Sprintf("components/schemas/%s/%s", schemaName, propName),
					RuleName:   r.Name(),
					Severity:   SeverityError,
				})
			}

			if slices.Contains(propSchema.Type, "null") {
				violations = append(violations, Violation{
					Suggestion: "Replace type: [string, null] with type: string and nullable: true",
					Message:    "Property must not use type array with null",
					Location:   fmt.Sprintf("components/schemas/%s/%s", schemaName, propName),
					RuleName:   r.Name(),
					Severity:   SeverityError,
				})
			}
		}
	}

	return violations
}
