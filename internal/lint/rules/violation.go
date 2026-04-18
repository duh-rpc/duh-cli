package rules

import "fmt"

// Severity represents the severity level of a violation
type Severity int

const (
	SeverityError   Severity = iota
	SeverityWarning
)

// String returns the string representation of the severity
func (s Severity) String() string {
	switch s {
	case SeverityWarning:
		return "WARNING"
	default:
		return "ERROR"
	}
}

// Violation represents a single DUH-RPC compliance violation
type Violation struct {
	Suggestion string
	RuleName   string
	Location   string
	Message    string
	Severity   Severity
}

// String formats violation for display
func (v Violation) String() string {
	return fmt.Sprintf("[%s] [%s] %s\n  %s\n  %s", v.Severity, v.RuleName, v.Location, v.Message, v.Suggestion)
}
