package rules

import (
	"fmt"

	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

type DescriptionRequiredRule struct{}

func NewDescriptionRequiredRule() *DescriptionRequiredRule {
	return &DescriptionRequiredRule{}
}

func (r *DescriptionRequiredRule) Name() string {
	return "DESCRIPTION_REQUIRED"
}

func (r *DescriptionRequiredRule) Validate(doc *v3.Document) []Violation {
	var violations []Violation

	if doc == nil {
		return violations
	}

	// Check operations and parameters
	if doc.Paths != nil && doc.Paths.PathItems != nil {
		for pathName, pathItem := range doc.Paths.PathItems.FromOldest() {
			if pathItem == nil {
				continue
			}

			// Check path-level parameters
			for _, param := range pathItem.Parameters {
				if param == nil {
					continue
				}
				if param.Description == "" {
					violations = append(violations, Violation{
						Suggestion: "Add a description to document this element",
						Message:    fmt.Sprintf("Parameter '%s' must have a description", param.Name),
						Location:   fmt.Sprintf("%s parameter %s", pathName, param.Name),
						RuleName:   r.Name(),
						Severity:   SeverityWarning,
					})
				}
			}

			operations := map[string]*v3.Operation{
				"POST":    pathItem.Post,
				"GET":     pathItem.Get,
				"PUT":     pathItem.Put,
				"DELETE":  pathItem.Delete,
				"PATCH":   pathItem.Patch,
				"HEAD":    pathItem.Head,
				"OPTIONS": pathItem.Options,
				"TRACE":   pathItem.Trace,
			}

			for method, operation := range operations {
				if operation == nil {
					continue
				}

				if operation.Description == "" {
					violations = append(violations, Violation{
						Suggestion: "Add a description to document this element",
						Message:    "Operation must have a description",
						Location:   fmt.Sprintf("%s %s", method, pathName),
						RuleName:   r.Name(),
						Severity:   SeverityWarning,
					})
				}

				for _, param := range operation.Parameters {
					if param == nil {
						continue
					}
					if param.Description == "" {
						violations = append(violations, Violation{
							Suggestion: "Add a description to document this element",
							Message:    fmt.Sprintf("Parameter '%s' must have a description", param.Name),
							Location:   fmt.Sprintf("%s %s parameter %s", method, pathName, param.Name),
							RuleName:   r.Name(),
							Severity:   SeverityWarning,
						})
					}
				}
			}
		}
	}

	// Check schema properties
	if doc.Components != nil && doc.Components.Schemas != nil {
		for schemaName, schemaProxy := range doc.Components.Schemas.FromOldest() {
			schema := schemaProxy.Schema()
			if schema == nil || schema.Properties == nil {
				continue
			}

			for propName, propProxy := range schema.Properties.FromOldest() {
				propSchema := propProxy.Schema()
				if propSchema == nil {
					continue
				}

				if propSchema.Description == "" {
					violations = append(violations, Violation{
						Suggestion: "Add a description to document this element",
						Message:    fmt.Sprintf("Property '%s' must have a description", propName),
						Location:   fmt.Sprintf("components/schemas/%s/%s", schemaName, propName),
						RuleName:   r.Name(),
						Severity:   SeverityWarning,
					})
				}
			}
		}
	}

	return violations
}
