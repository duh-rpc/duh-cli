package lint

import (
	"github.com/duh-rpc/duh-cli/internal/lint/rules"
)

// Violation is an alias for types.Violation for backwards compatibility
type Violation = rules.Violation

// ValidationResult contains all violations found in a document
type ValidationResult struct {
	Violations []Violation
	FilePath   string
}

// Valid returns true if no ERROR-severity violations exist
func (vr ValidationResult) Valid() bool {
	return vr.ErrorCount() == 0
}

// ErrorCount returns the number of ERROR-severity violations
func (vr ValidationResult) ErrorCount() int {
	var count int
	for _, v := range vr.Violations {
		if v.Severity == rules.SeverityError {
			count++
		}
	}
	return count
}

// WarningCount returns the number of WARNING-severity violations
func (vr ValidationResult) WarningCount() int {
	var count int
	for _, v := range vr.Violations {
		if v.Severity == rules.SeverityWarning {
			count++
		}
	}
	return count
}
