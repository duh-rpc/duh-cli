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
	return "ERROR_SCHEMA"
}

const errorSchemaSuggestion = "Error response schema must be type 'object' with required field [message] (string). " +
	"Optional fields: code (string), type (string), details (object with additionalProperties: { type: string })."

func (r *ErrorResponseRule) Validate(doc *v3.Document) []Violation {
	var violations []Violation

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
				if len(statusCode) != 3 || (statusCode[0] != '4' && statusCode[0] != '5') {
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

					schema := mediaType.Schema.Schema()
					if schema == nil {
						continue
					}

					if err := r.validateErrorSchema(schema, make(map[*base.Schema]bool)); err != nil {
						violations = append(violations, Violation{
							Suggestion: errorSchemaSuggestion,
							Message:    err.Error(),
							Location:   location,
							RuleName:   r.Name(),
							Severity:   SeverityError,
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
		allProperties := make(map[string]*base.SchemaProxy)
		allRequired := make(map[string]bool)
		hasObjectType := false

		for _, subSchemaProxy := range schema.AllOf {
			subSchema := subSchemaProxy.Schema()
			if subSchema == nil {
				continue
			}

			if len(subSchema.Type) > 0 && subSchema.Type[0] == "object" {
				hasObjectType = true
			}

			if subSchema.Properties != nil {
				for propName, propProxy := range subSchema.Properties.FromOldest() {
					allProperties[propName] = propProxy
				}
			}

			for _, req := range subSchema.Required {
				allRequired[req] = true
			}
		}

		if !hasObjectType {
			return fmt.Errorf("error schema must have explicit type 'object'")
		}

		if !allRequired["message"] {
			return fmt.Errorf("error schema must have required field: message")
		}

		return r.validateFields(allProperties)
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

	hasMessage := false
	for _, req := range schema.Required {
		if req == "message" {
			hasMessage = true
		}
	}

	if !hasMessage {
		return fmt.Errorf("error schema must have required field: message")
	}

	if schema.Properties == nil {
		return fmt.Errorf("error schema must define properties for message")
	}

	// Collect properties into a map for shared validation
	properties := make(map[string]*base.SchemaProxy)
	for propName, propProxy := range schema.Properties.FromOldest() {
		properties[propName] = propProxy
	}

	return r.validateFields(properties)
}

// validateFields checks that known fields have the correct types
func (r *ErrorResponseRule) validateFields(properties map[string]*base.SchemaProxy) error {
	// message must exist and be type string
	if msgProxy, ok := properties["message"]; ok {
		msgSchema := msgProxy.Schema()
		if msgSchema == nil || len(msgSchema.Type) == 0 || msgSchema.Type[0] != "string" {
			return fmt.Errorf("message field must be type string")
		}
	} else {
		return fmt.Errorf("message field not found in schema properties")
	}

	// code is optional, but if present must be type string
	if codeProxy, ok := properties["code"]; ok {
		codeSchema := codeProxy.Schema()
		if codeSchema == nil || len(codeSchema.Type) == 0 || codeSchema.Type[0] != "string" {
			return fmt.Errorf("code field must be type string")
		}
	}

	// type is optional, but if present must be type string
	if typeProxy, ok := properties["type"]; ok {
		typeSchema := typeProxy.Schema()
		if typeSchema == nil || len(typeSchema.Type) == 0 || typeSchema.Type[0] != "string" {
			return fmt.Errorf("type field must be type string")
		}
	}

	// details is optional, but if present must be type object with additionalProperties: { type: string }
	if detailsProxy, ok := properties["details"]; ok {
		detailsSchema := detailsProxy.Schema()
		if detailsSchema == nil || len(detailsSchema.Type) == 0 || detailsSchema.Type[0] != "object" {
			return fmt.Errorf("details field must be type object")
		}

		if detailsSchema.AdditionalProperties == nil {
			return fmt.Errorf("details field must have additionalProperties with type string")
		}

		if detailsSchema.AdditionalProperties.IsB() {
			return fmt.Errorf("details field must have additionalProperties with type string")
		}

		if detailsSchema.AdditionalProperties.IsA() {
			addlSchema := detailsSchema.AdditionalProperties.A.Schema()
			if addlSchema == nil || len(addlSchema.Type) == 0 || addlSchema.Type[0] != "string" {
				return fmt.Errorf("details field additionalProperties must have type string")
			}
		}
	}

	return nil
}
