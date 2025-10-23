package rules

import (
	"strings"

	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

// ContentTypeRule validates only allowed content types are used
type ContentTypeRule struct{}

// NewContentTypeRule creates a new content type rule
func NewContentTypeRule() *ContentTypeRule {
	return &ContentTypeRule{}
}

func (r *ContentTypeRule) Name() string {
	return "content-type"
}

func (r *ContentTypeRule) Validate(doc *v3.Document) []Violation {
	var violations []Violation

	allowedTypes := map[string]bool{
		"application/json":         true,
		"application/protobuf":     true,
		"application/octet-stream": true,
	}

	if doc.Paths == nil || doc.Paths.PathItems == nil {
		return violations
	}

	for pathName, pathItem := range doc.Paths.PathItems.FromOldest() {
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

			// Check request body content types
			if operation.RequestBody != nil && operation.RequestBody.Content != nil {
				hasJSON := false
				hasValidContentType := false
				for contentType := range operation.RequestBody.Content.FromOldest() {
					normalized := strings.ToLower(contentType)
					hasJSON = hasJSON || normalized == "application/json"

					v := r.validateContentType(contentType, allowedTypes, method, pathName, "request body")
					if v != nil {
						violations = append(violations, *v)
					} else {
						hasValidContentType = true
					}
				}

				// Only report missing JSON if we have valid content types but none is JSON
				if !hasJSON && hasValidContentType {
					violations = append(violations, Violation{
						Message:    "Request body must include application/json content type",
						Suggestion: "Add application/json to request body content types",
						RuleName:   r.Name(),
						Location:   method + " " + pathName,
					})
				}
			}

			// Check response content types
			if operation.Responses != nil && operation.Responses.Codes != nil {
				for statusCode, response := range operation.Responses.Codes.FromOldest() {
					if response == nil || response.Content == nil {
						continue
					}

					for contentType := range response.Content.FromOldest() {
						location := method + " " + pathName + " response " + statusCode
						v := r.validateContentType(contentType, allowedTypes, method, pathName+" response "+statusCode, "")
						if v != nil {
							v.Location = location
							violations = append(violations, *v)
						}
					}
				}
			}
		}
	}

	return violations
}

func (r *ContentTypeRule) validateContentType(contentType string, allowedTypes map[string]bool, method, path, context string) *Violation {
	normalized := strings.ToLower(contentType)

	// Check for MIME parameters
	if strings.Contains(normalized, ";") {
		msg := "MIME parameters not allowed in content type"
		if context != "" {
			msg = "MIME parameters not allowed in " + context + " content type"
		}
		return &Violation{
			Message:    msg,
			Suggestion: "Remove parameters from content type (use '" + strings.Split(normalized, ";")[0] + "' instead of '" + contentType + "')",
			RuleName:   r.Name(),
			Location:   method + " " + path,
		}
	}

	// Check if content type is allowed
	if !allowedTypes[normalized] {
		msg := "Invalid content type: " + contentType
		if context != "" {
			msg = "Invalid " + context + " content type: " + contentType
		}
		return &Violation{
			Message:    msg,
			Suggestion: "Use one of: application/json, application/protobuf, application/octet-stream",
			RuleName:   r.Name(),
			Location:   method + " " + path,
		}
	}

	return nil
}
