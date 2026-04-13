package rules

import (
	"fmt"

	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

type ProhibitedAnyOfRule struct{}

func NewProhibitedAnyOfRule() *ProhibitedAnyOfRule {
	return &ProhibitedAnyOfRule{}
}

func (r *ProhibitedAnyOfRule) Name() string {
	return "PROHIBITED_ANYOF"
}

func (r *ProhibitedAnyOfRule) Validate(doc *v3.Document) []Violation {
	var violations []Violation

	if doc == nil || doc.Components == nil || doc.Components.Schemas == nil {
		return violations
	}

	for schemaName, schemaProxy := range doc.Components.Schemas.FromOldest() {
		schema := schemaProxy.Schema()
		if schema == nil {
			continue
		}

		if len(schema.AnyOf) > 0 {
			violations = append(violations, Violation{
				Suggestion: "Replace anyOf with a discriminated oneOf using a 'type' discriminator property",
				Message:    "Schema uses anyOf which is not allowed; use discriminated oneOf instead",
				Location:   fmt.Sprintf("components/schemas/%s", schemaName),
				RuleName:   r.Name(),
				Severity:   SeverityError,
			})
		}
	}

	return violations
}
