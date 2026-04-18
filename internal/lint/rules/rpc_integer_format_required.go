package rules

import (
	"fmt"

	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

type RPCIntegerFormatRequiredRule struct{}

func NewRPCIntegerFormatRequiredRule() *RPCIntegerFormatRequiredRule {
	return &RPCIntegerFormatRequiredRule{}
}

func (r *RPCIntegerFormatRequiredRule) Name() string {
	return "INTEGER_FORMAT_REQUIRED"
}

func (r *RPCIntegerFormatRequiredRule) Validate(doc *v3.Document) []Violation {
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
			propSchema := propProxy.Schema()
			if propSchema == nil || len(propSchema.Type) == 0 {
				continue
			}

			typeName := propSchema.Type[0]
			location := fmt.Sprintf("components/schemas/%s/%s", schemaName, propName)

			switch typeName {
			case "integer":
				if propSchema.Format == "" {
					violations = append(violations, Violation{
						Suggestion: "Add format: int32 or format: int64 to integer fields for unambiguous proto3 mapping",
						Message:    "Integer field must specify format (int32 or int64)",
						Location:   location,
						RuleName:   r.Name(),
						Severity:   SeverityError,
					})
				} else if propSchema.Format != "int32" && propSchema.Format != "int64" {
					violations = append(violations, Violation{
						Suggestion: "Add format: int32 or format: int64 to integer fields for unambiguous proto3 mapping",
						Message:    fmt.Sprintf("Integer field has invalid format '%s'; must be int32 or int64", propSchema.Format),
						Location:   location,
						RuleName:   r.Name(),
						Severity:   SeverityError,
					})
				}
			case "number":
				if propSchema.Format == "" {
					violations = append(violations, Violation{
						Suggestion: "Add format: float or format: double to number fields for unambiguous proto3 mapping",
						Message:    "Number field must specify format (float or double)",
						Location:   location,
						RuleName:   r.Name(),
						Severity:   SeverityError,
					})
				} else if propSchema.Format != "float" && propSchema.Format != "double" {
					violations = append(violations, Violation{
						Suggestion: "Add format: float or format: double to number fields for unambiguous proto3 mapping",
						Message:    fmt.Sprintf("Number field has invalid format '%s'; must be float or double", propSchema.Format),
						Location:   location,
						RuleName:   r.Name(),
						Severity:   SeverityError,
					})
				}
			}
		}
	}

	return violations
}
