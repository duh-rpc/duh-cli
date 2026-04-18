package rules

import (
	"fmt"

	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

type DiscriminatorPropertyNameRule struct{}

func NewDiscriminatorPropertyNameRule() *DiscriminatorPropertyNameRule {
	return &DiscriminatorPropertyNameRule{}
}

func (r *DiscriminatorPropertyNameRule) Name() string {
	return "DISCRIMINATOR_PROPERTY_NAME"
}

func (r *DiscriminatorPropertyNameRule) Validate(doc *v3.Document) []Violation {
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

		if len(schema.OneOf) > 0 && schema.Discriminator != nil {
			if schema.Discriminator.PropertyName != "type" {
				violations = append(violations, Violation{
					Suggestion: "Set discriminator propertyName to 'type'",
					Message:    fmt.Sprintf("Discriminator property must be named 'type' (found: '%s')", schema.Discriminator.PropertyName),
					Location:   fmt.Sprintf("components/schemas/%s", schemaName),
					RuleName:   r.Name(),
					Severity:   SeverityError,
				})
			}
		}
	}

	return violations
}
