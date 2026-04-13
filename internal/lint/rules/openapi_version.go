package rules

import (
	"fmt"
	"strings"

	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

type OpenAPIVersionRule struct{}

func NewOpenAPIVersionRule() *OpenAPIVersionRule {
	return &OpenAPIVersionRule{}
}

func (r *OpenAPIVersionRule) Name() string {
	return "OPENAPI_VERSION"
}

func (r *OpenAPIVersionRule) Validate(doc *v3.Document) []Violation {
	var violations []Violation

	if doc == nil {
		return violations
	}

	version := doc.Version
	if version == "" || !strings.HasPrefix(version, "3.") {
		display := version
		if display == "" {
			display = "empty"
		}
		violations = append(violations, Violation{
			Suggestion: "Set openapi field to a 3.x version (e.g., 3.0.3)",
			Message:    fmt.Sprintf("OpenAPI version must be 3.x (found: '%s')", display),
			Location:   "openapi",
			RuleName:   r.Name(),
			Severity:   SeverityError,
		})
	}

	return violations
}
