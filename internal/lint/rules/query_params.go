package rules

import (
	"fmt"

	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

// QueryParamsRule validates no query parameters are used
type QueryParamsRule struct{}

func NewQueryParamsRule() *QueryParamsRule {
	return &QueryParamsRule{}
}

func (r *QueryParamsRule) Name() string {
	return "query-parameters"
}

func (r *QueryParamsRule) Validate(doc *v3.Document) []Violation {
	var violations []Violation

	if doc == nil || doc.Paths == nil || doc.Paths.PathItems == nil {
		return violations
	}

	for path, pathItem := range doc.Paths.PathItems.FromOldest() {
		if pathItem == nil {
			continue
		}

		// Check all operations
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

			// Check operation parameters
			if operation.Parameters != nil {
				for _, param := range operation.Parameters {
					if param != nil && param.In == "query" {
						violations = append(violations, Violation{
							RuleName:   r.Name(),
							Location:   fmt.Sprintf("%s %s", method, path),
							Message:    fmt.Sprintf("Query parameter '%s' is not allowed in DUH-RPC", param.Name),
							Suggestion: fmt.Sprintf("Move '%s' to request body", param.Name),
						})
					}
				}
			}
		}
	}

	return violations
}
