package rules

import (
	"fmt"

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

// check returns a violation when an integer enum declares an *_UNSPECIFIED sentinel
// that is not its first variant. Only type: integer enums become closed Protobuf
// enums where number 0 (the wire default for unset) must be owned by the sentinel;
// string enums generate open proto string fields and sentinel-less integer enums
// legitimately put a real value at 0, so neither is this rule's concern. This mirrors
// fieldmap.assertUnspecifiedFirst so lint rejects exactly what generate rejects.
func (r *EnumUnspecifiedVariantRule) check(types []string, enum []*yaml.Node, location string) *Violation {
	if len(enum) == 0 || !isIntegerEnum(types) {
		return nil
	}

	hasSentinel := false
	for _, node := range enum {
		if isUnspecifiedSentinel(node.Value) {
			hasSentinel = true
			break
		}
	}
	if !hasSentinel || isUnspecifiedSentinel(enum[0].Value) {
		return nil
	}

	return &Violation{
		Suggestion: "Move the *_UNSPECIFIED variant to the first position so it owns Protobuf number 0 (the wire default for unset)",
		Message:    fmt.Sprintf("Enum declares an *_UNSPECIFIED sentinel but its first variant is '%s'", enum[0].Value),
		Location:   location,
		RuleName:   r.Name(),
		Severity:   SeverityError,
	}
}
