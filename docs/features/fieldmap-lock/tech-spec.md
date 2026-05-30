# `fieldmap.lock` — Stable Proto Field Numbers Tech Spec

_PRD: [docs/features/fieldmap-lock/prd.md](./prd.md)_
_ADRs: [0002](../../adr/0002-reserved-proto-field-numbers-are-permanent.md), [0003](../../adr/0003-fieldmap-lock-is-the-checked-in-field-number-record.md), [0004](../../adr/0004-duh-generate-requires-unspecified-as-first-enum-variant.md), [0005](../../adr/0005-field-numbers-injected-via-structured-library-option.md), [0006](../../adr/0006-removed-fields-recorded-as-tombstone-entries.md), [0007](../../adr/0007-enum-variants-keyed-by-literal-openapi-value.md)_

## Overview

`duh generate` assigns Protobuf field numbers positionally, so a mid-schema insert or reorder
silently renumbers every later field and corrupts the proto wire. This feature introduces a
checked-in `fieldmap.lock` that pins JSON-field-name → proto-field-number per message and
literal-enum-value → proto-integer per enum. `duh generate` writes and append-only maintains the
lock; `duh lint` validates a committed lock against the spec as a CI gate.

The pivotal implementation fact: **field numbering lives inside the proto-conversion library, not
in duh-cli.** Today duh-cli calls `github.com/duh-rpc/openapi-proto.go` v0.2.0, which numbers
positionally inside a non-importable `internal` package and exposes no injection point. That
library has been renamed to `github.com/duh-rpc/openapi-schema.go` (the old import path no longer
exists). This feature therefore does three things at once:

1. **Migrates duh-cli onto `openapi-schema.go`** and its `*ConvertResult` API.
2. **Extends `openapi-schema.go`** with a structured, programmatic field-number injection point
   (`ConvertOptions.FieldNumbers`), `reserved` rendering, and corrected enum numbering.
3. **Adds the lock** — a new `internal/fieldmap` package owning the lock file model and
   reconciliation, wired into `generate` (write path) and `lint` (validate path).

## Component Design

```
duh generate <spec>                         duh lint <spec>
  │                                            │
  ├─ lint.Validate(doc, …)  (spec rules only)  ├─ lint.Validate(doc, …)  (spec rules)
  │                                            ├─ lint.ValidateLock(doc, lockPath)  ← NEW
  ├─ fieldmap.Reconcile(doc, existingLock) ←NEW│      └─ uses fieldmap model + checks
  │      → resolved lock + FieldNumbers        │
  ├─ converter.Convert(spec, …, FieldNumbers)  │
  │      → proto (locked numbers + reserved)   │
  └─ write proto + code + lock  (atomic)        (read-only; never writes the lock)
```

### `internal/fieldmap` (new package)

Owns the lock file model and all reconciliation logic. Imports neither `lint` nor `generate`
(both import it), avoiding the existing `generate → lint` cycle.

```go
package fieldmap

// Lock is the in-memory model of fieldmap.lock.
type Lock struct {
    Version  int                 // file schema version, currently 1
    Messages map[string]*Message // key: OpenAPI component schema name
    Enums    map[string]*Enum    // key: OpenAPI component schema name
}

type Message struct {
    Fields map[string]*Entry // key: JSON field name
}

type Enum struct {
    Variants map[string]*Entry // key: literal OpenAPI enum value
}

// Entry is one name→number binding. Reserved marks a tombstone: the name is gone
// from the spec but its number is retained forever (ADR 0002).
type Entry struct {
    Number   int
    Reserved bool // true = tombstone (removed field/variant)
}

// Load parses a lock file. Returns (nil, nil) when path does not exist
// (absent lock is a valid state, not an error).
func Load(path string) (*Lock, error)

// Save serializes deterministically (see Determinism). Used only by generate.
func (l *Lock) Save(path string) error

// Reconcile produces the next lock and the library FieldNumbers from the current
// spec and the existing lock (nil = first-run seeding; a non-nil lock with no entry
// for a given message or enum is treated identically to nil for that unit). One
// uniform rule applies: for each message/enum, copy every existing live and reserved
// entry verbatim, then assign each absent field/variant name the next high-water-mark
// number in OpenAPI declaration order. A wholly new message or enum simply has an
// empty prior set, so its high-water mark starts at 1 (message) or 0 (enum) — seeding
// is not a distinct code path, it is append-only with an empty prior set. The ADR 0004
// precondition (first declared variant must be *_UNSPECIFIED) applies to any enum
// introduced for the first time, whether on first run or when a new enum is added to
// an existing lock. An empty-but-present lock (version: 1, no messages/enums) is
// treated the same as a nil lock for every message and enum in the spec.
// Reconcile is pure: it writes nothing and returns an error (with no partial state)
// on any forced reassignment, collision, corruption, or seeding-precondition violation.
func Reconcile(doc *v3.Document, existing *Lock) (*Lock, *schema.FieldNumbers, error)

// Check runs the lock-consistency checks against a committed lock + spec.
// Returns structured findings the lint package converts to Violations.
func Check(doc *v3.Document, lock *Lock) []Finding
```

### Library extension (`openapi-schema.go`)

`ConvertOptions` gains an optional `FieldNumbers`. When non-nil it fully overrides positional
assignment; when nil the library behaves as today (back-compat for non-duh consumers).

```go
type ConvertOptions struct {
    PackageName   string
    PackagePath   string
    GoPackagePath string
    FieldNumbers  *FieldNumbers // NEW; nil → positional numbering
}

type FieldNumbers struct {
    Messages map[string]MessageNumbers // key: OpenAPI schema name (matches ProtoMessage.OriginalSchema)
    Enums    map[string]EnumNumbers    // key: OpenAPI schema name
}
type MessageNumbers struct {
    Fields   map[string]int // key: JSON field name (matches ProtoField.JSONName)
    Reserved []int          // emitted as `reserved N;`
}
type EnumNumbers struct {
    Variants map[string]int // key: literal enum value (matches enum value.Value)
    Reserved []int
}
```

Three library changes:

1. **Field/enum lookup by name.** In `buildMessage`/`buildNestedMessage`, when `FieldNumbers` is
   set, look up each field's number by `ProtoField.JSONName` in `Messages[msg.OriginalSchema]`
   instead of the positional counter. In `buildEnum`, look up by the literal `value.Value`.
2. **`reserved` rendering.** `generator.go` emits `reserved N, M;` (ascending) for each message
   and enum from the `Reserved` lists.
3. **Enum numbering correction** (see below).

`FieldNumbers` carries plain integers (not the lock's tombstone model); duh-cli flattens
tombstones into the `Reserved` lists when building it.

### Enum numbering correction

The library currently (a) synthesizes `<ENUM>_UNSPECIFIED = 0` unconditionally and (b) numbers
every declared value at `i + 1`. Against a DUH spec — which declares `*_UNSPECIFIED` as the first
value per ADR 0004 — this produces a duplicate sentinel and an off-by-one. The fix:

- **Remove auto-synthesis.** The library no longer invents an `UNSPECIFIED`.
- **Declaration-order from zero.** Without `FieldNumbers`, declared values map `0, 1, 2, …` in
  declaration order; the first declared value (the sentinel, by ADR 0004) becomes `0` with no
  special case.
- **With `FieldNumbers`**, variant numbers come from the map verbatim (sentinel is `0` because the
  lock says so).

This implements ADR 0004 in the library; it is **not a new decision**, so no new ADR. It is a
breaking change to the library's default behavior for any non-DUH consumer that relied on
auto-synthesis (such a consumer must now declare its own zero value, which proto3 requires
anyway).

[NEEDS LIBRARY FIX: `ToEnumValueName("Status", "STATUS_UNSPECIFIED")` must not double-prefix into
`STATUS_STATUS_UNSPECIFIED`; verify sentinel and already-prefixed values render once.]

### `converter.go` migration

`ProtoConverter` changes to the new library and threads `FieldNumbers`:

```go
type ProtoConverter interface {
    Convert(openapi []byte, packageName, packagePath string, nums *schema.FieldNumbers) ([]byte, error)
}
```

The wrapper calls `schema.Convert` and returns `ConvertResult.Protobuf`. DUH specs cannot contain
`oneOf`/`allOf` (lint rules `PROHIBITED_ANYOF`, `PROHIBITED_ALLOF_UNION`,
`RPC_PROHIBITED_ONEOF_AND_ALLOF`) and `generate` requires a clean lint, so `ConvertResult.Golang`
is expected empty; if it is non-empty the wrapper errors (a union slipped past lint).

## Data Model

### Lock file format

YAML, co-located with the spec, deterministic and timestamp-free.

```yaml
version: 1
messages:
  CreateUserRequest:
    fields:
      name:        {number: 1}
      email:       {number: 2}
      displayName: {number: 4}
      nickname:    {number: 3, reserved: true}   # tombstone: removed, number retained
  UpdateUserRequest:
    fields:
      user_id: {number: 1}
enums:
  Status:
    variants:
      STATUS_UNSPECIFIED: {number: 0}
      active:             {number: 1}
      inactive:           {number: 2}
      suspended:          {number: 3}
```

**Reserved as tombstones.** A removed field/variant keeps its entry with `reserved: true`,
retaining the original name→number binding. Chosen over a bare `reserved: [3]` list so the
removal history (`nickname` once held `3`) is self-evident in a `go.sum`-style diff review, which
is the PRD's primary guard against a bad hand-edit.

**Keys are OpenAPI-native, never proto-derived:**
- Message key = OpenAPI component schema name (joins to `ProtoMessage.OriginalSchema`).
- Field key = JSON field name (joins to `ProtoField.JSONName`).
- Enum variant key = literal OpenAPI enum value (joins to enum `value.Value`); this is the
  JSON-wire identifier and survives proto-constant-name derivation changes.

Nested/synthesized message types are **not represented**: `duh lint` rejects inline/nested objects
and nested arrays, and `generate` refuses to run on a spec that does not lint clean, so no
synthesized message can reach the lock. Every locked message is a top-level component schema.
`map<>` fields are ordinary fields in their parent (keyed by JSON field name); a message-typed map
value is itself a named component schema, keyed normally.

### Determinism

`Save` must produce byte-identical output for an unchanged mapping regardless of spec field order
(acceptance #3). Requirements: no timestamps or volatile data in the file; map keys serialized in
a fixed order (ascending field/variant **number**, which is stable under spec reordering and
groups tombstones naturally); messages and enums serialized by ascending name.

### Number assignment (high-water mark)

Per message and per enum independently: the next number is `1 + max(all numbers ever used in this
unit, live and reserved)`. Removed → tombstone. When multiple new fields/variants appear in one
run, they are assigned in **OpenAPI declaration order among the new ones** (deterministic, matches
seeding semantics). A rename is remove + add: the old name becomes a tombstone, the new name gets
the next available number.

## Correctness

### Invariant Preservation

- **A published number is never changed.** `Reconcile` copies every existing live `Entry` number
  verbatim and only ever *adds* entries or flips live→reserved; it has no code path that rewrites
  an existing live number. If the spec state would force a different number for an existing name
  (impossible under append-only, reachable only via a corrupt/hand-mangled lock), `Reconcile`
  returns an error and writes nothing. Enforced by application logic — covered by tests.
- **A number is never reused.** New numbers come strictly from the high-water mark over *all*
  entries including reserved tombstones, so a reserved number can never be drawn. Structural: the
  assignment function cannot produce a used-or-reserved number.
- **Numbers are unique within a message/enum.** Guaranteed at write by the high-water-mark
  algorithm; verified at read by `Check` (uniqueness across live + reserved).
- **`*_UNSPECIFIED` is always `0`.** First-run seeding requires each enum's first declared variant
  to be its `*_UNSPECIFIED` (ADR 0004); `Reconcile` hard-errors (no files written) otherwise.
  After seeding the number is pinned by literal-value key, so reordering cannot disturb it.
  `Check` validates the sentinel maps to `0`. Enforced independently of the
  `ENUM_UNSPECIFIED_VARIANT` lint rule because `generate` cannot assume `lint` ran.
- **Every spec field/variant has a lock entry.** `generate` appends, so its output is always
  complete; `Check` completeness verifies a committed lock covers every live spec field/variant.

### Illegal State Analysis

- Structurally prevented: number reuse (high-water mark), in-unit duplicates at write time, the
  positional-renumber hazard itself (numbering keyed by name, not order).
- Enforced by application logic (needs test coverage): no-reassignment of an existing live number,
  tombstone retention across runs, seeding precondition.
- **Not detectable by design** (single-snapshot lint, no prior state): a hand-edit that changes a
  live number to another internally-consistent value, or deletes a tombstone outright. Guarded by
  the reviewable lock diff, not by lint (faithful to PRD *Concurrency / merge*).

### Behavioral Constraints

- **Never silently reassign / writes nothing on error.** `Reconcile` is pure and `Convert` can
  error on a forced collision; `generate` computes the resolved lock and runs conversion fully
  **in memory**, and only after both succeed does it write proto, code, and lock. This is a change
  from today's incremental `writeFile` calls in `internal/generate/duh/duh.go`.
- **Pure reorder is a no-op.** Numbers are keyed by name and the lock serializes by number, so a
  reorder yields a byte-identical lock; `FieldNumbers` makes proto numbering name-driven, so the
  proto is byte-identical too.
- **`duh lint` never writes the lock.** The lint path only `Load`s and `Check`s.

### Contracts at component boundaries

`converter.Convert` with non-nil `FieldNumbers` (precondition, guaranteed by `Reconcile`): every
live field/variant in the spec has an entry in the map; numbers are unique and in
`[1, 536870911]` (enum variants `[0, …]`) excluding the proto reserved range 19000–19999; reserved
lists are disjoint from live numbers. Postcondition: emitted proto uses exactly those numbers and
renders the reserved lists. The library should error (not silently fall back to positional) if a
live field/variant lacks a mapped number — that signals a duh-cli reconciliation bug.

## API Design

### `duh generate`

New flag `--lock-path` (a **file** path; default `<spec-dir>/fieldmap.lock`). A nonexistent parent
directory errors, matching `--output-dir`. The lock is written to this path, never under
`--output-dir`. Flow: load spec → `lint.Validate` (spec rules only, **not** the lock-consistency
checks) → `fieldmap.Load` existing lock → `fieldmap.Reconcile` → `converter.Convert` with
`FieldNumbers` → on success, write all artifacts including the lock.

`generate` deliberately skips the lock-consistency checks: running them pre-flight would deadlock
(adding a spec field fails completeness, but `generate` is what adds the entry). `generate` instead
relies on `Reconcile`'s own append-only invariants and fail-loud behavior. It independently enforces
the ADR 0004 enum seeding precondition.

### `duh lint`

New flag `--lock-path` (same semantics; default `<spec-dir>/fieldmap.lock`). Flow: existing spec
rules, then `ValidateLock(doc, lockPath)`. If no lock exists at the resolved path, the lock checks
are inactive and lint passes with **no warning** (absent lock is a first-class state — linting a
spec during API design, or a service that has not adopted a lock). When present, the checks run.

### `lint.ValidateLock` (new, lint package)

```go
func ValidateLock(doc *v3.Document, lockPath string) []Violation
```

Loads the lock via `fieldmap.Load`, runs `fieldmap.Check`, and maps `Finding`s to the existing
`Violation` type so they flow through the current reporter. The 50-rule `Rule` interface
(`Validate(doc *v3.Document) []Violation`) is **unchanged**; `ValidateLock` is a parallel pass
invoked by `run_cmd` alongside the rule chain, merged into the same `ValidationResult`.

`Check` findings (PRD checks 1–6): completeness, uniqueness (live + reserved), no-reassignment
(single-snapshot internal consistency — name→number identity, so a reorder is definitionally not a
change), reserved retention (no live number falls in the reserved set; no reserved number is also
live), enum invariant (sentinel = 0), structural validity (well-formed YAML in the expected shape;
catches a mangled merge).

## Error Handling

Follow the existing `Violation` reporter format (`internal/lint/rules/violation.go`) for lint
findings. `generate` fail-loud errors name the offending message/field/number and the likely
cause, e.g.:

- Seeding precondition: `enum "Status": first declared variant "active" is not an *_UNSPECIFIED variant; reorder it first (see ADR 0004)`
- Forced collision / corruption: `message "CreateUserRequest": field "email" maps to 2 in the lock but the spec state forces 5; refusing to reassign a published number (no files written)`

Exact wording is an implementation detail; the contract is "names message/field/number, points at
cause, writes nothing."

## Dependencies

- **`github.com/duh-rpc/openapi-schema.go`** — duh-cli migrates off `openapi-proto.go` v0.2.0 onto
  this renamed library, which must ship a release carrying the `FieldNumbers` API, `reserved`
  rendering, and the enum-numbering correction. During development, a `go.mod` `replace` points at
  the local `~/Development/openapi-schema.go` checkout.
- **`buf` toolchain (unchanged contract).** Generated proto stays buf-native; `reserved`
  statements now let `buf breaking` see retired numbers (PRD principle 6).
- **`libopenapi`** — already used; `Reconcile`/`Check` read the `*v3.Document` directly to
  enumerate component schemas, properties (JSON names, declaration order), and enum literal values.

## Testing

Testing follows the `surface-testing` skill: all tests drive `duh.RunCmd(stdout, args)` and assert
exit codes, stdout/stderr, and written artifacts (the lock file and proto). No internal calls.

Key surfaces:
- **[integration: `duh generate`]** seeding, mid-schema insert, reorder no-op (byte-identical lock
  + proto), removal→tombstone, rename, multi-new-field determinism, enum reorder/append/remove,
  independent-regeneration agreement, fail-loud (corrupt lock / forced collision → non-zero, no
  files written), pre-publish cleanup (delete lock + regenerate reseeds), `--lock-path` override,
  lock never written under `--output-dir`.
- **[integration: `duh lint`]** absent lock passes silently; completeness/uniqueness/reassignment/
  reserved/enum/structural failures; lint leaves a present lock byte-identical; lint with no lock
  creates none; `--lock-path` override.
- **No fakes needed.** The schema library is in-process; filesystem via `t.TempDir()`; no network,
  no async behavior, no wall-clock dependence in the lock path. The existing `ProtoConverter`
  interface stays injectable but tests use the real converter.

These map 1:1 to PRD acceptance criteria 1–11.

## Migration and Deployment

1. Land the `openapi-schema.go` changes (`FieldNumbers`, `reserved` rendering, enum correction);
   cut a release.
2. Migrate duh-cli's `converter.go` to the new library + `ConvertResult`; update `go.mod`.
3. Add `internal/fieldmap`; wire `Reconcile` into `generate` (with in-memory-then-atomic-write) and
   `ValidateLock` into `lint`; add `--lock-path` to both commands.

No data migration: existing greenfield specs seed a fresh lock on the next `generate`. Already-shipped
services whose declaration order does not match deployed numbers are an out-of-scope migration
caveat (PRD); v1 seeds from declaration order.

## Open Questions

None blocking. One library-side soft flag is recorded inline: the `ToEnumValueName` double-prefix
check for already-`UNSPECIFIED`-suffixed sentinels.
