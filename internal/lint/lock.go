package lint

import (
	"github.com/duh-rpc/duh-cli/internal/fieldmap"
	"github.com/duh-rpc/duh-cli/internal/lint/rules"
	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

// LockRuleName is the rule name reported for fieldmap.lock consistency violations.
const LockRuleName = "FIELDMAP_LOCK"

// ValidateLock validates a committed fieldmap.lock against the spec. It is a
// parallel pass to the rule chain, not a Rule: it loads the lock and runs the
// fieldmap consistency checks, mapping findings onto Violations so they flow
// through the normal reporter.
//
// An absent lock is a first-class state: the lock checks simply do not apply and
// no violation (and no warning) is produced. A present-but-malformed lock is a
// structural violation.
func ValidateLock(doc *v3.Document, lockPath string) []Violation {
	lock, err := fieldmap.Load(lockPath)
	if err != nil {
		return []Violation{{
			RuleName:   LockRuleName,
			Severity:   rules.SeverityError,
			Location:   lockPath,
			Message:    err.Error(),
			Suggestion: "fix the malformed lock or regenerate it with 'duh generate'",
		}}
	}

	if lock == nil {
		return nil
	}

	var violations []Violation
	for _, f := range fieldmap.Check(doc, lock) {
		violations = append(violations, Violation{
			RuleName:   LockRuleName,
			Severity:   rules.SeverityError,
			Location:   f.Location,
			Message:    f.Message,
			Suggestion: f.Suggestion,
		})
	}
	return violations
}
