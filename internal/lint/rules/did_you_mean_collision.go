package rules

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pb33f/libopenapi/datamodel/high/v3"
	yaml "go.yaml.in/yaml/v4"
)

const didYouMeanExtension = "x-duh-did-you-mean"

// DidYouMeanCollisionRule enforces the switch-path uniqueness invariant on the
// x-duh-did-you-mean teaching extension: every entry must be a well-formed
// spec-relative path, must not equal a canonical route, and must be declared
// exactly once across the spec. It gives authors the same feedback 'duh generate'
// enforces, only earlier (ENG-133).
type DidYouMeanCollisionRule struct{}

func NewDidYouMeanCollisionRule() *DidYouMeanCollisionRule {
	return &DidYouMeanCollisionRule{}
}

func (r *DidYouMeanCollisionRule) Name() string {
	return "DID_YOU_MEAN_COLLISION"
}

func (r *DidYouMeanCollisionRule) Validate(doc *v3.Document) []Violation {
	var violations []Violation

	if doc == nil || doc.Paths == nil || doc.Paths.PathItems == nil {
		return violations
	}

	canonical := make(map[string]bool)
	for path := range doc.Paths.PathItems.FromOldest() {
		canonical[path] = true
	}

	declaredBy := make(map[string][]string)

	for path, item := range doc.Paths.PathItems.FromOldest() {
		if item == nil || item.Extensions == nil {
			continue
		}
		node, ok := item.Extensions.Get(didYouMeanExtension)
		if !ok || node == nil {
			continue
		}
		if node.Kind != yaml.SequenceNode {
			violations = append(violations, Violation{
				Suggestion: "Declare x-duh-did-you-mean as a list of paths, each beginning with '/'",
				Message:    "x-duh-did-you-mean must be a list of strings, each beginning with '/'",
				Location:   path,
				RuleName:   r.Name(),
				Severity:   SeverityError,
			})
			continue
		}
		for _, child := range node.Content {
			if child.Kind != yaml.ScalarNode || child.Tag != "!!str" {
				violations = append(violations, Violation{
					Suggestion: "Each x-duh-did-you-mean entry must be a string beginning with '/'",
					Message:    "x-duh-did-you-mean entry is malformed: each entry must be a string beginning with '/'",
					Location:   path,
					RuleName:   r.Name(),
					Severity:   SeverityError,
				})
				continue
			}
			entry := child.Value
			if !strings.HasPrefix(entry, "/") {
				violations = append(violations, Violation{
					Suggestion: "Each x-duh-did-you-mean entry must begin with '/'",
					Message:    fmt.Sprintf("x-duh-did-you-mean entry %q is malformed: must begin with '/'", entry),
					Location:   path,
					RuleName:   r.Name(),
					Severity:   SeverityError,
				})
				continue
			}
			if canonical[entry] {
				violations = append(violations, Violation{
					Suggestion: "A did-you-mean path may not equal a real route; remove the entry or rename the route",
					Message:    fmt.Sprintf("x-duh-did-you-mean path %q collides with the canonical route %q", entry, entry),
					Location:   path,
					RuleName:   r.Name(),
					Severity:   SeverityError,
				})
			}
			declaredBy[entry] = append(declaredBy[entry], path)
		}
	}

	for _, entry := range sortedKeys(declaredBy) {
		decls := declaredBy[entry]
		if len(decls) <= 1 {
			continue
		}
		// A within-list duplicate declares the same path item twice; name each
		// declaring path once so the message never reads "on X and X".
		seen := make(map[string]bool, len(decls))
		var quoted []string
		for _, d := range decls {
			if seen[d] {
				continue
			}
			seen[d] = true
			quoted = append(quoted, fmt.Sprintf("%q", d))
		}
		violations = append(violations, Violation{
			Suggestion: "Declare each did-you-mean path on exactly one operation",
			Message:    fmt.Sprintf("x-duh-did-you-mean path %q is declared more than once (on %s)", entry, strings.Join(quoted, " and ")),
			Location:   entry,
			RuleName:   r.Name(),
			Severity:   SeverityError,
		})
	}

	return violations
}

func sortedKeys(m map[string][]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
