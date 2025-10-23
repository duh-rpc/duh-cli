package rules

import (
	"github.com/duh-rpc/duhrpc-lint/internal/types"
	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

// RequestBodyRule validates all operations have required request body
type RequestBodyRule struct{}

// NewRequestBodyRule creates a new request body rule
func NewRequestBodyRule() *RequestBodyRule {
	return &RequestBodyRule{}
}

// Name returns the rule name
func (r *RequestBodyRule) Name() string {
	return "request-body-required"
}

// Validate checks that all operations have a required request body
func (r *RequestBodyRule) Validate(doc *v3.Document) []types.Violation {
	var violations []types.Violation

	if doc == nil || doc.Paths == nil || doc.Paths.PathItems == nil {
		return violations
	}

	for path, pathItem := range doc.Paths.PathItems.FromOldest() {
		if pathItem == nil {
			continue
		}

		operations := map[string]*v3.Operation{
			"POST": pathItem.Post,
		}

		for method, op := range operations {
			if op == nil {
				continue
			}

			location := method + " " + path

			// Check if request body is missing
			if op.RequestBody == nil {
				violations = append(violations, types.Violation{
					Suggestion: "Add a required request body to this operation",
					Message:    "Operation is missing a request body",
					Location:   location,
					RuleName:   r.Name(),
				})
				continue
			}

			// Check if request body is not required
			if op.RequestBody.Required == nil || !*op.RequestBody.Required {
				violations = append(violations, types.Violation{
					Suggestion: "Set requestBody.required to true",
					Message:    "Request body must be marked as required",
					Location:   location,
					RuleName:   r.Name(),
				})
			}
		}
	}

	return violations
}
