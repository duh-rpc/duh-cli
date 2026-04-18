package rules

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

var pathFormatRegex = regexp.MustCompile(`^/[a-z][a-z0-9_-]{0,49}\.[a-z][a-z0-9_-]{0,49}$`)
var pathParamRegex = regexp.MustCompile(`\{[^}]+\}`)

// PathFormatRule validates DUH-RPC path format
type PathFormatRule struct{}

// NewPathFormatRule creates a new path format rule
func NewPathFormatRule() *PathFormatRule {
	return &PathFormatRule{}
}

func (r *PathFormatRule) Name() string {
	return "PATH_FORMAT"
}

func (r *PathFormatRule) Validate(doc *v3.Document) []Violation {
	var violations []Violation

	if doc == nil || doc.Paths == nil || doc.Paths.PathItems == nil {
		return violations
	}

	for path, pathItem := range doc.Paths.PathItems.FromOldest() {
		if pathItem == nil {
			continue
		}

		// Check for path parameters in the path string
		if pathParamRegex.MatchString(path) {
			violations = append(violations, Violation{
				Suggestion: "Remove path parameters and use request body fields instead",
				Message:    "Path contains path parameters, which are not allowed in DUH-RPC",
				Location:   path,
				RuleName:   r.Name(),
				Severity:   SeverityError,
			})
			continue
		}

		// Check if path parameters are defined in PathItem
		if len(pathItem.Parameters) > 0 {
			for _, param := range pathItem.Parameters {
				if param != nil && param.In == "path" {
					violations = append(violations, Violation{
						Message:    fmt.Sprintf("Path parameter '%s' is not allowed in DUH-RPC", param.Name),
						Suggestion: "Move path parameters to request body fields",
						Location:   path,
						RuleName:   r.Name(),
						Severity:   SeverityError,
					})
				}
			}
		}

		// Check path format
		if !pathFormatRegex.MatchString(path) {
			violations = append(violations, Violation{
				Suggestion: r.generateSuggestion(path),
				Message:    r.generateErrorMessage(path),
				Location:   path,
				RuleName:   r.Name(),
				Severity:   SeverityError,
			})
		}
	}

	return violations
}

func (r *PathFormatRule) generateErrorMessage(path string) string {
	if strings.Contains(path, "{") || strings.Contains(path, "}") {
		return "Path contains parameters, which are not allowed in DUH-RPC"
	}

	// Strip leading slash for analysis
	trimmed := strings.TrimPrefix(path, "/")

	if !strings.Contains(trimmed, ".") {
		return "Path must have format /{resource}.{method} with a dot separator"
	}

	segments := strings.Split(trimmed, ".")
	if len(segments) > 2 {
		return "Path must have exactly one dot separating resource and method"
	}

	resource, method := segments[0], segments[1]

	if len(resource) > 0 && !regexp.MustCompile(`^[a-z]`).MatchString(resource) {
		return "Resource/Method must start with a lowercase letter"
	}
	if !regexp.MustCompile(`^[a-z][a-z0-9_-]{0,49}$`).MatchString(resource) {
		return "Resource/Method must contain only lowercase letters, numbers, hyphens, and underscores"
	}

	if len(method) > 0 && !regexp.MustCompile(`^[a-z]`).MatchString(method) {
		return "Resource/Method must start with a lowercase letter"
	}
	if !regexp.MustCompile(`^[a-z][a-z0-9_-]{0,49}$`).MatchString(method) {
		return "Resource/Method must contain only lowercase letters, numbers, hyphens, and underscores"
	}

	return "Path does not match format: /{resource}.{method}"
}

func (r *PathFormatRule) generateSuggestion(path string) string {
	if !strings.Contains(path, ".") {
		return "Use format /{resource}.{method} (e.g., /users.create)"
	}

	return "Ensure path follows format /{resource}.{method} with lowercase letters, numbers, hyphens, and underscores only (e.g., /users.create, /pets.list)"
}
