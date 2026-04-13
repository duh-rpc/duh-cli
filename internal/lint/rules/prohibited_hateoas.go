package rules

import (
	"fmt"

	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

type ProhibitedHATEOASRule struct{}

func NewProhibitedHATEOASRule() *ProhibitedHATEOASRule {
	return &ProhibitedHATEOASRule{}
}

func (r *ProhibitedHATEOASRule) Name() string {
	return "PROHIBITED_HATEOAS"
}

func (r *ProhibitedHATEOASRule) Validate(doc *v3.Document) []Violation {
	var violations []Violation

	if doc == nil || doc.Paths == nil || doc.Paths.PathItems == nil {
		return violations
	}

	for path, pathItem := range doc.Paths.PathItems.FromOldest() {
		if pathItem == nil {
			continue
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

			if operation.Responses == nil || operation.Responses.Codes == nil {
				continue
			}

			for statusCode, response := range operation.Responses.Codes.FromOldest() {
				if response == nil {
					continue
				}

				if response.Links != nil && response.Links.Len() > 0 {
					violations = append(violations, Violation{
						Suggestion: "Remove links from responses; use explicit API endpoints instead",
						Message:    "Response contains links (HATEOAS) which is not allowed",
						Location:   fmt.Sprintf("%s %s response %s", method, path, statusCode),
						RuleName:   r.Name(),
						Severity:   SeverityError,
					})
				}
			}
		}
	}

	return violations
}
