package rules

import (
	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

type PaginationNoLimitOffsetRule struct{}

func NewPaginationNoLimitOffsetRule() *PaginationNoLimitOffsetRule {
	return &PaginationNoLimitOffsetRule{}
}

func (r *PaginationNoLimitOffsetRule) Name() string {
	return "PAGINATION_NO_LIMIT_OFFSET"
}

func (r *PaginationNoLimitOffsetRule) Validate(doc *v3.Document) []Violation {
	var violations []Violation

	if doc == nil || doc.Paths == nil || doc.Paths.PathItems == nil {
		return violations
	}

	prohibited := map[string]string{
		"limit":  "Do not use 'limit' for pagination; use cursor-based pagination with 'pagination.first' and 'pagination.after'",
		"offset": "Do not use 'offset' for pagination; use cursor-based pagination with 'pagination.first' and 'pagination.after'",
	}

	for path, pathItem := range doc.Paths.PathItems.FromOldest() {
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

		location := "POST " + path

		for propName, propProxy := range schema.Properties.FromOldest() {
			normalized := normalize(propName)

			for prohibited, suggestion := range prohibited {
				if normalized == prohibited {
					violations = append(violations, Violation{
						Suggestion: suggestion,
						Message:    "Request body must not use '" + propName + "' for pagination",
						Location:   location,
						RuleName:   r.Name(),
						Severity:   SeverityError,
					})
				}
			}

			if normalized == "page" {
				propSchema := propProxy.Schema()
				if propSchema != nil && len(propSchema.Type) > 0 && propSchema.Type[0] == "integer" {
					violations = append(violations, Violation{
						Suggestion: "Do not use 'page' as an integer parameter; use cursor-based pagination with 'pagination.first' and 'pagination.after'",
						Message:    "Request body must not use '" + propName + "' as an integer page number",
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
