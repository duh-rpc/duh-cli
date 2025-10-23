package rules

import (
	"fmt"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

// ErrorResponseRule validates error response schemas have required structure
type ErrorResponseRule struct{}

// NewErrorResponseRule creates a new error response rule
func NewErrorResponseRule() *ErrorResponseRule {
	return &ErrorResponseRule{}
}

func (r *ErrorResponseRule) Name() string {
	return "error-response-schema"
}

func (r *ErrorResponseRule) Validate(doc *v3.Document) []Violation {
	var violations []Violation

	errorStatusCodes := map[string]bool{
		"400": true,
		"401": true,
		"403": true,
		"404": true,
		"429": true,
		"452": true,
		"453": true,
		"454": true,
		"455": true,
		"500": true,
	}

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

			if operation.Responses == nil || operation.Responses.Codes == nil {
				continue
			}

			for statusCode, response := range operation.Responses.Codes.FromOldest() {
				if !errorStatusCodes[statusCode] {
					continue
				}

				if response == nil || response.Content == nil {
					continue
				}

				// Check each content type's schema
				for contentType, mediaType := range response.Content.FromOldest() {
					if mediaType == nil || mediaType.Schema == nil {
						continue
					}

					location := method + " " + pathName + " response " + statusCode + " (" + contentType + ")"

					// Get the schema from the SchemaProxy
					schema := mediaType.Schema.Schema()
					if schema == nil {
						// Unresolved reference - skip validation
						continue
					}

					if err := r.validateErrorSchema(schema, make(map[*base.Schema]bool)); err != nil {
						violations = append(violations, Violation{
							Message:    err.Error(),
							Suggestion: "Error response schema must be type 'object' with required fields [code, message] where code is integer and message is string. Optional details field must be type object.",
							RuleName:   r.Name(),
							Location:   location,
						})
					}
				}
			}
		}
	}

	return violations
}

// validateErrorSchema checks if a schema meets DUH-RPC error response requirements
// visited tracks schemas we've already checked to handle circular references
func (r *ErrorResponseRule) validateErrorSchema(schema *base.Schema, visited map[*base.Schema]bool) error {
	if schema == nil {
		return fmt.Errorf("schema is nil")
	}

	// Prevent infinite loops from circular references
	if visited[schema] {
		return nil
	}
	visited[schema] = true

	// Handle allOf - all sub-schemas must combine to meet requirements
	if len(schema.AllOf) > 0 {
		// Collect all properties and required fields from allOf schemas
		allProperties := make(map[string]*base.SchemaProxy)
		allRequired := make(map[string]bool)
		hasObjectType := false

		for _, subSchemaProxy := range schema.AllOf {
			subSchema := subSchemaProxy.Schema()
			if subSchema == nil {
				continue
			}

			// Check if any sub-schema has type object
			if len(subSchema.Type) > 0 && subSchema.Type[0] == "object" {
				hasObjectType = true
			}

			// Collect properties
			if subSchema.Properties != nil {
				for propName, propProxy := range subSchema.Properties.FromOldest() {
					allProperties[propName] = propProxy
				}
			}

			// Collect required fields
			for _, req := range subSchema.Required {
				allRequired[req] = true
			}
		}

		// Now validate combined schema
		if !hasObjectType {
			return fmt.Errorf("error schema must have explicit type 'object'")
		}

		if !allRequired["code"] || !allRequired["message"] {
			return fmt.Errorf("error schema must have required fields: code and message")
		}

		// Validate field types
		if codeProxy, ok := allProperties["code"]; ok {
			codeSchema := codeProxy.Schema()
			if codeSchema == nil || len(codeSchema.Type) == 0 || codeSchema.Type[0] != "integer" {
				return fmt.Errorf("code field must be type integer")
			}
		} else {
			return fmt.Errorf("code field not found in schema properties")
		}

		if msgProxy, ok := allProperties["message"]; ok {
			msgSchema := msgProxy.Schema()
			if msgSchema == nil || len(msgSchema.Type) == 0 || msgSchema.Type[0] != "string" {
				return fmt.Errorf("message field must be type string")
			}
		} else {
			return fmt.Errorf("message field not found in schema properties")
		}

		// Check details field if present
		if detailsProxy, ok := allProperties["details"]; ok {
			detailsSchema := detailsProxy.Schema()
			if detailsSchema != nil && len(detailsSchema.Type) > 0 && detailsSchema.Type[0] != "object" {
				return fmt.Errorf("details field must be type object")
			}
		}

		return nil
	}

	// Handle oneOf/anyOf - at least one sub-schema must meet requirements
	if len(schema.OneOf) > 0 || len(schema.AnyOf) > 0 {
		schemas := schema.OneOf
		if len(schemas) == 0 {
			schemas = schema.AnyOf
		}

		for _, subSchemaProxy := range schemas {
			subSchema := subSchemaProxy.Schema()
			if subSchema == nil {
				continue
			}

			if err := r.validateErrorSchema(subSchema, visited); err == nil {
				return nil
			}
		}

		return fmt.Errorf("none of the oneOf/anyOf schemas meet error response requirements")
	}

	// Direct schema validation
	if len(schema.Type) == 0 || schema.Type[0] != "object" {
		return fmt.Errorf("error schema must have explicit type 'object'")
	}

	// Check required fields
	hasCode := false
	hasMessage := false
	for _, req := range schema.Required {
		if req == "code" {
			hasCode = true
		}
		if req == "message" {
			hasMessage = true
		}
	}

	if !hasCode || !hasMessage {
		return fmt.Errorf("error schema must have required fields: code and message")
	}

	// Check that properties exist and have correct types
	if schema.Properties == nil {
		return fmt.Errorf("error schema must define properties for code and message")
	}

	codeProxy, hasCodeProp := schema.Properties.Get("code")
	if !hasCodeProp {
		return fmt.Errorf("code field not found in schema properties")
	}
	codeSchema := codeProxy.Schema()
	if codeSchema == nil || len(codeSchema.Type) == 0 || codeSchema.Type[0] != "integer" {
		return fmt.Errorf("code field must be type integer")
	}

	msgProxy, hasMsgProp := schema.Properties.Get("message")
	if !hasMsgProp {
		return fmt.Errorf("message field not found in schema properties")
	}
	msgSchema := msgProxy.Schema()
	if msgSchema == nil || len(msgSchema.Type) == 0 || msgSchema.Type[0] != "string" {
		return fmt.Errorf("message field must be type string")
	}

	// Check details field if present
	if detailsProxy, hasDetails := schema.Properties.Get("details"); hasDetails {
		detailsSchema := detailsProxy.Schema()
		if detailsSchema != nil && len(detailsSchema.Type) > 0 && detailsSchema.Type[0] != "object" {
			return fmt.Errorf("details field must be type object")
		}
	}

	return nil
}
