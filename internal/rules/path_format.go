package rules

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/duh-rpc/duhrpc-lint/internal/types"
	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

var pathFormatRegex = regexp.MustCompile(`^/v(0|[1-9][0-9]*)/[a-z][a-z0-9_-]{0,49}\.[a-z][a-z0-9_-]{0,49}$`)
var pathParamRegex = regexp.MustCompile(`\{[^}]+\}`)

// PathFormatRule validates DUH-RPC path format
type PathFormatRule struct{}

// NewPathFormatRule creates a new path format rule
func NewPathFormatRule() *PathFormatRule {
	return &PathFormatRule{}
}

func (r *PathFormatRule) Name() string {
	return "path-format"
}

func (r *PathFormatRule) Validate(doc *v3.Document) []types.Violation {
	var violations []types.Violation

	if doc == nil || doc.Paths == nil || doc.Paths.PathItems == nil {
		return violations
	}

	for path, pathItem := range doc.Paths.PathItems.FromOldest() {
		if pathItem == nil {
			continue
		}

		// Check for path parameters in the path string
		if pathParamRegex.MatchString(path) {
			violations = append(violations, types.Violation{
				RuleName:   r.Name(),
				Location:   path,
				Message:    "Path contains path parameters, which are not allowed in DUH-RPC",
				Suggestion: "Remove path parameters and use request body fields instead",
			})
			continue
		}

		// Check if path parameters are defined in PathItem
		if len(pathItem.Parameters) > 0 {
			for _, param := range pathItem.Parameters {
				if param != nil && param.In == "path" {
					violations = append(violations, types.Violation{
						RuleName:   r.Name(),
						Location:   path,
						Message:    fmt.Sprintf("Path parameter '%s' is not allowed in DUH-RPC", param.Name),
						Suggestion: "Move path parameters to request body fields",
					})
				}
			}
		}

		// Check path format
		if !pathFormatRegex.MatchString(path) {
			violations = append(violations, types.Violation{
				RuleName:   r.Name(),
				Location:   path,
				Message:    r.generateErrorMessage(path),
				Suggestion: r.generateSuggestion(path),
			})
		}
	}

	return violations
}

func (r *PathFormatRule) generateErrorMessage(path string) string {
	if !strings.HasPrefix(path, "/v") {
		return "Path must start with version prefix (e.g., /v1/)"
	}

	if strings.Contains(path, "{") || strings.Contains(path, "}") {
		return "Path contains parameters, which are not allowed in DUH-RPC"
	}

	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) < 2 {
		return "Path must have format /v{N}/subject.method"
	}

	// Check version format
	versionPart := parts[0]
	if !regexp.MustCompile(`^v(0|[1-9][0-9]*)$`).MatchString(versionPart) {
		return "Version must be /v{N}/ where N is a non-negative integer (e.g., /v1/, /v2/)"
	}

	// Check subject.method format
	if len(parts) < 2 {
		return "Path must include subject.method after version"
	}

	methodPart := parts[1]
	if !strings.Contains(methodPart, ".") {
		return "Path must have format /v{N}/subject.method with a dot separator"
	}

	segments := strings.Split(methodPart, ".")
	if len(segments) != 2 {
		return "Path must have exactly one dot separating subject and method"
	}

	subject, method := segments[0], segments[1]

	// Check subject format
	if !regexp.MustCompile(`^[a-z][a-z0-9_-]{0,49}$`).MatchString(subject) {
		if len(subject) > 50 {
			return "Subject segment exceeds 50 characters"
		}
		if len(subject) > 0 && !regexp.MustCompile(`^[a-z]`).MatchString(subject) {
			return "Subject must start with a lowercase letter"
		}
		return "Subject must contain only lowercase letters, numbers, hyphens, and underscores"
	}

	// Check method format
	if !regexp.MustCompile(`^[a-z][a-z0-9_-]{0,49}$`).MatchString(method) {
		if len(method) > 50 {
			return "Method segment exceeds 50 characters"
		}
		if len(method) > 0 && !regexp.MustCompile(`^[a-z]`).MatchString(method) {
			return "Method must start with a lowercase letter"
		}
		return "Method must contain only lowercase letters, numbers, hyphens, and underscores"
	}

	return "Path does not match DUH-RPC format: /v{N}/subject.method"
}

func (r *PathFormatRule) generateSuggestion(path string) string {
	if !strings.HasPrefix(path, "/v") {
		return "Add a version prefix like /v1/"
	}

	if !strings.Contains(path, ".") {
		return "Use format /v1/subject.method (e.g., /v1/users.create)"
	}

	return "Ensure path follows format /v{N}/subject.method with lowercase letters, numbers, hyphens, and underscores only"
}
