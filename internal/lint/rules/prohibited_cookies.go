package rules

import (
	"fmt"

	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

type ProhibitedCookiesRule struct{}

func NewProhibitedCookiesRule() *ProhibitedCookiesRule {
	return &ProhibitedCookiesRule{}
}

func (r *ProhibitedCookiesRule) Name() string {
	return "PROHIBITED_COOKIES"
}

func (r *ProhibitedCookiesRule) Validate(doc *v3.Document) []Violation {
	var violations []Violation

	if doc != nil && doc.Paths != nil && doc.Paths.PathItems != nil {
		for path, pathItem := range doc.Paths.PathItems.FromOldest() {
			if pathItem == nil {
				continue
			}

			// Check path-level parameters
			for _, param := range pathItem.Parameters {
				if param != nil && param.In == "cookie" {
					violations = append(violations, Violation{
						Suggestion: "Remove cookie parameters; use request body or authorization headers instead",
						Message:    fmt.Sprintf("Cookie parameter '%s' is not allowed", param.Name),
						Location:   path,
						RuleName:   r.Name(),
						Severity:   SeverityError,
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

				for _, param := range operation.Parameters {
					if param != nil && param.In == "cookie" {
						violations = append(violations, Violation{
							Suggestion: "Remove cookie parameters; use request body or authorization headers instead",
							Message:    fmt.Sprintf("Cookie parameter '%s' is not allowed", param.Name),
							Location:   fmt.Sprintf("%s %s", method, path),
							RuleName:   r.Name(),
							Severity:   SeverityError,
						})
					}
				}
			}
		}
	}

	// Check security schemes
	if doc != nil && doc.Components != nil && doc.Components.SecuritySchemes != nil {
		for schemeName, scheme := range doc.Components.SecuritySchemes.FromOldest() {
			if scheme != nil && scheme.Type == "apiKey" && scheme.In == "cookie" {
				violations = append(violations, Violation{
					Suggestion: "Use Bearer token or other header-based authentication",
					Message:    "Cookie-based security scheme is not allowed",
					Location:   fmt.Sprintf("components/securitySchemes/%s", schemeName),
					RuleName:   r.Name(),
					Severity:   SeverityError,
				})
			}
		}
	}

	return violations
}
