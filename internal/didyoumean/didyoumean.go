// Package didyoumean holds the single validation of the x-duh-did-you-mean
// teaching extension shared by 'duh generate' and 'duh lint' (ENG-133). The two
// commands enforce the invariant independently, but from one implementation so
// they cannot drift.
package didyoumean

import (
	"fmt"
	"sort"
	"strings"

	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	yaml "go.yaml.in/yaml/v4"
)

// Extension is the path-item OpenAPI extension declaring the likely wrong-guess
// paths that 'duh generate' turns into teaching 404 routes.
const Extension = "x-duh-did-you-mean"

// Problem is one validation failure against the switch-path uniqueness and
// well-formedness invariant, ready to render as a generate error or a lint
// Violation. Entry is empty when the whole extension (not one entry) is at fault.
type Problem struct {
	Path       string
	Entry      string
	Message    string
	Suggestion string
}

// Node returns the raw extension sequence node on a path item, or (nil, false)
// when the extension is absent.
func Node(item *v3.PathItem) (*yaml.Node, bool) {
	if item == nil || item.Extensions == nil {
		return nil, false
	}
	node, ok := item.Extensions.Get(Extension)
	if !ok || node == nil {
		return nil, false
	}
	return node, true
}

// Resolve returns the full teaching route paths (base + declared) for a path
// item. It trusts Validate has already passed and silently skips anything
// malformed, so a template never emits a broken case. The well-formedness test
// matches Validate's, so a validated spec resolves exactly the entries it checked.
func Resolve(base string, item *v3.PathItem) []string {
	node, ok := Node(item)
	if !ok || node.Kind != yaml.SequenceNode {
		return nil
	}
	var out []string
	for _, child := range node.Content {
		if child.Kind == yaml.ScalarNode && child.Tag == "!!str" && wellFormed(child.Value) {
			out = append(out, base+child.Value)
		}
	}
	return out
}

// Validate enforces, across the whole spec: the extension sits beside a post:
// operation; every entry is a well-formed spec-relative path (a string, leading
// '/', no character that would break the generated switch); no entry equals a
// canonical route; and no entry is declared more than once. Problems are sorted
// by message for deterministic output. It returns nil when no path uses the
// extension, so a spec without it behaves exactly as before the feature.
func Validate(doc *v3.Document) []Problem {
	if doc == nil || doc.Paths == nil || doc.Paths.PathItems == nil {
		return nil
	}

	canonical := make(map[string]bool)
	for pair := orderedmap.First(doc.Paths.PathItems); pair != nil; pair = pair.Next() {
		canonical[pair.Key()] = true
	}

	var problems []Problem
	declaredBy := make(map[string][]string)

	for pair := orderedmap.First(doc.Paths.PathItems); pair != nil; pair = pair.Next() {
		path := pair.Key()
		item := pair.Value()
		node, ok := Node(item)
		if !ok {
			continue
		}
		// A teaching route stands in for a real operation, so the extension must sit
		// beside a post:; declared on a path item with none, its entries would be
		// silently dropped from the generated switch.
		if item.Post == nil {
			problems = append(problems, Problem{
				Path:       path,
				Message:    fmt.Sprintf("%s on path %q must sit beside a post: operation; this path item has none", Extension, path),
				Suggestion: "Move x-duh-did-you-mean to the path item whose operation it teaches, or add the post: operation",
			})
			continue
		}
		if node.Kind != yaml.SequenceNode {
			problems = append(problems, Problem{
				Path:       path,
				Message:    fmt.Sprintf("%s on path %q must be a list of strings, each beginning with '/'", Extension, path),
				Suggestion: "Declare x-duh-did-you-mean as a list of paths, each beginning with '/'",
			})
			continue
		}
		for _, child := range node.Content {
			if child.Kind != yaml.ScalarNode || child.Tag != "!!str" {
				problems = append(problems, Problem{
					Path:       path,
					Message:    fmt.Sprintf("%s entry on path %q is malformed: each entry must be a string beginning with '/'", Extension, path),
					Suggestion: "Each x-duh-did-you-mean entry must be a string beginning with '/'",
				})
				continue
			}
			entry := child.Value
			if !strings.HasPrefix(entry, "/") {
				problems = append(problems, Problem{
					Path:       path,
					Entry:      entry,
					Message:    fmt.Sprintf("%s entry %q on path %q is malformed: must begin with '/'", Extension, entry, path),
					Suggestion: "Each x-duh-did-you-mean entry must begin with '/'",
				})
				continue
			}
			// The entry is interpolated verbatim into a Go `case "..."` label, so a
			// character that cannot appear in a Go string literal must be rejected here
			// by name rather than break template rendering with an opaque error later.
			if bad := unsafeChar(entry); bad != "" {
				problems = append(problems, Problem{
					Path:       path,
					Entry:      entry,
					Message:    fmt.Sprintf("%s entry %q on path %q is malformed: contains %s, which is not allowed in a path", Extension, entry, path, bad),
					Suggestion: "Remove the illegal character; a did-you-mean entry is a plain URL path",
				})
				continue
			}
			if canonical[entry] {
				problems = append(problems, Problem{
					Path:       path,
					Entry:      entry,
					Message:    fmt.Sprintf("%s path %q on %q collides with the canonical route %q", Extension, entry, path, entry),
					Suggestion: "A did-you-mean path may not equal a real route; remove the entry or rename the route",
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
		problems = append(problems, Problem{
			Entry:      entry,
			Message:    fmt.Sprintf("%s path %q is declared more than once (on %s)", Extension, entry, strings.Join(quoted, " and ")),
			Suggestion: "Declare each did-you-mean path on exactly one operation",
		})
	}

	sort.Slice(problems, func(i, j int) bool { return problems[i].Message < problems[j].Message })
	return problems
}

// wellFormed reports whether an entry is a resolvable teaching path: leading '/'
// and free of characters that would break the generated switch.
func wellFormed(entry string) bool {
	return strings.HasPrefix(entry, "/") && unsafeChar(entry) == ""
}

// unsafeChar names the first character in entry that cannot appear in a Go
// string-literal case label, or "" when the entry is safe to emit.
func unsafeChar(entry string) string {
	for _, r := range entry {
		switch {
		case r == '"':
			return "a double-quote"
		case r == '\\':
			return "a backslash"
		case r == '\n' || r == '\r':
			return "a newline"
		case r < 0x20:
			return "a control character"
		}
	}
	return ""
}

func sortedKeys(m map[string][]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
