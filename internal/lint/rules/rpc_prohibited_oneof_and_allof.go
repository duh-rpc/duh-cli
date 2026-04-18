package rules

import (
	"fmt"

	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

type RPCProhibitedOneOfAndAllOfRule struct{}

func NewRPCProhibitedOneOfAndAllOfRule() *RPCProhibitedOneOfAndAllOfRule {
	return &RPCProhibitedOneOfAndAllOfRule{}
}

func (r *RPCProhibitedOneOfAndAllOfRule) Name() string {
	return "PROHIBITED_ONEOF"
}

func (r *RPCProhibitedOneOfAndAllOfRule) Validate(doc *v3.Document) []Violation {
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

		location := fmt.Sprintf("components/schemas/%s", schemaName)

		if len(schema.OneOf) > 0 {
			violations = append(violations, Violation{
				Suggestion: "Replace with a flat object using optional properties that map to proto3 optional message fields",
				Message:    "Schema uses oneOf which cannot be mapped to proto3; use plain optional properties instead",
				Location:   location,
				RuleName:   r.Name(),
				Severity:   SeverityError,
			})
		}
	}

	return violations
}
