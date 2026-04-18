package rules

import (
	"fmt"
	"strings"

	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

// PathPluralResourcesRule flags resource names not ending in "s"
type PathPluralResourcesRule struct{}

func NewPathPluralResourcesRule() *PathPluralResourcesRule {
	return &PathPluralResourcesRule{}
}

func (r *PathPluralResourcesRule) Name() string {
	return "PATH_PLURAL_RESOURCES"
}

func (r *PathPluralResourcesRule) Validate(doc *v3.Document) []Violation {
	var violations []Violation

	if doc == nil || doc.Paths == nil || doc.Paths.PathItems == nil {
		return violations
	}

	for path := range doc.Paths.PathItems.FromOldest() {
		trimmed := strings.TrimPrefix(path, "/")
		segments := strings.Split(trimmed, ".")
		resource := segments[0]

		if !strings.HasSuffix(resource, "s") {
			violations = append(violations, Violation{
				RuleName:   r.Name(),
				Location:   path,
				Message:    fmt.Sprintf("Resource name '%s' should be plural", resource),
				Suggestion: fmt.Sprintf("Use plural nouns for resource names (e.g., /%ss.create instead of /%s.create)", resource, resource),
				Severity:   SeverityWarning,
			})
		}
	}

	return violations
}
