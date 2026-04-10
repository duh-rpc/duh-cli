package rules

import (
	"strings"

	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

// RPCPaginationParametersRule flags paginated endpoints that use 'limit' in the request body
type RPCPaginationParametersRule struct{}

func NewRPCPaginationParametersRule() *RPCPaginationParametersRule {
	return &RPCPaginationParametersRule{}
}

func (r *RPCPaginationParametersRule) Name() string {
	return "RPC_PAGINATION_PARAMETERS"
}

func (r *RPCPaginationParametersRule) Validate(doc *v3.Document) []Violation {
	var violations []Violation

	if doc == nil || doc.Paths == nil || doc.Paths.PathItems == nil {
		return violations
	}

	for path, pathItem := range doc.Paths.PathItems.FromOldest() {
		if !isPaginatedEndpoint(path) {
			continue
		}

		if pathItem == nil || pathItem.Post == nil {
			continue
		}

		if pathItem.Post.RequestBody == nil || pathItem.Post.RequestBody.Content == nil {
			continue
		}

		jsonContent, ok := pathItem.Post.RequestBody.Content.Get("application/json")
		if !ok || jsonContent == nil || jsonContent.Schema == nil {
			continue
		}

		schema := jsonContent.Schema.Schema()
		if schema == nil || schema.Properties == nil {
			continue
		}

		if _, exists := schema.Properties.Get("limit"); exists {
			violations = append(violations, Violation{
				Suggestion: "Replace 'limit' with 'first' parameter nested under a 'page' sub-object",
				Message:    "Paginated endpoint uses 'limit'; use 'first' for cursor-based pagination",
				Location:   "POST " + path,
				RuleName:   r.Name(),
				Severity:   SeverityError,
			})
		}
	}

	return violations
}

func isPaginatedEndpoint(path string) bool {
	return strings.HasSuffix(path, ".list") ||
		strings.HasSuffix(path, ".search") ||
		strings.HasSuffix(path, ".query")
}
