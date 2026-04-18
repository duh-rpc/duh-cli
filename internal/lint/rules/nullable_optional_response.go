package rules

import (
	"fmt"
	"slices"
	"strings"

	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

type NullableOptionalResponseRule struct{}

func NewNullableOptionalResponseRule() *NullableOptionalResponseRule {
	return &NullableOptionalResponseRule{}
}

func (r *NullableOptionalResponseRule) Name() string {
	return "NULLABLE_OPTIONAL_RESPONSE"
}

func (r *NullableOptionalResponseRule) Validate(doc *v3.Document) []Violation {
	var violations []Violation

	if doc == nil || doc.Paths == nil || doc.Paths.PathItems == nil {
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

			if operation.Responses == nil || operation.Responses.Codes == nil {
				continue
			}

			for statusCode, response := range operation.Responses.Codes.FromOldest() {
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
				schema := jsonContent.Schema.Schema()
				if schema == nil {
					continue
				}

				if isSchemaIgnored(schema, r.Name()) {
					continue
				}

				if schema.Properties == nil {
					continue
				}

				for propName, propProxy := range schema.Properties.FromOldest() {
					propSchema := propProxy.Schema()
					if propSchema == nil {
						continue
					}

					if propSchema.Nullable == nil || !*propSchema.Nullable {
						continue
					}

					if slices.Contains(schema.Required, propName) {
						continue
					}

					var location string
					if ref != "" {
						location = fmt.Sprintf("components/schemas/%s/%s", extractSchemaName(ref), propName)
					} else {
						location = fmt.Sprintf("%s %s response %s/%s", strings.ToUpper(method), pathName, statusCode, propName)
					}

					violations = append(violations, Violation{
						Suggestion: fmt.Sprintf("Either add '%s' to the required array or remove nullable: true", propName),
						Message:    fmt.Sprintf("Response property '%s' must not be both optional and nullable", propName),
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
