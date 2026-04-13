package rules

import (
	"fmt"

	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

type IdempotencyKeyDefinitionRule struct{}

func NewIdempotencyKeyDefinitionRule() *IdempotencyKeyDefinitionRule {
	return &IdempotencyKeyDefinitionRule{}
}

func (r *IdempotencyKeyDefinitionRule) Name() string {
	return "IDEMPOTENCY_KEY_DEFINITION"
}

func (r *IdempotencyKeyDefinitionRule) Validate(doc *v3.Document) []Violation {
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
			if propName != "idempotencyKey" {
				continue
			}

			propSchema := propProxy.Schema()
			if propSchema == nil {
				continue
			}

			location := fmt.Sprintf("components/schemas/%s/idempotencyKey", schemaName)

			if len(propSchema.Type) == 0 || propSchema.Type[0] != "string" {
				violations = append(violations, Violation{
					Suggestion: "Define idempotencyKey as type: string with maxLength: 128",
					Message:    "Property 'idempotencyKey' must be type string",
					Location:   location,
					RuleName:   r.Name(),
					Severity:   SeverityError,
				})
			}

			if propSchema.MaxLength == nil || *propSchema.MaxLength != 128 {
				violations = append(violations, Violation{
					Suggestion: "Define idempotencyKey as type: string with maxLength: 128",
					Message:    "Property 'idempotencyKey' must have maxLength: 128",
					Location:   location,
					RuleName:   r.Name(),
					Severity:   SeverityError,
				})
			}
		}
	}

	return violations
}
