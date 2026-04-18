package rules

import (
	"fmt"
	"strings"

	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

type TimestampFormatRule struct{}

func NewTimestampFormatRule() *TimestampFormatRule {
	return &TimestampFormatRule{}
}

func (r *TimestampFormatRule) Name() string {
	return "TIMESTAMP_FORMAT"
}

func (r *TimestampFormatRule) Validate(doc *v3.Document) []Violation {
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

		for propName, propProxy := range schema.Properties.FromOldest() {
			if !strings.HasSuffix(propName, "At") && !strings.HasSuffix(propName, "Timestamp") {
				continue
			}

			propSchema := propProxy.Schema()
			if propSchema == nil {
				continue
			}

			if len(propSchema.Type) == 0 || propSchema.Type[0] != "string" || propSchema.Format != "date-time" {
				violations = append(violations, Violation{
					Suggestion: "Set type to 'string' and format to 'date-time' for timestamp fields",
					Message:    fmt.Sprintf("Field '%s' ending in 'At'/'Timestamp' must be type string with format: date-time", propName),
					Location:   fmt.Sprintf("components/schemas/%s/%s", schemaName, propName),
					RuleName:   r.Name(),
					Severity:   SeverityError,
				})
			}
		}
	}

	return violations
}
