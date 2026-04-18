package rules

import (
	"fmt"

	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

type AmountDecimalStringRule struct{}

func NewAmountDecimalStringRule() *AmountDecimalStringRule {
	return &AmountDecimalStringRule{}
}

func (r *AmountDecimalStringRule) Name() string {
	return "AMOUNT_DECIMAL_STRING"
}

func (r *AmountDecimalStringRule) Validate(doc *v3.Document) []Violation {
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
			if propName != "amount" {
				continue
			}

			propSchema := propProxy.Schema()
			if propSchema == nil {
				continue
			}

			if len(propSchema.Type) == 0 || propSchema.Type[0] != "string" {
				violations = append(violations, Violation{
					Suggestion: "Use type: string for amount fields to avoid floating-point precision issues",
					Message:    "Property 'amount' must be type string for precise decimal representation",
					Location:   fmt.Sprintf("components/schemas/%s/%s", schemaName, propName),
					RuleName:   r.Name(),
					Severity:   SeverityError,
				})
			}
		}
	}

	return violations
}
