package rules

import (
	"fmt"

	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

type DiscriminatorVariantFieldRule struct{}

func NewDiscriminatorVariantFieldRule() *DiscriminatorVariantFieldRule {
	return &DiscriminatorVariantFieldRule{}
}

func (r *DiscriminatorVariantFieldRule) Name() string {
	return "DISCRIMINATOR_VARIANT_FIELD"
}

func (r *DiscriminatorVariantFieldRule) Validate(doc *v3.Document) []Violation {
	var violations []Violation

	if doc == nil || doc.Components == nil || doc.Components.Schemas == nil {
		return violations
	}

	for schemaName, schemaProxy := range doc.Components.Schemas.FromOldest() {
		schema := schemaProxy.Schema()
		if schema == nil {
			continue
		}

		if len(schema.OneOf) == 0 || schema.Discriminator == nil {
			continue
		}

		discProp := schema.Discriminator.PropertyName

		for _, variant := range schema.OneOf {
			variantSchema := variant.Schema()
			if variantSchema == nil || variantSchema.Properties == nil {
				violations = append(violations, Violation{
					Suggestion: fmt.Sprintf("Add a '%s' property to each oneOf variant schema", discProp),
					Message:    fmt.Sprintf("OneOf variant is missing discriminator field '%s'", discProp),
					Location:   fmt.Sprintf("components/schemas/%s", schemaName),
					RuleName:   r.Name(),
					Severity:   SeverityError,
				})
				continue
			}

			if _, ok := variantSchema.Properties.Get(discProp); !ok {
				violations = append(violations, Violation{
					Suggestion: fmt.Sprintf("Add a '%s' property to each oneOf variant schema", discProp),
					Message:    fmt.Sprintf("OneOf variant is missing discriminator field '%s'", discProp),
					Location:   fmt.Sprintf("components/schemas/%s", schemaName),
					RuleName:   r.Name(),
					Severity:   SeverityError,
				})
			}
		}
	}

	return violations
}
