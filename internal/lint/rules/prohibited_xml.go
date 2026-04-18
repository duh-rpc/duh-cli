package rules

import (
	"fmt"

	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

type ProhibitedXMLRule struct{}

func NewProhibitedXMLRule() *ProhibitedXMLRule {
	return &ProhibitedXMLRule{}
}

func (r *ProhibitedXMLRule) Name() string {
	return "PROHIBITED_XML"
}

func (r *ProhibitedXMLRule) Validate(doc *v3.Document) []Violation {
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

		if schema.XML != nil {
			violations = append(violations, Violation{
				Suggestion: "Remove the xml property; use application/json or application/protobuf",
				Message:    "XML property is not allowed on schemas",
				Location:   fmt.Sprintf("components/schemas/%s", schemaName),
				RuleName:   r.Name(),
				Severity:   SeverityError,
			})
		}

		if schema.Properties == nil {
			continue
		}

		for propName, propProxy := range schema.Properties.FromOldest() {
			propSchema := propProxy.Schema()
			if propSchema == nil {
				continue
			}

			if propSchema.XML != nil {
				violations = append(violations, Violation{
					Suggestion: "Remove the xml property; use application/json or application/protobuf",
					Message:    "XML property is not allowed on schemas",
					Location:   fmt.Sprintf("components/schemas/%s/%s", schemaName, propName),
					RuleName:   r.Name(),
					Severity:   SeverityError,
				})
			}
		}
	}

	return violations
}
