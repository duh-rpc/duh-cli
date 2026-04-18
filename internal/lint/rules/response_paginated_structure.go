package rules

import (
	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

type ResponsePaginatedStructureRule struct{}

func NewResponsePaginatedStructureRule() *ResponsePaginatedStructureRule {
	return &ResponsePaginatedStructureRule{}
}

func (r *ResponsePaginatedStructureRule) Name() string {
	return "RESPONSE_PAGINATED_STRUCTURE"
}

func (r *ResponsePaginatedStructureRule) Validate(doc *v3.Document) []Violation {
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

		if isOperationIgnored(pathItem.Post, r.Name()) {
			continue
		}

		if pathItem.Post.Responses == nil || pathItem.Post.Responses.Codes == nil {
			continue
		}

		for statusCode, response := range pathItem.Post.Responses.Codes.FromOldest() {
			if len(statusCode) == 0 || statusCode[0] != '2' {
				continue
			}

			if response == nil || response.Content == nil {
				continue
			}

			jsonContent, ok := response.Content.Get("application/json")
			if !ok || jsonContent == nil || jsonContent.Schema == nil {
				continue
			}

			schema := jsonContent.Schema.Schema()
			if schema == nil || schema.Properties == nil {
				continue
			}

			location := "POST " + path + " response " + statusCode

			itemsProxy, hasItems := schema.Properties.Get("items")
			if !hasItems {
				violations = append(violations, Violation{
					Suggestion: "Paginated responses must include an 'items' array and a 'pagination' object with 'endCursor'",
					Message:    "Paginated response must have an 'items' property",
					Location:   location,
					RuleName:   r.Name(),
					Severity:   SeverityError,
				})
			} else {
				itemsSchema := itemsProxy.Schema()
				if itemsSchema != nil && (len(itemsSchema.Type) == 0 || itemsSchema.Type[0] != "array") {
					violations = append(violations, Violation{
						Suggestion: "Paginated responses must include an 'items' array and a 'pagination' object with 'endCursor'",
						Message:    "The 'items' property must be type array",
						Location:   location,
						RuleName:   r.Name(),
						Severity:   SeverityError,
					})
				}
			}

			pageProxy, hasPage := schema.Properties.Get("pagination")
			if !hasPage {
				violations = append(violations, Violation{
					Suggestion: "Paginated responses must include an 'items' array and a 'pagination' object with 'endCursor'",
					Message:    "Paginated response must have a 'pagination' property",
					Location:   location,
					RuleName:   r.Name(),
					Severity:   SeverityError,
				})
			} else {
				pageSchema := pageProxy.Schema()
				if pageSchema != nil && pageSchema.Properties != nil {
					if _, hasEndCursor := pageSchema.Properties.Get("endCursor"); !hasEndCursor {
						violations = append(violations, Violation{
							Suggestion: "Paginated responses must include an 'items' array and a 'pagination' object with 'endCursor'",
							Message:    "Paginated response 'pagination' must contain 'endCursor' property",
							Location:   location,
							RuleName:   r.Name(),
							Severity:   SeverityError,
						})
					}
				} else if pageSchema != nil && pageSchema.Properties == nil {
					violations = append(violations, Violation{
						Suggestion: "Paginated responses must include an 'items' array and a 'pagination' object with 'endCursor'",
						Message:    "Paginated response 'pagination' must contain 'endCursor' property",
						Location:   location,
						RuleName:   r.Name(),
						Severity:   SeverityError,
					})
				}
			}
		}
	}

	return violations
}
