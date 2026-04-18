package rules

import (
	"strings"

	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

type NoPlainTextResponseRule struct{}

func NewNoPlainTextResponseRule() *NoPlainTextResponseRule {
	return &NoPlainTextResponseRule{}
}

func (r *NoPlainTextResponseRule) Name() string {
	return "NO_PLAIN_TEXT_RESPONSE"
}

func (r *NoPlainTextResponseRule) Validate(doc *v3.Document) []Violation {
	var violations []Violation

	if doc == nil || doc.Paths == nil || doc.Paths.PathItems == nil {
		return violations
	}

	for path, pathItem := range doc.Paths.PathItems.FromOldest() {
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

		for method, op := range operations {
			if op == nil {
				continue
			}

			if isOperationIgnored(op, r.Name()) {
				continue
			}

			if op.Responses == nil || op.Responses.Codes == nil {
				continue
			}

			for statusCode, response := range op.Responses.Codes.FromOldest() {
				if response == nil || response.Content == nil {
					continue
				}

				for contentType := range response.Content.FromOldest() {
					if strings.EqualFold(contentType, "text/plain") {
						violations = append(violations, Violation{
							Suggestion: "Use application/json or application/protobuf instead of text/plain",
							Message:    "Response must not use text/plain content type",
							Location:   method + " " + path + " response " + statusCode,
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
