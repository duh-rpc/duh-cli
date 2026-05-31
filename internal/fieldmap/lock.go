// Package fieldmap owns fieldmap.lock: the checked-in artifact that pins each
// message's JSON-field-name → proto-field-number mapping and each proto enum's
// variant → number mapping. duh generate writes it (append-only); duh lint
// validates it. The package imports neither lint nor generate so both can import
// it without a cycle.
package fieldmap

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"go.yaml.in/yaml/v4"
)

// Version is the current lock file schema version.
const Version = 1

// LockFileName is the conventional lock filename, co-located with the spec.
const LockFileName = "fieldmap.lock"

// DefaultLockPath returns the conventional lock location: fieldmap.lock next to
// the spec. Both duh generate and duh lint resolve an unset --lock-path this way.
func DefaultLockPath(specPath string) string {
	return filepath.Join(filepath.Dir(specPath), LockFileName)
}

// Lock is the in-memory model of fieldmap.lock.
type Lock struct {
	Version  int
	Messages map[string]*Message // key: OpenAPI component schema name
	Enums    map[string]*Enum    // key: OpenAPI component schema name
}

// Message holds the field-name → number bindings for one proto message. Reserved
// holds orphaned reserved numbers: numbers retained forever (ADR 0002) whose
// original field name has since been re-added live, so the name now keys a live
// entry and can no longer carry the tombstone. Normal removals stay as named
// tombstones (Fields entries with Reserved=true); this bare list is only for the
// re-added-same-name case.
type Message struct {
	Fields   map[string]*Entry // key: JSON field name
	Reserved []int
}

// Enum holds the variant-value → number bindings for one proto enum. Reserved holds
// orphaned reserved numbers (see Message.Reserved).
type Enum struct {
	Variants map[string]*Entry // key: literal OpenAPI enum value
	Reserved []int
}

// Entry is one name→number binding. Reserved marks a tombstone: the name is gone
// from the spec but its number is retained forever (ADR 0002).
type Entry struct {
	Number   int
	Reserved bool
}

// lockFile is the on-disk YAML shape used for parsing.
type lockFile struct {
	Version  int                    `yaml:"version"`
	Messages map[string]messageFile `yaml:"messages"`
	Enums    map[string]enumFile    `yaml:"enums"`
}

type messageFile struct {
	Fields   map[string]entryFile `yaml:"fields"`
	Reserved []int                `yaml:"reserved,omitempty"`
}

type enumFile struct {
	Variants map[string]entryFile `yaml:"variants"`
	Reserved []int                `yaml:"reserved,omitempty"`
}

type entryFile struct {
	Number   int  `yaml:"number"`
	Reserved bool `yaml:"reserved,omitempty"`
}

// Load parses a lock file. Returns (nil, nil) when the path does not exist (an
// absent lock is a valid state, not an error). A present-but-malformed file
// returns an error so callers can report a structural violation.
func Load(path string) (*Lock, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read lock file: %w", err)
	}

	var file lockFile
	if err := yaml.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("malformed fieldmap.lock: %w", err)
	}

	lock := &Lock{
		Version:  file.Version,
		Messages: make(map[string]*Message, len(file.Messages)),
		Enums:    make(map[string]*Enum, len(file.Enums)),
	}

	for name, m := range file.Messages {
		fields := make(map[string]*Entry, len(m.Fields))
		for field, e := range m.Fields {
			fields[field] = &Entry{Number: e.Number, Reserved: e.Reserved}
		}
		lock.Messages[name] = &Message{Fields: fields, Reserved: m.Reserved}
	}

	for name, e := range file.Enums {
		variants := make(map[string]*Entry, len(e.Variants))
		for variant, v := range e.Variants {
			variants[variant] = &Entry{Number: v.Number, Reserved: v.Reserved}
		}
		lock.Enums[name] = &Enum{Variants: variants, Reserved: e.Reserved}
	}

	return lock, nil
}

// Save serializes the lock deterministically: version, then messages and enums by
// ascending name, with fields and variants ordered by ascending number. Output is
// timestamp-free so an unchanged mapping is byte-identical regardless of spec field
// order. Only duh generate calls Save.
func (l *Lock) Save(path string) error {
	root := &yaml.Node{Kind: yaml.MappingNode}
	root.Content = append(root.Content, scalar("version", "!!str"), intNode(l.Version))

	if len(l.Messages) > 0 {
		root.Content = append(root.Content, scalar("messages", "!!str"), messagesNode(l.Messages))
	}
	if len(l.Enums) > 0 {
		root.Content = append(root.Content, scalar("enums", "!!str"), enumsNode(l.Enums))
	}

	out, err := yaml.Marshal(root)
	if err != nil {
		return fmt.Errorf("failed to serialize lock: %w", err)
	}

	return os.WriteFile(path, out, 0644)
}

func messagesNode(messages map[string]*Message) *yaml.Node {
	node := &yaml.Node{Kind: yaml.MappingNode}
	for _, name := range sortedKeys(messages) {
		fields := &yaml.Node{Kind: yaml.MappingNode}
		for _, key := range entriesByNumber(messages[name].Fields) {
			fields.Content = append(fields.Content, scalar(key, "!!str"), entryNode(messages[name].Fields[key]))
		}
		body := &yaml.Node{Kind: yaml.MappingNode}
		body.Content = append(body.Content, scalar("fields", "!!str"), fields)
		if seq := reservedNode(messages[name].Reserved); seq != nil {
			body.Content = append(body.Content, scalar("reserved", "!!str"), seq)
		}
		node.Content = append(node.Content, scalar(name, "!!str"), body)
	}
	return node
}

func enumsNode(enums map[string]*Enum) *yaml.Node {
	node := &yaml.Node{Kind: yaml.MappingNode}
	for _, name := range sortedKeys(enums) {
		variants := &yaml.Node{Kind: yaml.MappingNode}
		for _, key := range entriesByNumber(enums[name].Variants) {
			variants.Content = append(variants.Content, scalar(key, "!!str"), entryNode(enums[name].Variants[key]))
		}
		body := &yaml.Node{Kind: yaml.MappingNode}
		body.Content = append(body.Content, scalar("variants", "!!str"), variants)
		if seq := reservedNode(enums[name].Reserved); seq != nil {
			body.Content = append(body.Content, scalar("reserved", "!!str"), seq)
		}
		node.Content = append(node.Content, scalar(name, "!!str"), body)
	}
	return node
}

// reservedNode renders an ascending inline `[1, 2]` sequence of orphaned reserved
// numbers, or nil when there are none.
func reservedNode(numbers []int) *yaml.Node {
	if len(numbers) == 0 {
		return nil
	}
	sorted := append([]int(nil), numbers...)
	sort.Ints(sorted)

	seq := &yaml.Node{Kind: yaml.SequenceNode, Style: yaml.FlowStyle}
	for _, n := range sorted {
		seq.Content = append(seq.Content, intNode(n))
	}
	return seq
}

// entryNode renders a single binding as an inline flow mapping: {number: N} or
// {number: N, reserved: true}.
func entryNode(e *Entry) *yaml.Node {
	node := &yaml.Node{Kind: yaml.MappingNode, Style: yaml.FlowStyle}
	node.Content = append(node.Content, scalar("number", "!!str"), intNode(e.Number))
	if e.Reserved {
		node.Content = append(node.Content, scalar("reserved", "!!str"),
			&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: "true"})
	}
	return node
}

func scalar(value, tag string) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: tag, Value: value}
}

func intNode(n int) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!int", Value: strconv.Itoa(n)}
}

// entriesByNumber returns entry keys ordered by ascending number, then name, so
// the serialized order is stable and groups tombstones with their numbers.
func entriesByNumber(entries map[string]*Entry) []string {
	keys := make([]string, 0, len(entries))
	for key := range entries {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		if entries[keys[i]].Number != entries[keys[j]].Number {
			return entries[keys[i]].Number < entries[keys[j]].Number
		}
		return keys[i] < keys[j]
	})
	return keys
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
