package rules

import (
	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

// RPCPaginatedRequestStructureRule flags paginated endpoints with first/after at root level
type RPCPaginatedRequestStructureRule struct{}

func NewRPCPaginatedRequestStructureRule() *RPCPaginatedRequestStructureRule {
	return &RPCPaginatedRequestStructureRule{}
}

func (r *RPCPaginatedRequestStructureRule) Name() string {
	return "RPC_PAGINATED_REQUEST_STRUCTURE"
}

func (r *RPCPaginatedRequestStructureRule) Validate(doc *v3.Document) []Violation {
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

		if _, exists := schema.Properties.Get("first"); exists {
			violations = append(violations, Violation{
				Suggestion: "Move pagination parameters under a 'page' sub-object in the request body",
				Message:    "Pagination parameter 'first' must be nested under 'page' sub-object",
				Location:   "POST " + path,
				RuleName:   r.Name(),
				Severity:   SeverityError,
			})
		}

		if _, exists := schema.Properties.Get("after"); exists {
			violations = append(violations, Violation{
				Suggestion: "Move pagination parameters under a 'page' sub-object in the request body",
				Message:    "Pagination parameter 'after' must be nested under 'page' sub-object",
				Location:   "POST " + path,
				RuleName:   r.Name(),
				Severity:   SeverityError,
			})
		}
	}

	return violations
}
