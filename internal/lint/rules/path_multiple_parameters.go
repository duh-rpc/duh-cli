package rules

import (
	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

// PathMultipleParametersRule flags paths with more than one path parameter
type PathMultipleParametersRule struct{}

func NewPathMultipleParametersRule() *PathMultipleParametersRule {
	return &PathMultipleParametersRule{}
}

func (r *PathMultipleParametersRule) Name() string {
	return "PATH_MULTIPLE_PARAMETERS"
}

func (r *PathMultipleParametersRule) Validate(doc *v3.Document) []Violation {
	var violations []Violation

	if doc == nil || doc.Paths == nil || doc.Paths.PathItems == nil {
		return violations
	}

	for path := range doc.Paths.PathItems.FromOldest() {
		matches := pathParamRegex.FindAllString(path, -1)
		if len(matches) > 1 {
			violations = append(violations, Violation{
				RuleName:   r.Name(),
				Location:   path,
				Message:    "Path contains multiple parameters; at most one path parameter is allowed",
				Suggestion: "Reduce to at most one path parameter per path",
				Severity:   SeverityError,
			})
		}
	}

	return violations
}
