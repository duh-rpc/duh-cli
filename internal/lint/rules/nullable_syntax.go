package rules

import (
	"fmt"
	"slices"

	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

type NullableSyntaxRule struct{}

func NewNullableSyntaxRule() *NullableSyntaxRule {
	return &NullableSyntaxRule{}
}

func (r *NullableSyntaxRule) Name() string {
	return "NULLABLE_SYNTAX"
}

func (r *NullableSyntaxRule) Validate(doc *v3.Document) []Violation {
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
			if propSchema == nil {
				continue
			}

			if slices.Contains(propSchema.Type, "null") {
				violations = append(violations, Violation{
					Suggestion: "Replace type: [string, null] with type: string and nullable: true",
					Message:    "Property uses type array with null; use nullable: true instead",
					Location:   fmt.Sprintf("components/schemas/%s/%s", schemaName, propName),
					RuleName:   r.Name(),
					Severity:   SeverityError,
				})
			}
		}
	}

	return violations
}
