package rules

import (
	"strings"

	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

// RPCResponseStandardNameRule validates that response $ref schemas follow naming conventions
type RPCResponseStandardNameRule struct{}

func NewRPCResponseStandardNameRule() *RPCResponseStandardNameRule {
	return &RPCResponseStandardNameRule{}
}

func (r *RPCResponseStandardNameRule) Name() string {
	return "RPC_RESPONSE_STANDARD_NAME"
}

func (r *RPCResponseStandardNameRule) Validate(doc *v3.Document) []Violation {
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

		op := pathItem.Post
		if op.Responses == nil || op.Responses.Codes == nil {
			continue
		}

		service, method := extractServiceMethod(path)
		methodResponse := method + "Response"
		serviceMethodResponse := service + method + "Response"

		for statusCode, response := range op.Responses.Codes.FromOldest() {
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

			schemaName := extractSchemaName(ref)

			if schemaName != methodResponse && schemaName != serviceMethodResponse {
				violations = append(violations, Violation{
					Suggestion: "Rename to '" + methodResponse + "' or '" + serviceMethodResponse + "'",
					Message:    "Response schema '" + schemaName + "' does not follow naming convention",
					Location:   "POST " + path + " response " + statusCode,
					RuleName:   r.Name(),
					Severity:   SeverityError,
				})
			}
		}
	}

	return violations
}
