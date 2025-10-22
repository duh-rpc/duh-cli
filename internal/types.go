package internal

import (
	"fmt"

	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

// Violation represents a single DUH-RPC compliance violation
type Violation struct {
	Suggestion string
	RuleName   string
	Location   string
	Message    string
}

// String formats violation for display
func (v Violation) String() string {
	return fmt.Sprintf("[%s] %s\n  %s\n  %s", v.RuleName, v.Location, v.Message, v.Suggestion)
}

// ValidationResult contains all violations found in a document
type ValidationResult struct {
	Violations []Violation
	FilePath   string
}

// Valid returns true if no violations found
func (vr ValidationResult) Valid() bool {
	return len(vr.Violations) == 0
}

// Rule interface that all validation rules must implement
type Rule interface {
	Name() string
	Validate(doc *v3.Document) []Violation
}
