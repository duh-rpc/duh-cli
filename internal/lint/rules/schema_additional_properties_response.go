package rules

import (
	"fmt"

	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

type SchemaAdditionalPropertiesResponseRule struct{}

func NewSchemaAdditionalPropertiesResponseRule() *SchemaAdditionalPropertiesResponseRule {
	return &SchemaAdditionalPropertiesResponseRule{}
}

func (r *SchemaAdditionalPropertiesResponseRule) Name() string {
	return "SCHEMA_ADDITIONAL_PROPERTIES_RESPONSE"
}

func (r *SchemaAdditionalPropertiesResponseRule) Validate(doc *v3.Document) []Violation {
	var violations []Violation

	if doc == nil || doc.Paths == nil || doc.Paths.PathItems == nil {
		return violations
	}

	// First pass: collect response schema names
	responseSchemas := make(map[string]bool)

	for _, pathItem := range doc.Paths.PathItems.FromOldest() {
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

		for _, operation := range operations {
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
				if ref == "" {
					continue
				}

				responseSchemas[extractSchemaName(ref)] = true
			}
		}
	}

	// Second pass: check collected response schemas
	if doc.Components == nil || doc.Components.Schemas == nil {
		return violations
	}

	for schemaName, schemaProxy := range doc.Components.Schemas.FromOldest() {
		if !responseSchemas[schemaName] {
			continue
		}

		schema := schemaProxy.Schema()
		if schema == nil {
			continue
		}

		if schema.AdditionalProperties != nil && schema.AdditionalProperties.IsB() && !schema.AdditionalProperties.B {
			violations = append(violations, Violation{
				Suggestion: "Remove additionalProperties: false from response schemas to allow forward-compatible extensions",
				Message:    "Response schema must not use additionalProperties: false",
				Location:   fmt.Sprintf("components/schemas/%s", schemaName),
				RuleName:   r.Name(),
				Severity:   SeverityError,
			})
		}
	}

	return violations
}
