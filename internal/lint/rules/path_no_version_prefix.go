package rules

import (
	"regexp"

	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

var versionPrefixRegex = regexp.MustCompile(`^/v\d+/`)

// PathNoVersionPrefixRule flags paths with a version prefix like /v1/
type PathNoVersionPrefixRule struct{}

func NewPathNoVersionPrefixRule() *PathNoVersionPrefixRule {
	return &PathNoVersionPrefixRule{}
}

func (r *PathNoVersionPrefixRule) Name() string {
	return "PATH_NO_VERSION_PREFIX"
}

func (r *PathNoVersionPrefixRule) Validate(doc *v3.Document) []Violation {
	var violations []Violation

	if doc == nil || doc.Paths == nil || doc.Paths.PathItems == nil {
		return violations
	}

	for path := range doc.Paths.PathItems.FromOldest() {
		if versionPrefixRegex.MatchString(path) {
			violations = append(violations, Violation{
				Suggestion: "Remove version prefix from path; version belongs in servers[].url",
				Message:    "Path must not contain version prefix",
				Location:   path,
				RuleName:   r.Name(),
				Severity:   SeverityError,
			})
		}
	}

	return violations
}
