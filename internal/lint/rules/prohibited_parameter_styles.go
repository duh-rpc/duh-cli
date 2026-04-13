package rules

import (
	"fmt"

	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

type ProhibitedParameterStylesRule struct{}

func NewProhibitedParameterStylesRule() *ProhibitedParameterStylesRule {
	return &ProhibitedParameterStylesRule{}
}

func (r *ProhibitedParameterStylesRule) Name() string {
	return "PROHIBITED_PARAMETER_STYLES"
}

func (r *ProhibitedParameterStylesRule) Validate(doc *v3.Document) []Violation {
	var violations []Violation

	if doc == nil || doc.Paths == nil || doc.Paths.PathItems == nil {
		return violations
	}

	for path, pathItem := range doc.Paths.PathItems.FromOldest() {
		if pathItem == nil {
			continue
		}

		// Check path-level parameters
		for _, param := range pathItem.Parameters {
			if param == nil {
				continue
			}
			violations = append(violations, r.checkParam(param, path)...)
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

			for _, param := range operation.Parameters {
				if param == nil {
					continue
				}
				violations = append(violations, r.checkParam(param, fmt.Sprintf("%s %s", method, path))...)
			}
		}
	}

	return violations
}

func (r *ProhibitedParameterStylesRule) checkParam(param *v3.Parameter, location string) []Violation {
	var violations []Violation

	if param.Style != "" {
		violations = append(violations, Violation{
			Suggestion: "Remove style and explode from parameters; use default serialization",
			Message:    fmt.Sprintf("Parameter '%s' uses 'style' which is not allowed", param.Name),
			Location:   location,
			RuleName:   r.Name(),
			Severity:   SeverityError,
		})
	}

	if param.Explode != nil {
		violations = append(violations, Violation{
			Suggestion: "Remove style and explode from parameters; use default serialization",
			Message:    fmt.Sprintf("Parameter '%s' uses 'explode' which is not allowed; this field must not be set", param.Name),
			Location:   location,
			RuleName:   r.Name(),
			Severity:   SeverityError,
		})
	}

	return violations
}
