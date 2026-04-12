package rules

import (
	"fmt"
	"strings"

	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

// PathHyphenSeparatorRule flags path segments containing underscores or camelCase
type PathHyphenSeparatorRule struct{}

func NewPathHyphenSeparatorRule() *PathHyphenSeparatorRule {
	return &PathHyphenSeparatorRule{}
}

func (r *PathHyphenSeparatorRule) Name() string {
	return "PATH_HYPHEN_SEPARATOR"
}

func (r *PathHyphenSeparatorRule) Validate(doc *v3.Document) []Violation {
	var violations []Violation

	if doc == nil || doc.Paths == nil || doc.Paths.PathItems == nil {
		return violations
	}

	for path := range doc.Paths.PathItems.FromOldest() {
		trimmed := strings.TrimPrefix(path, "/")

		for segment := range strings.SplitSeq(trimmed, ".") {
			if strings.Contains(segment, "{") {
				continue
			}

			if strings.Contains(segment, "_") {
				violations = append(violations, Violation{
					RuleName:   r.Name(),
					Location:   path,
					Message:    fmt.Sprintf("Path segment '%s' uses underscores; multi-word segments must use hyphens", segment),
					Suggestion: "Use hyphens to separate words (e.g., /user-accounts.create)",
					Severity:   SeverityError,
				})
			}

			if segment != strings.ToLower(segment) {
				violations = append(violations, Violation{
					RuleName:   r.Name(),
					Location:   path,
					Message:    fmt.Sprintf("Path segment '%s' uses camelCase; multi-word segments must use hyphens", segment),
					Suggestion: "Use hyphens to separate words (e.g., /user-accounts.create)",
					Severity:   SeverityError,
				})
			}
		}
	}

	return violations
}
