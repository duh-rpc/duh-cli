package fieldmap

import (
	"strings"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// isUnspecifiedSentinel reports whether an enum variant name is the *_UNSPECIFIED
// sentinel that must own proto number 0 (the wire default; see ADR 0004).
func isUnspecifiedSentinel(name string) bool {
	return strings.HasSuffix(name, "UNSPECIFIED")
}

// messageSpec is a top-level object component schema and its field JSON names in
// OpenAPI declaration order.
type messageSpec struct {
	Name   string
	Fields []string
}

// enumSpec is a top-level proto-enum component schema (an integer enum) and its
// literal variant values in declaration order. String enums are intentionally not
// proto enums in DUH (they render as proto string fields to keep the JSON wire
// identical), so they are excluded here and their containing fields are tracked as
// ordinary message fields.
type enumSpec struct {
	Name     string
	Variants []string
}

// specUnits classifies the spec's top-level component schemas into the lockable
// units, mirroring how the conversion library treats each schema:
//   - object schemas → proto messages (fields keyed by JSON name)
//   - integer enums  → proto enums (variants keyed by literal value)
//   - string enums   → proto string fields (no lock unit of their own)
func specUnits(doc *v3.Document) (messages []messageSpec, enums []enumSpec) {
	if doc == nil || doc.Components == nil || doc.Components.Schemas == nil {
		return nil, nil
	}

	for name, proxy := range doc.Components.Schemas.FromOldest() {
		schema := proxy.Schema()
		if schema == nil {
			continue
		}

		if len(schema.Enum) > 0 {
			if hasType(schema, "string") {
				continue // string enum → proto string field, not a lock unit
			}
			enums = append(enums, enumSpec{Name: name, Variants: enumValues(schema)})
			continue
		}

		if hasType(schema, "object") || schema.Properties != nil {
			messages = append(messages, messageSpec{Name: name, Fields: fieldNames(schema)})
		}
	}

	return messages, enums
}

func hasType(schema *base.Schema, want string) bool {
	for _, t := range schema.Type {
		if t == want {
			return true
		}
	}
	return false
}

func fieldNames(schema *base.Schema) []string {
	if schema.Properties == nil {
		return nil
	}
	var names []string
	for name := range schema.Properties.FromOldest() {
		names = append(names, name)
	}
	return names
}

func enumValues(schema *base.Schema) []string {
	var values []string
	for _, node := range schema.Enum {
		if node != nil {
			values = append(values, node.Value)
		}
	}
	return values
}
