package rules

import (
	"strings"

	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

type SchemaNoInlineObjectsRule struct{}

func NewSchemaNoInlineObjectsRule() *SchemaNoInlineObjectsRule {
	return &SchemaNoInlineObjectsRule{}
}

func (r *SchemaNoInlineObjectsRule) Name() string {
	return "SCHEMA_NO_INLINE_OBJECTS"
}

func (r *SchemaNoInlineObjectsRule) Validate(doc *v3.Document) []Violation {
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

			if isOperationIgnored(operation, r.Name()) {
				continue
			}

			// Check request body
			if operation.RequestBody != nil && operation.RequestBody.Content != nil {
				for _, mediaType := range operation.RequestBody.Content.FromOldest() {
					if mediaType == nil || mediaType.Schema == nil {
						continue
					}

					if mediaType.Schema.GetReference() == "" {
						schema := mediaType.Schema.Schema()
						if schema != nil && schema.Properties != nil && schema.Properties.Len() > 0 {
							violations = append(violations, Violation{
								Suggestion: "Move inline schema to components/schemas and use $ref to reference it",
								Message:    "Schema must be defined in components/schemas and referenced via $ref",
								Location:   strings.ToUpper(method) + " " + pathName + " request body",
								RuleName:   r.Name(),
								Severity:   SeverityError,
							})
						}
					}
				}
			}

			// Check 2xx responses only
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

				for _, mediaType := range response.Content.FromOldest() {
					if mediaType == nil || mediaType.Schema == nil {
						continue
					}

					if mediaType.Schema.GetReference() == "" {
						schema := mediaType.Schema.Schema()
						if schema != nil && schema.Properties != nil && schema.Properties.Len() > 0 {
							violations = append(violations, Violation{
								Suggestion: "Move inline schema to components/schemas and use $ref to reference it",
								Message:    "Schema must be defined in components/schemas and referenced via $ref",
								Location:   strings.ToUpper(method) + " " + pathName + " response " + statusCode,
								RuleName:   r.Name(),
								Severity:   SeverityError,
							})
						}
					}
				}
			}
		}
	}

	return violations
}
