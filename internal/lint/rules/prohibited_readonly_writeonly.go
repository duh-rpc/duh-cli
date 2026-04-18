package rules

import (
	"fmt"

	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

type ProhibitedReadOnlyWriteOnlyRule struct{}

func NewProhibitedReadOnlyWriteOnlyRule() *ProhibitedReadOnlyWriteOnlyRule {
	return &ProhibitedReadOnlyWriteOnlyRule{}
}

func (r *ProhibitedReadOnlyWriteOnlyRule) Name() string {
	return "PROHIBITED_READONLY_WRITEONLY"
}

func (r *ProhibitedReadOnlyWriteOnlyRule) Validate(doc *v3.Document) []Violation {
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

			if propSchema.ReadOnly != nil && *propSchema.ReadOnly {
				violations = append(violations, Violation{
					Suggestion: "Remove readOnly/writeOnly and define separate Request and Response schema types",
					Message:    "Property uses readOnly which is not allowed; use separate Request/Response schemas instead",
					Location:   fmt.Sprintf("components/schemas/%s/%s", schemaName, propName),
					RuleName:   r.Name(),
					Severity:   SeverityError,
				})
			}

			if propSchema.WriteOnly != nil && *propSchema.WriteOnly {
				violations = append(violations, Violation{
					Suggestion: "Remove readOnly/writeOnly and define separate Request and Response schema types",
					Message:    "Property uses writeOnly which is not allowed; use separate Request/Response schemas instead",
					Location:   fmt.Sprintf("components/schemas/%s/%s", schemaName, propName),
					RuleName:   r.Name(),
					Severity:   SeverityError,
				})
			}
		}
	}

	return violations
}
