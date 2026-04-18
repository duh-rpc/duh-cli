package rules

import (
	"fmt"

	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

type RPCNoNestedArraysRule struct{}

func NewRPCNoNestedArraysRule() *RPCNoNestedArraysRule {
	return &RPCNoNestedArraysRule{}
}

func (r *RPCNoNestedArraysRule) Name() string {
	return "NO_NESTED_ARRAYS"
}

func (r *RPCNoNestedArraysRule) Validate(doc *v3.Document) []Violation {
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
			if propSchema == nil || len(propSchema.Type) == 0 {
				continue
			}

			if propSchema.Type[0] != "array" {
				continue
			}

			if propSchema.Items == nil || !propSchema.Items.IsA() {
				continue
			}

			itemsSchema := propSchema.Items.A.Schema()
			if itemsSchema == nil || len(itemsSchema.Type) == 0 {
				continue
			}

			if itemsSchema.Type[0] == "array" {
				violations = append(violations, Violation{
					Suggestion: "Wrap the inner array in a named schema object",
					Message:    "Array property contains nested array (array of arrays); proto3 has no repeated repeated type",
					Location:   fmt.Sprintf("components/schemas/%s/%s", schemaName, propName),
					RuleName:   r.Name(),
					Severity:   SeverityError,
				})
			}
		}
	}

	return violations
}
