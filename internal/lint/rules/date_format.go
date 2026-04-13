package rules

import (
	"fmt"
	"strings"

	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

type DateFormatRule struct{}

func NewDateFormatRule() *DateFormatRule {
	return &DateFormatRule{}
}

func (r *DateFormatRule) Name() string {
	return "DATE_FORMAT"
}

func (r *DateFormatRule) Validate(doc *v3.Document) []Violation {
	var violations []Violation

	if doc == nil || doc.Components == nil || doc.Components.Schemas == nil {
		return violations
	}

	for schemaName, schemaProxy := range doc.Components.Schemas.FromOldest() {
		schema := schemaProxy.Schema()
		if schema == nil || schema.Properties == nil {
			continue
		}

		for propName, propProxy := range schema.Properties.FromOldest() {
			if !strings.HasSuffix(propName, "Date") {
				continue
			}

			// Safety guard: skip if also matches timestamp suffixes
			if strings.HasSuffix(propName, "At") || strings.HasSuffix(propName, "Timestamp") {
				continue
			}

			propSchema := propProxy.Schema()
			if propSchema == nil {
				continue
			}

			if len(propSchema.Type) == 0 || propSchema.Type[0] != "string" || propSchema.Format != "date" {
				violations = append(violations, Violation{
					Suggestion: "Set type to 'string' and format to 'date' for date fields",
					Message:    fmt.Sprintf("Field '%s' ending in 'Date' must be type string with format: date", propName),
					Location:   fmt.Sprintf("components/schemas/%s/%s", schemaName, propName),
					RuleName:   r.Name(),
					Severity:   SeverityError,
				})
			}
		}
	}

	return violations
}
