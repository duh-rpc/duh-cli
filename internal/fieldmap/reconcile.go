package fieldmap

import (
	"fmt"
	"strings"

	schema "github.com/duh-rpc/openapi-schema.go"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// messageBase is the first proto field number for a message; enumBase is the first
// number for a proto enum (the zero value, reserved for the *_UNSPECIFIED sentinel).
const (
	messageBase = 1
	enumBase    = 0
)

// Reconcile produces the next lock and the library FieldNumbers from the current
// spec and the existing lock (nil = first-run seeding). For each message/enum it
// copies every existing live and reserved entry verbatim, flips removed live
// entries to reserved tombstones, and assigns each new field/variant the next
// high-water-mark number in OpenAPI declaration order. Seeding is not a distinct
// path: a wholly new unit simply has an empty prior set.
//
// Reconcile is pure — it writes nothing — and returns an error with no partial
// state on a corrupt lock (a number used twice) or a seeding-precondition
// violation, so the caller can fail loud and write nothing.
func Reconcile(doc *v3.Document, existing *Lock) (*Lock, *schema.FieldNumbers, error) {
	messages, enums := specUnits(doc)

	next := &Lock{
		Version:  Version,
		Messages: make(map[string]*Message, len(messages)),
		Enums:    make(map[string]*Enum, len(enums)),
	}
	nums := &schema.FieldNumbers{
		Messages: make(map[string]schema.MessageNumbers, len(messages)),
		Enums:    make(map[string]schema.EnumNumbers, len(enums)),
	}

	for _, m := range messages {
		prior, priorReserved := priorMessage(existing, m.Name)
		if err := assertNoDuplicateNumbers(prior, priorReserved, "message", m.Name); err != nil {
			return nil, nil, err
		}
		entries, reserved := reconcileEntries(prior, priorReserved, m.Fields, messageBase)
		next.Messages[m.Name] = &Message{Fields: entries, Reserved: reserved}
		nums.Messages[m.Name] = toMessageNumbers(entries, reserved)
	}

	for _, e := range enums {
		prior, priorReserved := priorEnum(existing, e.Name)
		if err := assertNoDuplicateNumbers(prior, priorReserved, "enum", e.Name); err != nil {
			return nil, nil, err
		}
		if prior == nil {
			if err := assertUnspecifiedFirst(e); err != nil {
				return nil, nil, err
			}
		}
		entries, reserved := reconcileEntries(prior, priorReserved, e.Variants, enumBase)
		next.Enums[e.Name] = &Enum{Variants: entries, Reserved: reserved}
		nums.Enums[e.Name] = toEnumNumbers(entries, reserved)
	}

	return next, nums, nil
}

// reconcileEntries applies the one uniform append-only rule: keep every prior live
// and reserved entry, tombstone a prior live entry whose name left the spec, and
// assign each spec name still needing a number the next high-water-mark value in
// declaration order. base is the unit's first number (1 for messages, 0 for enums).
// It returns the next entry set and the orphaned reserved numbers (carried from the
// prior orphaned list plus any tombstone whose name was re-added live this run).
func reconcileEntries(prior map[string]*Entry, priorReserved []int, ordered []string, base int) (map[string]*Entry, []int) {
	inSpec := make(map[string]bool, len(ordered))
	for _, name := range ordered {
		inSpec[name] = true
	}

	result := make(map[string]*Entry, len(prior)+len(ordered))
	highWater := base - 1

	// Orphaned reserved numbers persist forever; carry the prior list verbatim.
	orphaned := append([]int(nil), priorReserved...)
	for _, n := range priorReserved {
		if n > highWater {
			highWater = n
		}
	}

	for name, e := range prior {
		if e.Number > highWater {
			highWater = e.Number
		}
		switch {
		case e.Reserved:
			result[name] = &Entry{Number: e.Number, Reserved: true}
		case inSpec[name]:
			result[name] = &Entry{Number: e.Number}
		default:
			// Removed from the spec → reserve the number forever (ADR 0002).
			result[name] = &Entry{Number: e.Number, Reserved: true}
		}
	}

	// Assign numbers to spec names that have no live entry yet, in declaration
	// order so multiple new fields in one run are deterministic. Re-adding a
	// previously-removed name takes a fresh high-water number; its old number can
	// no longer be a named tombstone (the name is live again) so it moves to the
	// orphaned reserved list and is never reused.
	for _, name := range ordered {
		if e, ok := result[name]; ok && !e.Reserved {
			continue
		}
		if e, ok := result[name]; ok && e.Reserved {
			orphaned = append(orphaned, e.Number)
		}
		highWater++
		result[name] = &Entry{Number: highWater}
	}

	return result, orphaned
}

// assertNoDuplicateNumbers fails loud when a committed lock unit already maps two
// names (or a name and an orphaned reserved number) to the same number — an
// impossible append-only state, reachable only via a corrupt or hand-mangled lock.
func assertNoDuplicateNumbers(prior map[string]*Entry, priorReserved []int, kind, name string) error {
	seen := make(map[int]string, len(prior))
	for entryName, e := range prior {
		if other, dup := seen[e.Number]; dup {
			return fmt.Errorf("%s %q: number %d is mapped by both %q and %q in fieldmap.lock; "+
				"refusing to reassign a published number (no files written)", kind, name, e.Number, other, entryName)
		}
		seen[e.Number] = entryName
	}
	for _, n := range priorReserved {
		if other, dup := seen[n]; dup {
			return fmt.Errorf("%s %q: reserved number %d is also mapped by %q in fieldmap.lock; "+
				"refusing to reassign a published number (no files written)", kind, name, n, other)
		}
		seen[n] = "<reserved>"
	}
	return nil
}

// assertUnspecifiedFirst enforces the ADR 0004 seeding precondition for any enum
// that uses an *_UNSPECIFIED sentinel: the sentinel must be declared first so
// declaration-order seeding assigns it 0. Enums with no UNSPECIFIED variant (e.g.
// plain integer enums) are unaffected.
func assertUnspecifiedFirst(e enumSpec) error {
	hasSentinel := false
	for _, v := range e.Variants {
		if strings.HasSuffix(v, "UNSPECIFIED") {
			hasSentinel = true
			break
		}
	}
	if !hasSentinel {
		return nil
	}
	if len(e.Variants) == 0 || !strings.HasSuffix(e.Variants[0], "UNSPECIFIED") {
		first := ""
		if len(e.Variants) > 0 {
			first = e.Variants[0]
		}
		return fmt.Errorf("enum %q: first declared variant %q is not an *_UNSPECIFIED variant; "+
			"reorder it first (see ADR 0004) — no files written", e.Name, first)
	}
	return nil
}

func priorMessage(existing *Lock, name string) (map[string]*Entry, []int) {
	if existing == nil || existing.Messages[name] == nil {
		return nil, nil
	}
	return existing.Messages[name].Fields, existing.Messages[name].Reserved
}

func priorEnum(existing *Lock, name string) (map[string]*Entry, []int) {
	if existing == nil || existing.Enums[name] == nil {
		return nil, nil
	}
	return existing.Enums[name].Variants, existing.Enums[name].Reserved
}

func toMessageNumbers(entries map[string]*Entry, orphaned []int) schema.MessageNumbers {
	fields := make(map[string]int)
	reserved := append([]int(nil), orphaned...)
	for name, e := range entries {
		if e.Reserved {
			reserved = append(reserved, e.Number)
		} else {
			fields[name] = e.Number
		}
	}
	return schema.MessageNumbers{Fields: fields, Reserved: reserved}
}

func toEnumNumbers(entries map[string]*Entry, orphaned []int) schema.EnumNumbers {
	variants := make(map[string]int)
	reserved := append([]int(nil), orphaned...)
	for name, e := range entries {
		if e.Reserved {
			reserved = append(reserved, e.Number)
		} else {
			variants[name] = e.Number
		}
	}
	return schema.EnumNumbers{Variants: variants, Reserved: reserved}
}
