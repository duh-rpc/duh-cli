package rules

import (
	"strings"

	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

// RPCRequestStandardNameRule validates that request body $ref schemas follow naming conventions
type RPCRequestStandardNameRule struct{}

func NewRPCRequestStandardNameRule() *RPCRequestStandardNameRule {
	return &RPCRequestStandardNameRule{}
}

func (r *RPCRequestStandardNameRule) Name() string {
	return "REQUEST_STANDARD_NAME"
}

func (r *RPCRequestStandardNameRule) Validate(doc *v3.Document) []Violation {
	var violations []Violation

	if doc == nil || doc.Paths == nil || doc.Paths.PathItems == nil {
		return violations
	}

	for path, pathItem := range doc.Paths.PathItems.FromOldest() {
		if !strings.Contains(path, ".") {
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

		ref := jsonContent.Schema.GetReference()
		if ref == "" {
			continue
		}

		schemaName := extractSchemaName(ref)
		service, method := extractServiceMethod(path)

		methodRequest := method + "Request"
		serviceMethodRequest := service + method + "Request"

		if schemaName != methodRequest && schemaName != serviceMethodRequest {
			violations = append(violations, Violation{
				Suggestion: "Rename to '" + methodRequest + "' or '" + serviceMethodRequest + "'",
				Message:    "Request schema '" + schemaName + "' does not follow naming convention",
				Location:   "POST " + path,
				RuleName:   r.Name(),
				Severity:   SeverityError,
			})
		}
	}

	return violations
}

// extractSchemaName extracts the schema name from a $ref string
func extractSchemaName(ref string) string {
	parts := strings.Split(ref, "/")
	return parts[len(parts)-1]
}

// extractServiceMethod extracts PascalCase service and method from a path like /user-accounts.get-by-id
func extractServiceMethod(path string) (string, string) {
	trimmed := strings.TrimPrefix(path, "/")
	dotIndex := strings.Index(trimmed, ".")
	if dotIndex < 0 {
		return "", toPascalCase(trimmed)
	}
	return toPascalCase(trimmed[:dotIndex]), toPascalCase(trimmed[dotIndex+1:])
}

// toPascalCase converts a kebab-case string to PascalCase
func toPascalCase(s string) string {
	parts := strings.Split(s, "-")
	var result strings.Builder
	for _, part := range parts {
		if len(part) == 0 {
			continue
		}
		result.WriteString(strings.ToUpper(part[:1]) + part[1:])
	}
	return result.String()
}
