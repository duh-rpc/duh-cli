package rules

import (
	"fmt"
	"strings"

	"github.com/pb33f/libopenapi/datamodel/high/v3"
	"go.yaml.in/yaml/v4"
)

type EnumUnspecifiedVariantRule struct{}

func NewEnumUnspecifiedVariantRule() *EnumUnspecifiedVariantRule {
	return &EnumUnspecifiedVariantRule{}
}

func (r *EnumUnspecifiedVariantRule) Name() string {
	return "ENUM_UNSPECIFIED_VARIANT"
}

func (r *EnumUnspecifiedVariantRule) Validate(doc *v3.Document) []Violation {
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

		if v := r.check(schema.Enum, fmt.Sprintf("components/schemas/%s", schemaName)); v != nil {
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
			if v := r.check(propSchema.Enum, fmt.Sprintf("components/schemas/%s/%s", schemaName, propName)); v != nil {
				violations = append(violations, *v)
			}
		}
	}

	return violations
}

// check returns a violation when an enum does not declare an UNSPECIFIED variant
// as its first entry. The zero value is the Protobuf wire default, so reserving it
// for UNSPECIFIED keeps "unset" distinguishable from a real value. An empty list is
// a free-form string, not an enum, and is not this rule's concern.
func (r *EnumUnspecifiedVariantRule) check(enum []*yaml.Node, location string) *Violation {
	if len(enum) == 0 {
		return nil
	}

	if strings.HasSuffix(enum[0].Value, "UNSPECIFIED") {
		return nil
	}

	return &Violation{
		Suggestion: "Declare an UNSPECIFIED variant (e.g. STATUS_UNSPECIFIED) as the first enum entry so Protobuf field number 0 represents an unset value",
		Message:    fmt.Sprintf("Enum's first variant '%s' is not an UNSPECIFIED variant", enum[0].Value),
		Location:   location,
		RuleName:   r.Name(),
		Severity:   SeverityError,
	}
}
