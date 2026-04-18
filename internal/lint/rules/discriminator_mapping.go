package rules

import (
	"fmt"

	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

type DiscriminatorMappingRule struct{}

func NewDiscriminatorMappingRule() *DiscriminatorMappingRule {
	return &DiscriminatorMappingRule{}
}

func (r *DiscriminatorMappingRule) Name() string {
	return "DISCRIMINATOR_MAPPING"
}

func (r *DiscriminatorMappingRule) Validate(doc *v3.Document) []Violation {
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
			if schema.Discriminator.Mapping == nil || schema.Discriminator.Mapping.Len() == 0 {
				violations = append(violations, Violation{
					Suggestion: "Add a mapping to the discriminator that maps each variant value to its schema reference",
					Message:    "Discriminator must include a mapping",
					Location:   fmt.Sprintf("components/schemas/%s", schemaName),
					RuleName:   r.Name(),
					Severity:   SeverityError,
				})
			}
		}
	}

	return violations
}
