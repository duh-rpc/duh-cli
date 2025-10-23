package internal

import (
	"github.com/duh-rpc/duhrpc-lint/internal/types"
)

// Violation is an alias for types.Violation for backwards compatibility
type Violation = types.Violation

// ValidationResult contains all violations found in a document
type ValidationResult struct {
	Violations []Violation
	FilePath   string
}

// Valid returns true if no violations found
func (vr ValidationResult) Valid() bool {
	return len(vr.Violations) == 0
}
