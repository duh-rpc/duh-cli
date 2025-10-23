package lint

import (
	rules2 "github.com/duh-rpc/duhrpc-lint/internal/lint/rules"
	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

// Rule interface that all validation rules must implement
type Rule interface {
	Name() string
	Validate(doc *v3.Document) []Violation
}

// Validate runs all registered rules against the document
func Validate(doc *v3.Document, filePath string) ValidationResult {
	allRules := []Rule{
		rules2.NewPathFormatRule(),
		rules2.NewHTTPMethodRule(),
		rules2.NewQueryParamsRule(),
		rules2.NewRequestBodyRule(),
		rules2.NewStatusCodeRule(),
		rules2.NewSuccessResponseRule(),
		rules2.NewContentTypeRule(),
		rules2.NewErrorResponseRule(),
	}

	var violations []Violation
	for _, rule := range allRules {
		ruleViolations := rule.Validate(doc)
		violations = append(violations, ruleViolations...)
	}

	return ValidationResult{
		Violations: violations,
		FilePath:   filePath,
	}
}
