package rules

import (
	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

// SuccessResponseRule validates 200 response exists with content
type SuccessResponseRule struct{}

// NewSuccessResponseRule creates a new success response rule
func NewSuccessResponseRule() *SuccessResponseRule {
	return &SuccessResponseRule{}
}

// Name returns the rule name
func (r *SuccessResponseRule) Name() string {
	return "success-response"
}

// Validate checks that all operations have a 200 response with content
func (r *SuccessResponseRule) Validate(doc *v3.Document) []Violation {
	var violations []Violation

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
			if op == nil || op.Responses == nil || op.Responses.Codes == nil {
				continue
			}

			location := method + " " + path

			// Check if 200 response exists
			var response200 *v3.Response
			for code, resp := range op.Responses.Codes.FromOldest() {
				if code == "200" {
					response200 = resp
					break
				}
			}

			if response200 == nil {
				violations = append(violations, Violation{
					Suggestion: "Add a 200 response with content to this operation",
					Message:    "Operation is missing a 200 (success) response",
					Location:   location,
					RuleName:   r.Name(),
				})
				continue
			}

			// Check if 200 response has content
			if response200.Content == nil || response200.Content.Len() == 0 {
				violations = append(violations, Violation{
					Suggestion: "Add content with a schema to the 200 response",
					Message:    "200 response is missing content",
					Location:   location,
					RuleName:   r.Name(),
				})
				continue
			}

			// Check if at least one media type has a schema
			hasSchema := false
			for _, mediaType := range response200.Content.FromOldest() {
				if mediaType != nil && mediaType.Schema != nil {
					hasSchema = true
					break
				}
			}

			if !hasSchema {
				violations = append(violations, Violation{
					Suggestion: "Add a schema to at least one media type in the 200 response",
					Message:    "200 response content is missing a schema",
					Location:   location,
					RuleName:   r.Name(),
				})
			}
		}
	}

	return violations
}
