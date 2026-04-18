package rules

import (
	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

// RPCRequestResponseUniqueRule ensures each operation uses unique request/response $ref schemas
type RPCRequestResponseUniqueRule struct{}

func NewRPCRequestResponseUniqueRule() *RPCRequestResponseUniqueRule {
	return &RPCRequestResponseUniqueRule{}
}

func (r *RPCRequestResponseUniqueRule) Name() string {
	return "REQUEST_RESPONSE_UNIQUE"
}

func (r *RPCRequestResponseUniqueRule) Validate(doc *v3.Document) []Violation {
	var violations []Violation

	if doc == nil || doc.Paths == nil || doc.Paths.PathItems == nil {
		return violations
	}

	requestRefs := make(map[string][]string)
	responseRefs := make(map[string][]string)

	// Pass 1: Collect all $ref values
	for path, pathItem := range doc.Paths.PathItems.FromOldest() {
		if pathItem == nil || pathItem.Post == nil {
			continue
		}

		if isOperationIgnored(pathItem.Post, r.Name()) {
			continue
		}

		// Collect request body $ref
		if pathItem.Post.RequestBody != nil && pathItem.Post.RequestBody.Content != nil {
			jsonContent, ok := pathItem.Post.RequestBody.Content.Get("application/json")
			if ok && jsonContent != nil && jsonContent.Schema != nil {
				ref := jsonContent.Schema.GetReference()
				if ref != "" {
					requestRefs[ref] = append(requestRefs[ref], path)
				}
			}
		}

		// Collect 2xx response $refs (deduplicate per path)
		if pathItem.Post.Responses != nil && pathItem.Post.Responses.Codes != nil {
			seen := make(map[string]bool)
			for statusCode, response := range pathItem.Post.Responses.Codes.FromOldest() {
				if len(statusCode) != 3 || statusCode[0] != '2' {
					continue
				}

				if response == nil || response.Content == nil {
					continue
				}

				jsonContent, ok := response.Content.Get("application/json")
				if !ok || jsonContent == nil || jsonContent.Schema == nil {
					continue
				}

				ref := jsonContent.Schema.GetReference()
				if ref != "" && !seen[ref] {
					seen[ref] = true
					responseRefs[ref] = append(responseRefs[ref], path)
				}
			}
		}
	}

	// Pass 2: Emit violations for shared refs
	for ref, paths := range requestRefs {
		if len(paths) <= 1 {
			continue
		}
		schemaName := extractSchemaName(ref)
		for _, path := range paths {
			violations = append(violations, Violation{
				Suggestion: "Each operation must use a unique request/response schema",
				Message:    "Request schema '" + schemaName + "' is shared across multiple operations",
				Location:   "POST " + path,
				RuleName:   r.Name(),
				Severity:   SeverityError,
			})
		}
	}

	for ref, paths := range responseRefs {
		if len(paths) <= 1 {
			continue
		}
		schemaName := extractSchemaName(ref)
		for _, path := range paths {
			violations = append(violations, Violation{
				Suggestion: "Each operation must use a unique request/response schema",
				Message:    "Response schema '" + schemaName + "' is shared across multiple operations",
				Location:   "POST " + path,
				RuleName:   r.Name(),
				Severity:   SeverityError,
			})
		}
	}

	return violations
}
