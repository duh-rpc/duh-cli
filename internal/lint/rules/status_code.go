package rules

import (
	"fmt"

	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

var allowedStatusCodes = []string{"200", "400", "401", "403", "404", "429", "452", "453", "454", "455", "500"}

// StatusCodeRule validates only allowed HTTP status codes are used
type StatusCodeRule struct{}

// NewStatusCodeRule creates a new status code rule
func NewStatusCodeRule() *StatusCodeRule {
	return &StatusCodeRule{}
}

// Name returns the rule name
func (r *StatusCodeRule) Name() string {
	return "status-code"
}

// Validate checks that only allowed status codes are used
func (r *StatusCodeRule) Validate(doc *v3.Document) []Violation {
	var violations []Violation

	if doc == nil || doc.Paths == nil || doc.Paths.PathItems == nil {
		return violations
	}

	allowedMap := make(map[string]bool)
	for _, code := range allowedStatusCodes {
		allowedMap[code] = true
	}

	for path, pathItem := range doc.Paths.PathItems.FromOldest() {
		if pathItem == nil {
			continue
		}

		operations := map[string]*v3.Operation{
			"POST": pathItem.Post,
		}

		for method, op := range operations {
			if op == nil || op.Responses == nil || op.Responses.Codes == nil {
				continue
			}

			for statusCode := range op.Responses.Codes.FromOldest() {
				if !allowedMap[statusCode] {
					location := method + " " + path
					violations = append(violations, Violation{
						Suggestion: fmt.Sprintf("Use one of the allowed status codes: %v", allowedStatusCodes),
						Message:    fmt.Sprintf("Status code %s is not allowed", statusCode),
						Location:   location,
						RuleName:   r.Name(),
					})
				}
			}
		}
	}

	return violations
}
