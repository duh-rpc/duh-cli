package rules

import (
	"fmt"

	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

type ProhibitedMultipleExamplesRule struct{}

func NewProhibitedMultipleExamplesRule() *ProhibitedMultipleExamplesRule {
	return &ProhibitedMultipleExamplesRule{}
}

func (r *ProhibitedMultipleExamplesRule) Name() string {
	return "PROHIBITED_MULTIPLE_EXAMPLES"
}

func (r *ProhibitedMultipleExamplesRule) Validate(doc *v3.Document) []Violation {
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
			if param != nil && param.Examples != nil && param.Examples.Len() > 0 {
				violations = append(violations, Violation{
					Suggestion: "Replace 'examples' with a single 'example' value",
					Message:    "Use singular 'example' instead of plural 'examples'",
					Location:   fmt.Sprintf("%s parameter %s", path, param.Name),
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

			if isOperationIgnored(operation, r.Name()) {
				continue
			}

			// Check operation parameters
			for _, param := range operation.Parameters {
				if param != nil && param.Examples != nil && param.Examples.Len() > 0 {
					violations = append(violations, Violation{
						Suggestion: "Replace 'examples' with a single 'example' value",
						Message:    "Use singular 'example' instead of plural 'examples'",
						Location:   fmt.Sprintf("%s %s parameter %s", method, path, param.Name),
						RuleName:   r.Name(),
						Severity:   SeverityWarning,
					})
				}
			}

			// Check request body media types
			if operation.RequestBody != nil && operation.RequestBody.Content != nil {
				for contentType, mediaType := range operation.RequestBody.Content.FromOldest() {
					if mediaType != nil && mediaType.Examples != nil && mediaType.Examples.Len() > 0 {
						violations = append(violations, Violation{
							Suggestion: "Replace 'examples' with a single 'example' value",
							Message:    "Use singular 'example' instead of plural 'examples'",
							Location:   fmt.Sprintf("%s %s request body %s", method, path, contentType),
							RuleName:   r.Name(),
							Severity:   SeverityWarning,
						})
					}
				}
			}

			// Check response media types
			if operation.Responses == nil || operation.Responses.Codes == nil {
				continue
			}

			for statusCode, response := range operation.Responses.Codes.FromOldest() {
				if response == nil || response.Content == nil {
					continue
				}

				for contentType, mediaType := range response.Content.FromOldest() {
					if mediaType != nil && mediaType.Examples != nil && mediaType.Examples.Len() > 0 {
						violations = append(violations, Violation{
							Suggestion: "Replace 'examples' with a single 'example' value",
							Message:    "Use singular 'example' instead of plural 'examples'",
							Location:   fmt.Sprintf("%s %s response %s %s", method, path, statusCode, contentType),
							RuleName:   r.Name(),
							Severity:   SeverityWarning,
						})
					}
				}
			}
		}
	}

	return violations
}
