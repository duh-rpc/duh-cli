package types

import "fmt"

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
