package internal

import (
	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

// Validate runs all registered rules against the document
func Validate(doc *v3.Document, filePath string) ValidationResult {
	var rules []Rule

	var violations []Violation
	for _, rule := range rules {
		ruleViolations := rule.Validate(doc)
		violations = append(violations, ruleViolations...)
	}

	return ValidationResult{
		Violations: violations,
		FilePath:   filePath,
	}
}
