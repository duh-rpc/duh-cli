package rules

import (
	"fmt"

	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

type ProhibitedAllOfUnionRule struct{}

func NewProhibitedAllOfUnionRule() *ProhibitedAllOfUnionRule {
	return &ProhibitedAllOfUnionRule{}
}

func (r *ProhibitedAllOfUnionRule) Name() string {
	return "PROHIBITED_ALLOF"
}

func (r *ProhibitedAllOfUnionRule) Validate(doc *v3.Document) []Violation {
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

		if len(schema.AllOf) > 0 {
			violations = append(violations, Violation{
				Suggestion: "Use separate optional properties or a discriminated oneOf pattern instead",
				Message:    "Schema uses allOf for type unions which is not allowed",
				Location:   fmt.Sprintf("components/schemas/%s", schemaName),
				RuleName:   r.Name(),
				Severity:   SeverityError,
			})
		}
	}

	return violations
}
