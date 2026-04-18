package rules

import (
	"fmt"

	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

type DiscriminatorRequiredRule struct{}

func NewDiscriminatorRequiredRule() *DiscriminatorRequiredRule {
	return &DiscriminatorRequiredRule{}
}

func (r *DiscriminatorRequiredRule) Name() string {
	return "DISCRIMINATOR_REQUIRED"
}

func (r *DiscriminatorRequiredRule) Validate(doc *v3.Document) []Violation {
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

		if len(schema.OneOf) > 0 && schema.Discriminator == nil {
			violations = append(violations, Violation{
				Suggestion: "Add a discriminator with propertyName: 'type' and a mapping for each variant",
				Message:    "Schema with oneOf must have a discriminator",
				Location:   fmt.Sprintf("components/schemas/%s", schemaName),
				RuleName:   r.Name(),
				Severity:   SeverityError,
			})
		}
	}

	return violations
}
