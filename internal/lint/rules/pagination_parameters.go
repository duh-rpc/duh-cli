package rules

import (
	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

type PaginationParametersRule struct{}

func NewPaginationParametersRule() *PaginationParametersRule {
	return &PaginationParametersRule{}
}

func (r *PaginationParametersRule) Name() string {
	return "PAGINATION_PARAMETERS"
}

func (r *PaginationParametersRule) Validate(doc *v3.Document) []Violation {
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

		location := "POST " + path
		suggestion := "Paginated endpoints must use cursor-based pagination with 'first' (integer, min:1, max:100) and 'after' (string) under 'page'"

		pageProxy, hasPage := schema.Properties.Get("page")
		if !hasPage {
			violations = append(violations, Violation{
				Suggestion: suggestion,
				Message:    "Paginated endpoint must have a 'page' sub-object with 'first' and 'after' parameters",
				Location:   location,
				RuleName:   r.Name(),
				Severity:   SeverityError,
			})
			continue
		}

		pageSchema := pageProxy.Schema()
		if pageSchema == nil || pageSchema.Properties == nil {
			violations = append(violations, Violation{
				Suggestion: suggestion,
				Message:    "Paginated endpoint must have a 'page' sub-object with 'first' and 'after' parameters",
				Location:   location,
				RuleName:   r.Name(),
				Severity:   SeverityError,
			})
			continue
		}

		firstProxy, hasFirst := pageSchema.Properties.Get("first")
		if !hasFirst {
			violations = append(violations, Violation{
				Suggestion: suggestion,
				Message:    "Paginated endpoint 'page' must have a 'first' parameter",
				Location:   location,
				RuleName:   r.Name(),
				Severity:   SeverityError,
			})
		} else {
			firstSchema := firstProxy.Schema()
			if firstSchema != nil {
				if len(firstSchema.Type) == 0 || firstSchema.Type[0] != "integer" {
					violations = append(violations, Violation{
						Suggestion: suggestion,
						Message:    "Pagination parameter 'first' must be type integer",
						Location:   location,
						RuleName:   r.Name(),
						Severity:   SeverityError,
					})
				}

				if firstSchema.Minimum == nil || *firstSchema.Minimum != 1 {
					violations = append(violations, Violation{
						Suggestion: suggestion,
						Message:    "Pagination parameter 'first' must have minimum: 1",
						Location:   location,
						RuleName:   r.Name(),
						Severity:   SeverityError,
					})
				}

				if firstSchema.Maximum == nil || *firstSchema.Maximum != 100 {
					violations = append(violations, Violation{
						Suggestion: suggestion,
						Message:    "Pagination parameter 'first' must have maximum: 100",
						Location:   location,
						RuleName:   r.Name(),
						Severity:   SeverityError,
					})
				}
			}
		}

		afterProxy, hasAfter := pageSchema.Properties.Get("after")
		if !hasAfter {
			violations = append(violations, Violation{
				Suggestion: suggestion,
				Message:    "Paginated endpoint 'page' must have an 'after' parameter",
				Location:   location,
				RuleName:   r.Name(),
				Severity:   SeverityError,
			})
		} else {
			afterSchema := afterProxy.Schema()
			if afterSchema != nil && (len(afterSchema.Type) == 0 || afterSchema.Type[0] != "string") {
				violations = append(violations, Violation{
					Suggestion: suggestion,
					Message:    "Pagination parameter 'after' must be type string",
					Location:   location,
					RuleName:   r.Name(),
					Severity:   SeverityError,
				})
			}
		}
	}

	return violations
}
