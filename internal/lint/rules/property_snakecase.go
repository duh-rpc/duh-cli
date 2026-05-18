package rules

import (
	"fmt"
	"regexp"

	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

var snakeCaseRegex = regexp.MustCompile(`^[a-z][a-z0-9]*(_[a-z0-9]+)*$`)

type PropertySnakeCaseRule struct{}

func NewPropertySnakeCaseRule() *PropertySnakeCaseRule {
	return &PropertySnakeCaseRule{}
}

func (r *PropertySnakeCaseRule) Name() string {
	return "PROPERTY_SNAKECASE"
}

func (r *PropertySnakeCaseRule) Validate(doc *v3.Document) []Violation {
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
			if !snakeCaseRegex.MatchString(propName) {
				violations = append(violations, Violation{
					Suggestion: "Rename property to snake_case (e.g., 'firstName' should be 'first_name')",
					Message:    fmt.Sprintf("Property name '%s' is not snake_case", propName),
					Location:   fmt.Sprintf("components/schemas/%s/%s", schemaName, propName),
					RuleName:   r.Name(),
					Severity:   SeverityError,
				})
			}
		}
	}

	return violations
}
