package rules

import (
	"fmt"

	"github.com/pb33f/libopenapi/datamodel/high/v3"
	"go.yaml.in/yaml/v4"
)

type EnumStringSentinelNamesRule struct{}

func NewEnumStringSentinelNamesRule() *EnumStringSentinelNamesRule {
	return &EnumStringSentinelNamesRule{}
}

func (r *EnumStringSentinelNamesRule) Name() string {
	return "ENUM_STRING_SENTINEL_NAMES"
}

func (r *EnumStringSentinelNamesRule) Validate(doc *v3.Document) []Violation {
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

		if v := r.check(schema.Type, schema.Enum, fmt.Sprintf("components/schemas/%s", schemaName)); v != nil {
			violations = append(violations, *v)
		}

		if schema.Properties == nil {
			continue
		}

		// Inline enums defined on a property. References are skipped; the
		// referenced schema is validated where it is defined.
		for propName, propProxy := range schema.Properties.FromOldest() {
			if propProxy.IsReference() {
				continue
			}
			propSchema := propProxy.Schema()
			if propSchema == nil {
				continue
			}
			if v := r.check(propSchema.Type, propSchema.Enum, fmt.Sprintf("components/schemas/%s/%s", schemaName, propName)); v != nil {
				violations = append(violations, *v)
			}
		}
	}

	return violations
}

// check returns a warning when a type: string enum carries a value with proto-enum
// sentinel naming (ends in UNSPECIFIED). A string enum generates an open proto string
// field, so the sentinel buys nothing — the author almost certainly meant a closed
// type: integer enum. This is advisory only: the spec generates correctly, so the rule
// must never block generation. It is disjoint from ENUM_UNSPECIFIED_VARIANT, which
// requires type: integer.
func (r *EnumStringSentinelNamesRule) check(types []string, enum []*yaml.Node, location string) *Violation {
	if len(enum) == 0 || !isStringEnum(types) {
		return nil
	}

	for _, node := range enum {
		if isUnspecifiedSentinel(node.Value) {
			return &Violation{
				Suggestion: "For a closed, wire-safe proto enum use type: integer with named variants; to keep an open string field, drop the *_UNSPECIFIED sentinel",
				Message:    fmt.Sprintf("String enum variant '%s' uses proto-enum sentinel naming, but a string enum generates an open proto string field", node.Value),
				Location:   location,
				RuleName:   r.Name(),
				Severity:   SeverityWarning,
			}
		}
	}

	return nil
}
