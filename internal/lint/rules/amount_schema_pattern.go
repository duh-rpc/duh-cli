package rules

import (
	"fmt"

	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

type AmountSchemaPatternRule struct{}

func NewAmountSchemaPatternRule() *AmountSchemaPatternRule {
	return &AmountSchemaPatternRule{}
}

func (r *AmountSchemaPatternRule) Name() string {
	return "AMOUNT_SCHEMA_PATTERN"
}

func (r *AmountSchemaPatternRule) Validate(doc *v3.Document) []Violation {
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

		if _, hasAmount := schema.Properties.Get("amount"); hasAmount {
			if _, hasAssetType := schema.Properties.Get("assetType"); !hasAssetType {
				violations = append(violations, Violation{
					Suggestion: "Add an 'assetType' property to schemas that contain 'amount' for currency/asset clarity",
					Message:    "Schema has 'amount' property but missing 'assetType' property",
					Location:   fmt.Sprintf("components/schemas/%s", schemaName),
					RuleName:   r.Name(),
					Severity:   SeverityWarning,
				})
			}
		}
	}

	return violations
}
