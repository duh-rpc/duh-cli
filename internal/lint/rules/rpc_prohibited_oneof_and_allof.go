package rules

import (
	"fmt"
	"slices"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

type RPCProhibitedOneOfAndAllOfRule struct{}

func NewRPCProhibitedOneOfAndAllOfRule() *RPCProhibitedOneOfAndAllOfRule {
	return &RPCProhibitedOneOfAndAllOfRule{}
}

func (r *RPCProhibitedOneOfAndAllOfRule) Name() string {
	return "PROHIBITED_ONEOF"
}

// Validate prohibits the flat/discriminated oneOf — the form that cannot be served
// on both the JSON and protobuf wires — while allowing the wire-compatible "style B"
// nested union (one optional $ref property per variant, a oneOf of single-required
// branches, no discriminator), which maps to a protobuf oneof. A style-B-shaped
// schema that is malformed is reported with a specific message so the author can fix
// it. The classification mirrors openapi-schema.go's isStyleBOneOf so lint and
// generate agree.
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

		if len(schema.OneOf) == 0 {
			continue
		}

		location := fmt.Sprintf("components/schemas/%s", schemaName)

		if !isStyleBOneOf(schema) {
			// Flat/discriminated form: a discriminator, or $ref/inline variant
			// schemas. Cannot be expressed as protobuf on both wires.
			violations = append(violations, Violation{
				Suggestion: "Use a nested union: one optional $ref property per variant plus a oneOf of single-required branches and no discriminator, or a flat object with optional properties",
				Message:    "Schema uses a flat/discriminated oneOf which cannot be served on both the JSON and protobuf wires",
				Location:   location,
				RuleName:   r.Name(),
				Severity:   SeverityError,
			})
			continue
		}

		// Style-B-shaped: enforce the precise contract so a malformed union is
		// rejected here rather than failing downstream in generate/protoc.
		if msg := styleBOneOfViolation(schema); msg != "" {
			violations = append(violations, Violation{
				Suggestion: "Give each oneOf branch exactly one required entry naming a declared, non-array $ref property",
				Message:    msg,
				Location:   location,
				RuleName:   r.Name(),
				Severity:   SeverityError,
			})
		}
	}

	return violations
}

// isStyleBOneOf reports whether a oneOf schema is the wire-compatible style-B form:
// no discriminator, and every oneOf branch is a constraint object (not a $ref/inline
// variant schema) carrying a required list. This is a routing predicate only;
// styleBOneOfViolation enforces the precise shape.
func isStyleBOneOf(schema *base.Schema) bool {
	if schema.Discriminator != nil && schema.Discriminator.PropertyName != "" {
		return false
	}
	for _, branch := range schema.OneOf {
		if branch.IsReference() {
			return false
		}
		bs := branch.Schema()
		if bs == nil || len(bs.Required) == 0 {
			return false
		}
	}
	return true
}

// styleBOneOfViolation returns a specific message when a style-B-shaped schema is
// malformed, or "" when it is valid. The checks match openapi-schema.go's
// validateStyleBOneOf: each branch names exactly one required property, that property
// is declared, and it is not an array (proto3 forbids repeated inside a oneof).
func styleBOneOfViolation(schema *base.Schema) string {
	for _, branch := range schema.OneOf {
		bs := branch.Schema()
		if bs == nil {
			return "oneOf branch could not be resolved"
		}
		if len(bs.Required) != 1 {
			return fmt.Sprintf("oneOf branch must name exactly one required property, got %d", len(bs.Required))
		}

		propName := bs.Required[0]
		if schema.Properties == nil {
			return fmt.Sprintf("oneOf branch requires property '%s' but the schema declares no properties", propName)
		}
		propProxy := schema.Properties.GetOrZero(propName)
		if propProxy == nil {
			return fmt.Sprintf("oneOf branch requires property '%s' which is not declared in properties", propName)
		}

		propSchema := propProxy.Schema()
		if propSchema != nil && slices.Contains(propSchema.Type, "array") {
			return fmt.Sprintf("oneOf variant property '%s' is an array; repeated fields are not allowed in a protobuf oneof", propName)
		}
	}
	return ""
}
