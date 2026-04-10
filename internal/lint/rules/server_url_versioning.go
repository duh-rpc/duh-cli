package rules

import (
	"regexp"

	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

var serverVersionSuffixRegex = regexp.MustCompile(`/v\d+$`)

// ServerURLVersioningRule ensures servers[].url ends with /v{N}
type ServerURLVersioningRule struct{}

func NewServerURLVersioningRule() *ServerURLVersioningRule {
	return &ServerURLVersioningRule{}
}

func (r *ServerURLVersioningRule) Name() string {
	return "SERVER_URL_VERSIONING"
}

func (r *ServerURLVersioningRule) Validate(doc *v3.Document) []Violation {
	var violations []Violation

	if doc == nil {
		return violations
	}

	if len(doc.Servers) == 0 {
		violations = append(violations, Violation{
			Suggestion: "Add version to server URL (e.g., https://api.example.com/v1)",
			Message:    "No servers defined; servers[].url must include version",
			Location:   "servers",
			RuleName:   r.Name(),
			Severity:   SeverityError,
		})
		return violations
	}

	for _, server := range doc.Servers {
		if server == nil {
			continue
		}
		if !serverVersionSuffixRegex.MatchString(server.URL) {
			violations = append(violations, Violation{
				Suggestion: "Add version to server URL (e.g., https://api.example.com/v1)",
				Message:    "Server URL must end with version path (e.g., /v1)",
				Location:   server.URL,
				RuleName:   r.Name(),
				Severity:   SeverityError,
			})
		}
	}

	return violations
}
