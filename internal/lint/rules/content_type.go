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
	return "CONTENT_TYPE"
}

var standardRequestTypes = map[string]bool{
	"application/json":     true,
	"application/protobuf": true,
}

var standardResponseTypes = map[string]bool{
	"application/json":                true,
	"application/protobuf":            true,
	"application/octet-stream":        true,
	"application/duh-stream+json":     true,
	"application/duh-stream+protobuf": true,
}

func isStandardDUHType(contentType string) bool {
	normalized := strings.ToLower(strings.TrimSpace(contentType))
	return standardRequestTypes[normalized] || standardResponseTypes[normalized]
}

func (r *ContentTypeRule) Validate(doc *v3.Document) []Violation {
	var violations []Violation

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

			if isOperationIgnored(operation, r.Name()) {
				continue
			}

			r.validateRequestBody(&violations, operation, method, pathName)
			r.validateResponses(&violations, operation, method, pathName)
		}
	}

	return violations
}

func (r *ContentTypeRule) validateRequestBody(violations *[]Violation, operation *v3.Operation, method, pathName string) {
	if operation.RequestBody == nil || operation.RequestBody.Content == nil {
		return
	}

	hasJSON := false
	hasStandardType := false

	for contentType := range operation.RequestBody.Content.FromOldest() {
		normalized := strings.ToLower(contentType)
		hasJSON = hasJSON || normalized == "application/json"

		if isStandardDUHType(contentType) {
			hasStandardType = true
		}

		v := r.validateRequestContentType(contentType, method, pathName)
		if v != nil {
			*violations = append(*violations, *v)
		}
	}

	// Require JSON when standard DUH types are present (JSON fallback rule).
	// Content endpoint request bodies (only non-DUH types) don't need JSON.
	if !hasJSON && hasStandardType {
		*violations = append(*violations, Violation{
			Message:    "Request body must include application/json content type",
			Suggestion: "Add application/json to request body content types",
			Location:   method + " " + pathName,
			RuleName:   r.Name(),
			Severity:   SeverityError,
		})
	}
}

func (r *ContentTypeRule) validateRequestContentType(contentType, method, pathName string) *Violation {
	normalized := strings.ToLower(contentType)

	if strings.HasPrefix(normalized, "multipart/") || normalized == "application/x-www-form-urlencoded" {
		return &Violation{
			Message:    "Multipart and form-encoded content types are not allowed",
			Suggestion: "Use application/json or application/protobuf",
			Location:   method + " " + pathName,
			RuleName:   r.Name(),
			Severity:   SeverityError,
		}
	}

	// MIME parameters on standard DUH types are not allowed
	if strings.Contains(normalized, ";") {
		base := strings.TrimSpace(strings.Split(normalized, ";")[0])
		if isStandardDUHType(base) {
			return &Violation{
				Message:    "MIME parameters not allowed in request body content type",
				Suggestion: "Remove parameters from content type (use '" + base + "' instead of '" + contentType + "')",
				Location:   method + " " + pathName,
				RuleName:   r.Name(),
				Severity:   SeverityError,
			}
		}
		// MIME parameters on content endpoint types are allowed
		return nil
	}

	// Standard DUH types: validate against allowed list
	if standardRequestTypes[normalized] {
		return nil
	}

	// Streaming types are only allowed in responses
	if isStreamingType(normalized) {
		return &Violation{
			Message:    "Invalid request body content type: " + contentType,
			Suggestion: "Streaming content types are only allowed in responses",
			Location:   method + " " + pathName,
			RuleName:   r.Name(),
			Severity:   SeverityError,
		}
	}

	// Non-standard type: allowed as content endpoint request body
	return nil
}

func (r *ContentTypeRule) validateResponses(violations *[]Violation, operation *v3.Operation, method, pathName string) {
	if operation.Responses == nil || operation.Responses.Codes == nil {
		return
	}

	for statusCode, response := range operation.Responses.Codes.FromOldest() {
		if response == nil || response.Content == nil {
			continue
		}

		isError := len(statusCode) == 3 && (statusCode[0] == '4' || statusCode[0] == '5')

		for contentType := range response.Content.FromOldest() {
			location := method + " " + pathName + " response " + statusCode
			v := r.validateResponseContentType(contentType, isError, location)
			if v != nil {
				*violations = append(*violations, *v)
			}
		}
	}
}

func (r *ContentTypeRule) validateResponseContentType(contentType string, isError bool, location string) *Violation {
	normalized := strings.ToLower(contentType)

	// MIME parameters on standard DUH types are not allowed
	if strings.Contains(normalized, ";") {
		base := strings.TrimSpace(strings.Split(normalized, ";")[0])
		if isStandardDUHType(base) {
			return &Violation{
				Message:    "MIME parameters not allowed in content type",
				Suggestion: "Remove parameters from content type (use '" + base + "' instead of '" + contentType + "')",
				Location:   location,
				RuleName:   r.Name(),
				Severity:   SeverityError,
			}
		}
		// MIME parameters on content endpoint types are allowed
		if isError {
			return &Violation{
				Message:    "Error response must use a standard DUH-RPC content type",
				Suggestion: "Use application/json or application/protobuf for error responses",
				Location:   location,
				RuleName:   r.Name(),
				Severity:   SeverityError,
			}
		}
		return nil
	}

	// Multipart and form-encoded are always prohibited
	if strings.HasPrefix(normalized, "multipart/") || normalized == "application/x-www-form-urlencoded" {
		return &Violation{
			Message:    "Multipart and form-encoded content types are not allowed",
			Suggestion: "Use application/json or application/protobuf",
			Location:   location,
			RuleName:   r.Name(),
			Severity:   SeverityError,
		}
	}

	// Standard DUH response types are always allowed
	if standardResponseTypes[normalized] {
		return nil
	}

	// Error responses must use standard DUH types
	if isError {
		return &Violation{
			Message:    "Error response must use a standard DUH-RPC content type",
			Suggestion: "Use application/json or application/protobuf for error responses",
			Location:   location,
			RuleName:   r.Name(),
			Severity:   SeverityError,
		}
	}

	// Non-error responses: non-standard types are allowed (content endpoints)
	return nil
}

func isStreamingType(normalized string) bool {
	switch normalized {
	case "application/duh-stream+json", "application/duh-stream+protobuf":
		return true
	}
	return false
}
