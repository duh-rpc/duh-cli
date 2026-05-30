# `fieldmap.lock` — Stable Proto Field Numbers PRD

> Source material: [`background.md`](./background.md). That document captures the full
> problem analysis, the company API Design SOP intent, and the mono-repo ADR references.
> This PRD records the product decisions made on top of it.

## Problem

`duh generate` derives Protobuf messages from OpenAPI schemas, but the two formats have
incompatible backward-compatibility models. In JSON, field **order is meaningless** —
inserting or reordering a field is backward-compatible. In Protobuf, **field numbers
determine the wire encoding** and must never change for a field once published.

Today `duh` assigns field numbers **positionally** (the generator walks schema properties in
declaration order). Inserting or reordering a field mid-schema silently reassigns the numbers
of every field after it. Protobuf consumers then decode old wire data into the wrong fields —
**silent data corruption with no error**. JSON consumers are unaffected, which makes the
failure invisible in review and in any test that exercises only the JSON path. The same hazard
applies to **string enums**: on the JSON wire the variant *name* is transmitted (immune to
reordering), but on the proto wire each variant has an *integer*, so reordering an `enum` list
silently renumbers it.

This is the most dangerous class of contract bug in a Protobuf-over-OpenAPI system. The
current safety nets — the DUH linter rules and `buf breaking` in CI — do **not** prevent
positional field-number reassignment from a mid-schema edit. There is no mechanism that pins
field-name → field-number, so independent regenerations are not guaranteed wire-compatible.

This feature introduces a checked-in `fieldmap.lock` that pins the mapping and makes
regeneration deterministic and wire-safe.

## Users

- **Primary: the contract author.** The engineer who owns a service's `openapi.yaml` in its
  contract repository and runs `duh lint` / `duh generate`. Their mid-schema edit is what can
  silently break the wire, so the lock primarily protects them from themselves.
- **Secondary: the downstream consumer-service developer.** An engineer in another service who
  regenerates types from the provider's `openapi.yaml` **plus** `fieldmap.lock`. They need
  wire-compatible field numbers without coordinating with the provider. The shared lock is what
  makes "every service regenerates independently" safe — without it, two independent
  `duh generate` runs against the same spec could assign different numbers and silently disagree
  on the wire.

## Mental Model

`fieldmap.lock` is to proto field numbers what `go.sum` / `package-lock.json` is to
dependencies: a **machine-managed, checked-in artifact** that pins a mapping you must never
hand-edit casually. It records, per message, the JSON-field-name → proto-field-number mapping
(and per enum, the variant-name → proto-integer mapping). `duh generate` is the only thing that
writes it; `duh lint` verifies it.

It is deliberately **decoupled from proto syntax** — it maps the *JSON world* (field and variant
names, order-independent) to the *proto world* (field numbers, order-critical), which is exactly
the seam where silent corruption happens. The lock keys on **JSON-stable string identifiers** and
pins the **proto-side integers**.

It is **the one generated artifact that is checked in.** All other generated output (`.proto`,
`client.go`, `server.go`, buf configs) is regenerated on every build and typically gitignored. A
change to an *existing* entry in the lock should be treated with the same suspicion as a
surprising `go.sum` diff.

## Core Design Principles

These are the non-negotiables that shape every downstream decision. Principles 1, 2, and 5 are
the inviolable core; the rest follow from them.

1. **Existing mappings are immutable.** Once a field or enum variant has a number, it keeps it
   forever.
2. **Append-only assignment.** New fields/variants get the next available number regardless of
   their position in the OpenAPI schema.
3. **Removal reserves, never frees.** A removed field's/variant's number is retained as
   *reserved* and never reused (mirrors Protobuf `reserved` semantics).
4. **Machine-managed.** `duh generate` writes the lock; `duh lint` validates it. Hand-editing
   is not the intended workflow. The design *tolerates* the unavoidable human touch (resolving a
   merge conflict) and validates the result via `duh lint` — there is no tamper-seal.
5. **Fail loud, never silently reassign.** If the tool would be forced to change an existing
   mapping, it errors rather than producing a silently-incompatible result. This is the entire
   reason the feature exists.
6. **Second safety net, not a replacement.** `buf breaking` still runs and catches the
   categories the lock does not (type changes, message renames, removed-field breakage).

## Correctness Constraints

### State Invariants

- **A published number is never changed.** Violated by any edit that would map an existing
  field/variant name to a different number. Enforcement: `duh generate` preserves all existing
  entries and never rewrites one; `duh lint` fails if the committed lock maps a spec field to a
  number inconsistent with append-only history.
- **A number is never reused.** Violated by assigning a number that is already in use *or
  reserved* within the same message/enum. Enforcement: removed fields' numbers are retained as
  reserved; new assignments draw strictly from never-before-used numbers.
- **Numbers are unique within a message/enum.** Violated by two entries sharing a number.
  Enforcement: `duh lint` uniqueness check.
- **`*_UNSPECIFIED` enum variant is always `0`.** Violated by assigning it any other number or
  placing another variant at `0`. Enforcement: at first-run seeding, `duh generate` requires each
  enum's first declared variant to be its `*_UNSPECIFIED`, so declaration-order seeding assigns it
  `0` with no special case; if it is not first, `duh generate` hard-errors (non-zero exit, no files
  written) rather than seeding a violating lock. After seeding the number is pinned by name, so
  reordering cannot disturb it, and the lock lint check validates `*_UNSPECIFIED` is `0` thereafter.
  The `ENUM_UNSPECIFIED_VARIANT` linter rule enforces the same first-variant ordering at lint time;
  `duh generate` enforces it independently at seeding because it is a separate command that cannot
  assume `duh lint` was run.
- **Every spec field/variant has a lock entry.** Violated by editing the spec (adding a field)
  without regenerating, leaving the committed lock stale. Enforcement: `duh lint` completeness
  check fails.

#### Rationale: why `*_UNSPECIFIED` must be `0`

This is a Protobuf requirement, not a `duh` convention, and it is load-bearing for the same wire
safety the lock protects. The `ENUM_UNSPECIFIED_VARIANT` linter rule and the seeding precondition
above should reference this rationale, because "why must `0` be `*_UNSPECIFIED`?" predictably recurs.

- **Proto3 mandates a zero value.** An enum's first value must be `0`; `protoc` rejects otherwise.
  The zero value is the wire default — an enum field absent from the wire decodes as the `0` variant,
  and proto3 has no separate "unset" state for enum scalars.
- **`0` must mean "unset/unknown," never a real value.** Because `0` is what a decoder yields for an
  absent field, putting a real value there (e.g. `ACTIVE = 0`) makes "never set" indistinguishable
  from "deliberately `ACTIVE`" — a silent default-value confusion, the same class of silent mis-decode
  this feature exists to prevent. Reserving `0` for `*_UNSPECIFIED` keeps the two distinguishable.
- **It composes with append-only seeding.** Because authors declare `*_UNSPECIFIED` as the first
  `enum` entry, declaration-order seeding assigns it `0` with no special case; the seeding precondition
  (first variant is `*_UNSPECIFIED`) is what guarantees this.

### Behavioral Constraints

- **Never silently reassign a field number.** On any state that would force changing an existing
  mapping (corrupt lock, forced collision, impossible hand-edited state), `duh generate`
  **hard-errors with a non-zero exit and writes nothing** — no half-correct lock, no proto. The
  error names the offending message/field/number and points at the likely cause. There is no
  auto-repair, because auto-repair *is* the silent reassignment the feature forbids.
- **A pure field reorder is a no-op.** Once a message is established in the lock, reordering its
  fields in the OpenAPI schema (same names, no additions or removals) produces **zero change to
  the lock and zero change to proto numbers**; `duh generate` output is byte-identical and
  `duh lint` passes clean. This is the direct contrast to today's positional numbering. It also
  means `buf breaking` will not false-trigger on a reorder.
- **Independent regenerations agree on the wire.** Two `duh generate` runs against the same
  `openapi.yaml` + `fieldmap.lock` produce identical field numbering.
- **`duh lint` never writes the lock.** `duh generate` is the only command that creates or
  modifies `fieldmap.lock`. `duh lint` is read-only with respect to the lock: it validates a
  present lock and leaves its bytes unchanged, and it never creates a lock when one is absent.
- **Reserved numbers are never retired or reclaimed** (see *Out of Scope* for the detailed
  rationale).

### Concurrency / merge

The lock is a checked-in file, so the realistic concurrency hazard is two branches each
appending fields and producing a **merge conflict** (both claim the next number). Resolving such
a conflict by hand is expected and supported; `duh lint` validates that the resolved result is
consistent with the spec. There is no integrity hash — a hash would false-positive on every
legitimate merge resolution and is redundant with lint re-deriving the mapping from the spec.

## Scope

### In Scope

- A checked-in `fieldmap.lock` recording, per message, JSON-field-name → proto-field-number, and
  per enum, variant-name → proto-integer.
- `duh generate`:
  - On first run for a spec with no existing lock, **seeds the lock from current OpenAPI
    declaration order** and writes it to the checked-in contract directory (co-located with the
    spec), not to the gitignored `--output-dir`.
  - On subsequent runs, **preserves all existing entries** and **appends** new fields/variants at
    the next available number, regardless of their position in the schema. The generated `.proto`
    reflects locked numbers, not positional ones.
  - **Reserves** the numbers of removed fields/variants; never reuses a reserved number.
  - Treats a **field rename as remove + add**: the old name's number becomes reserved, the renamed
    field gets the next available number, and generation proceeds without error.
  - **Hard-errors** (non-zero exit, no files written) on a corrupt/inconsistent lock or a forced
    collision; never auto-repairs.
- **Enums in scope**, treated identically to message fields: variant names are the lock key,
  append-only for new variants, reserve on removal, `*_UNSPECIFIED` pinned at `0`.
- `duh lint` validates the lock against the spec (CI gate). Checks:
  1. **Completeness** — every spec field/variant has a lock entry (fails on a stale lock where the
     spec gained a field but `generate` was not re-run).
  2. **Uniqueness** — no number used twice within a message/enum.
  3. **No reassignment** — every spec field's number matches the append-only expectation, evaluated
     by name→number identity (so a reorder is definitionally not a change).
  4. **Reserved retention** — no number in the `reserved` list is also assigned to a live
     field/variant in the same message/enum, and no live field/variant's number falls in the
     `reserved` set. The committed `reserved` list is treated as authoritative; lint does not
     reconstruct removal history (it has no prior snapshot to do so).
  5. **Enum invariant** — `*_UNSPECIFIED` is `0`, validated by this check directly (independent of
     the separate `ENUM_UNSPECIFIED_VARIANT` rule, which enforces first-variant ordering at the spec
     level).
  6. **Structural validity** — well-formed YAML in the expected shape (catches a mangled merge).

  **Absent lock.** `duh lint` is a standalone OpenAPI linter; the lock-consistency checks above
  are an additional layer that applies only when a `fieldmap.lock` is present. If no lock exists
  at the resolved path, those checks are inactive and `duh lint` passes (its other rules still
  run) — with no warning, because nothing is wrong. Linting an `openapi.yaml` that has no lock is
  a first-class use case: working through an API design during PRD work, before any code or
  generated types exist, needs the linter but not the lock. The lock checks engage once a lock is
  present. This applies uniformly: a shipped service that has not yet adopted a lock also produces no
  warning — the signal that positional numbering is active is the absence of the lock file itself, not
  a lint diagnostic.
- **Versioning by convention.** One lock per version directory, co-located with the contract
  (`v1/openapi.yaml`, `v1/schema.proto`, `v1/fieldmap.lock`). The tool is version-agnostic and
  operates on one spec + one lock per invocation; it must *support* co-location (the lock path is
  specifiable/derivable so it lands next to the spec) but does not enforce or understand the
  `v1/`/`v2/` layout.

### Out of Scope / Non-Goals

- **Bootstrapping an already-shipped service from its deployed `.proto`.** v1 is greenfield-only:
  first `duh generate` seeds the lock from declaration order, and that order becomes the source of
  truth. A service that already shipped proto whose **declaration order does not match its deployed
  numbers** must reconcile the order before its first run — v1 does **not** detect or protect a
  mis-ordered already-shipped service from itself. This is a documented migration caveat. Reading an
  existing `.proto`/descriptor to seed deployed numbers is a possible future enhancement.
- **Retiring or reclaiming reserved numbers.** Reserved is permanent. Rationale (to document in
  detail for future users): old serialized wire data can persist at rest, in queues, in caches, and
  in old clients you do not control, so "no longer used in the *current* schema" never proves a
  number is safe to reclaim — reuse reintroduces exactly the silent mis-decode this feature exists
  to prevent. Protobuf itself works this way (`reserved` has no un-reserve workflow). The field-number
  space is effectively unlimited (~536M for messages), so there is no capacity pressure to reclaim;
  the only cost of a burned low number is losing a 1-byte tag (1–15) to a 2-byte one, a
  micro-optimization, not a correctness or capacity concern. The one legitimate case — churn *before
  anything ships*, where development added/removed fields and the lock accumulated tombstones — is
  handled without any retirement feature: an unpublished lock has no authority, so the cleanup is to
  **delete `fieldmap.lock` and regenerate**.
- **Reserving field/variant *names*.** Only numbers are reserved. The goal is JSON + protobuf wire
  safety (never reuse a number), not forcing perpetual new names. Re-adding a previously-used field
  name is allowed and gets the next available number while the old number stays reserved. (A
  JSON-name-reuse lint *warning* could be added later if it proves necessary.)
- **`duh lint` invoking `buf breaking`.** The two stay separate, complementary CI steps. `duh lint`
  validates spec + lock consistency on a single revision (no buf binary, no git base ref, no
  generated proto required). `buf breaking` is the separate cross-revision proto diff for type
  changes, renames, and removals.
- **A dedicated CI check for "a changed existing mapping" (SOP rule 5).** The inconsistent case is
  caught automatically by the `duh lint` no-reassignment check; the consistent-but-deliberate
  renumber is rare, requires editing two files against documented prohibition, and is served by the
  reviewable lock diff plus the "treat a surprising lock diff like a surprising `go.sum` diff"
  guidance. A diff-annotation/CODEOWNERS helper is a possible future enhancement.
- **Success metrics.** Deliberately omitted. This is insurance against a catastrophic-but-rare
  failure; its true success is a non-event (zero silent wire-corruption incidents), which is not
  meaningfully attributable or countable. Adoption counts and lint-firing counts are execution, not
  outcome, so no quantitative target is claimed.

## Acceptance Criteria

Turn each into a surface test (`duh.RunCmd(...)`, asserting exit code and stdout/stderr, plus the
written lock and proto):

1. **First-run seeding.** `duh generate` on a spec with no existing lock creates `fieldmap.lock`
   mapping every schema field (and enum variant) to a number assigned by current declaration
   order. The lock is written adjacent to the spec file at `<spec-dir>/fieldmap.lock` by default;
   a `--lock-path` flag overrides this location. The lock is never written to the `--output-dir`.
2. **Mid-schema insertion preserves numbers.** Re-running `duh generate` after inserting a field
   mid-schema preserves all existing numbers and assigns the new field the next available number;
   the generated `.proto` reflects locked numbers, not positional ones.
3. **Reorder is a no-op.** Reordering fields in an established schema (same names) yields a
   byte-identical lock and proto, and `duh lint` passes clean.
4. **Removal reserves.** Removing a field retains its number as reserved; a later-added field
   never reuses a reserved number.
5. **Rename = remove + add.** Renaming a field reserves the old name's number and assigns the new
   name the next available number, without error.
6. **Enum reordering is safe.** Reordering an enum's variants does not change their locked proto
   integers; adding a variant appends; removing one reserves; `*_UNSPECIFIED` stays `0`. First-run
   seeding requires each enum's first declared variant to be its `*_UNSPECIFIED`; seeding a spec
   whose enum lists `*_UNSPECIFIED` non-first exits non-zero without writing.
7. **Independent regenerations agree.** Two `duh generate` runs against the same
   `openapi.yaml` + `fieldmap.lock` produce identical field numbering.
8. **Lint catches inconsistency.** `duh lint` fails when the lock is missing a mapping the spec
   requires, contains a reused number, has an existing mapping changed, assigns a live field a number
   that is also reserved, or has been hand-edited/merge-mangled into an inconsistent or structurally
   invalid state. Lint validates a single snapshot with no tamper-seal: a hand-edit that deletes a
   reserved entry outright is not detectable by lint and is guarded by the reviewable lock diff (see
   *Concurrency / merge*), not by lint.
9. **Generate fails loud.** When `duh generate` would be forced to change an existing mapping
   (corrupt lock or forced collision), it exits non-zero, writes no files, and reports the
   offending message/field/number.
10. **Pre-publish cleanup.** Deleting `fieldmap.lock` and regenerating reseeds from current
    declaration order (valid recovery path while nothing has shipped).
11. **Lint never writes the lock.** Running `duh lint` against a present lock leaves the lock file
    byte-identical; running `duh lint` when no lock exists creates none.

## Dependencies and Constraints

- **`buf` toolchain (unchanged contract).** `duh generate` already emits `.proto`, `buf.yaml`, and
  `buf.gen.yaml` (templates in `internal/generate/duh/templates/`), so output is buf-native by
  construction. This feature **does not change the buf-facing output contract** — only field-number
  *assignment* changes (lock-driven instead of positional). `buf generate` and `buf breaking`
  continue to operate on the generated proto as today.
- **Current positional numbering** is performed inside the external `github.com/duh-rpc/openapi-proto.go`
  library, invoked via `conv.Convert()` in `internal/generate/duh/converter.go`. `parser.go` and
  `matcher.go` build Go template data; they do not assign field numbers. The feature replaces the
  positional assignment inside the library by feeding locked numbers through the `conv.Convert()` call
  site in `converter.go`.
- **Existing lint framework** in `internal/lint/validator.go` (~50 registered rules) is where the
  new lock consistency checks are added. The enum-zero invariant is enforced by the new lock check
  plus the seeding precondition that an enum's first declared variant be its `*_UNSPECIFIED` (see
  Correctness Constraints); the `ENUM_UNSPECIFIED_VARIANT` rule described in
  `docs/duh-linter-rules.md` enforces that same first-variant ordering at lint time, complementing
  the seeding precondition.
- **CLI wiring** in `run_cmd.go` / `cmd/` exposes the generate/lint commands and flags.
- **Surface-testing requirement** (project CLAUDE.md): all tests exercise `duh.RunCmd()` and assert
  exit codes and stdout/stderr; no direct internal calls.

## Open Questions

These are **tech-spec-level** decisions deferred to `tech-spec.md`:

1. **Lock file format details.** Confirm the exact YAML shape (message name → field name → number;
   enum name → variant name → number). How are reserved numbers represented — a `reserved: [...]`
   list per message/enum, or retained tombstone entries?
2. **Lock path flag details.** The lock defaults to `<spec-dir>/fieldmap.lock`, overridable by
   `--lock-path`. Tech-spec to decide: the exact flag name (`--lock-path` vs `--fieldmap`),
   behavior when the flag points to a nonexistent directory, and whether `duh lint` accepts the
   same flag or always reads from the default.
3. **Nested / synthesized messages.** The SOP wraps inner arrays in messages and forbids
   `oneOf`/`allOf`. How are synthesized message types keyed and numbered in the lock? How are
   `map<>` value messages handled?
4. **Determinism for multiple new fields in one run.** Define the ordering used when assigning numbers
   to several newly-added fields/variants in a single `generate` (e.g., OpenAPI declaration order among
   the new ones) so the result is reproducible.
5. **Exact wording and structure of generate/lint error messages** for the fail-loud cases.
