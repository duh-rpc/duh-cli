package rules

import (
	"fmt"
	"regexp"

	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

var camelCaseRegex = regexp.MustCompile(`^[a-z][a-zA-Z0-9]*$`)

type PropertyCamelCaseRule struct{}

func NewPropertyCamelCaseRule() *PropertyCamelCaseRule {
	return &PropertyCamelCaseRule{}
}

func (r *PropertyCamelCaseRule) Name() string {
	return "PROPERTY_CAMELCASE"
}

func (r *PropertyCamelCaseRule) Validate(doc *v3.Document) []Violation {
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

		for propName := range schema.Properties.FromOldest() {
			if !camelCaseRegex.MatchString(propName) {
				violations = append(violations, Violation{
					Suggestion: "Rename property to camelCase (e.g., 'first_name' should be 'firstName')",
					Message:    fmt.Sprintf("Property name '%s' is not camelCase", propName),
					Location:   fmt.Sprintf("components/schemas/%s/%s", schemaName, propName),
					RuleName:   r.Name(),
					Severity:   SeverityError,
				})
			}
		}
	}

	return violations
}
