package fieldmap

import (
	"fmt"

	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// Finding is one lock-consistency problem discovered by Check. The lint package
// maps it onto its Violation type so findings flow through the normal reporter.
type Finding struct {
	Check      string
	Location   string
	Message    string
	Suggestion string
}

// Check runs the lock-consistency checks against a committed lock and the current
// spec, returning structured findings. It is read-only and assumes lock is
// non-nil (an absent lock means the checks do not apply and the caller skips Check
// entirely). The checks are: completeness, uniqueness, no-reassignment,
// reserved-retention, the enum *_UNSPECIFIED=0 invariant, and structural validity.
func Check(doc *v3.Document, lock *Lock) []Finding {
	var findings []Finding

	if lock.Version != Version {
		findings = append(findings, Finding{
			Check:      "structural",
			Location:   "fieldmap.lock",
			Message:    fmt.Sprintf("unexpected lock version %d (expected %d)", lock.Version, Version),
			Suggestion: "regenerate the lock with 'duh generate'",
		})
	}

	messages, enums := specUnits(doc)

	for _, m := range messages {
		lm := lock.Messages[m.Name]
		if lm == nil {
			findings = append(findings, missingSection("message", m.Name))
			continue
		}
		findings = append(findings, checkUnit("message", m.Name, m.Fields, lm.Fields, lm.Reserved)...)
	}

	for _, e := range enums {
		le := lock.Enums[e.Name]
		if le == nil {
			findings = append(findings, missingSection("enum", e.Name))
			continue
		}
		findings = append(findings, checkUnit("enum", e.Name, e.Variants, le.Variants, le.Reserved)...)
		findings = append(findings, checkEnumInvariant(e.Name, le.Variants)...)
	}

	return findings
}

func missingSection(kind, name string) Finding {
	return Finding{
		Check:      "completeness",
		Location:   fmt.Sprintf("fieldmap.lock/%ss/%s", kind, name),
		Message:    fmt.Sprintf("spec %s %q has no entry in fieldmap.lock", kind, name),
		Suggestion: "run 'duh generate' to update the lock",
	}
}

// checkUnit validates one message/enum: numbers are unique across live and
// reserved entries (uniqueness + reserved retention), every spec name maps to a
// live entry (completeness), and no live entry exists for a name the spec dropped
// (no-reassignment / stale lock).
func checkUnit(kind, unit string, specNames []string, entries map[string]*Entry, reserved []int) []Finding {
	var findings []Finding
	location := fmt.Sprintf("fieldmap.lock/%ss/%s", kind, unit)
	member := memberWord(kind)

	seen := make(map[int]string, len(entries))
	for name, e := range entries {
		if other, dup := seen[e.Number]; dup {
			findings = append(findings, Finding{
				Check:      "uniqueness",
				Location:   location,
				Message:    fmt.Sprintf("number %d is used by both %q and %q", e.Number, other, name),
				Suggestion: "each number must be unique within a message/enum; a number is never reused, even after removal",
			})
		} else {
			seen[e.Number] = name
		}
	}

	// Orphaned reserved numbers must not collide with any live or tombstoned entry
	// (reserved retention + uniqueness across live + reserved).
	for _, n := range reserved {
		if other, dup := seen[n]; dup {
			findings = append(findings, Finding{
				Check:      "reserved",
				Location:   location,
				Message:    fmt.Sprintf("reserved number %d is also used by %q", n, other),
				Suggestion: "a reserved number is never reassigned to a live or tombstoned member",
			})
		} else {
			seen[n] = "<reserved>"
		}
	}

	inSpec := make(map[string]bool, len(specNames))
	for _, name := range specNames {
		inSpec[name] = true
	}

	for _, name := range specNames {
		e, ok := entries[name]
		if !ok {
			findings = append(findings, Finding{
				Check:      "completeness",
				Location:   location,
				Message:    fmt.Sprintf("spec %s %q has no entry in fieldmap.lock", member, name),
				Suggestion: "run 'duh generate' to add the entry",
			})
			continue
		}
		if e.Reserved {
			findings = append(findings, Finding{
				Check:      "reserved",
				Location:   location,
				Message:    fmt.Sprintf("%s %q is present in the spec but reserved in fieldmap.lock", member, name),
				Suggestion: "a reserved number is never reassigned; run 'duh generate' so the live field gets a new number",
			})
		}
	}

	for name, e := range entries {
		if !e.Reserved && !inSpec[name] {
			findings = append(findings, Finding{
				Check:      "no-reassignment",
				Location:   location,
				Message:    fmt.Sprintf("%s %q is live in fieldmap.lock but absent from the spec", member, name),
				Suggestion: "run 'duh generate' so the removed field is recorded as reserved",
			})
		}
	}

	return findings
}

// checkEnumInvariant enforces that an *_UNSPECIFIED sentinel maps to 0 and that no
// other variant occupies 0 (the proto wire default must stay distinguishable). The
// invariant only applies to enums that actually use a sentinel; a plain integer
// enum with no *_UNSPECIFIED variant legitimately numbers its first value 0.
func checkEnumInvariant(name string, variants map[string]*Entry) []Finding {
	hasSentinel := false
	for variant := range variants {
		if isUnspecifiedSentinel(variant) {
			hasSentinel = true
			break
		}
	}
	if !hasSentinel {
		return nil
	}

	var findings []Finding
	location := fmt.Sprintf("fieldmap.lock/enums/%s", name)

	for variant, e := range variants {
		sentinel := isUnspecifiedSentinel(variant)
		if sentinel && e.Number != 0 {
			findings = append(findings, Finding{
				Check:      "enum",
				Location:   location,
				Message:    fmt.Sprintf("variant %q must be 0 but is %d", variant, e.Number),
				Suggestion: "the *_UNSPECIFIED variant must map to proto 0 (the wire default)",
			})
		}
		if !sentinel && e.Number == 0 {
			findings = append(findings, Finding{
				Check:      "enum",
				Location:   location,
				Message:    fmt.Sprintf("variant %q occupies proto 0, which is reserved for the *_UNSPECIFIED sentinel", variant),
				Suggestion: "declare an *_UNSPECIFIED variant first so it owns proto 0",
			})
		}
	}

	return findings
}

func memberWord(kind string) string {
	if kind == "enum" {
		return "variant"
	}
	return "field"
}
