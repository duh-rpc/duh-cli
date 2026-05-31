# 3. fieldmap.lock is the checked-in record of proto field numbers

Date: 2026-05-29

## Status

Accepted

## Context

`duh generate` derives Protobuf messages from OpenAPI schemas, but the two formats have
incompatible compatibility models:

- OpenAPI field order is meaningless — reordering or inserting a field is JSON-compatible.
- Protobuf field numbers are wire-critical — a number must never change for a field once
  published.

With positional numbering, inserting a field mid-schema silently renumbers every later field,
and Protobuf consumers decode old wire data into the wrong fields with no error. Two services
that regenerate independently from the same spec can also assign different numbers and disagree
on the wire.

Something must pin field-name → field-number stably, survive arbitrary schema reordering, and be
shared across every service that regenerates from the contract. Two alternatives were weighed:

- **Check in the generated `.proto`** as the stability record. Rejected: it ties the record to
  Protobuf syntax and the generator's proto-emission choices, instead of a language-neutral
  mapping any consumer can regenerate against.
- **Rely on `buf breaking` plus lint rules alone.** Rejected: neither prevents positional
  field-number reassignment from a mid-schema insert.

## Decision

We will introduce a checked-in `fieldmap.lock`: a machine-managed file mapping JSON field and
enum-variant names to Protobuf field numbers, per message and per enum, decoupled from Protobuf
syntax. It is the single generated artifact that is committed; all other generated output is
regenerated on build and gitignored. The lock is machine-managed: numbers are assigned
append-only, so existing numbers never change and removed numbers are reserved rather than
reused.

## Consequences

- A field can be inserted or reordered anywhere in a schema without changing any existing wire
  number; the lock, not declaration order, determines numbering.
- Every consuming service must regenerate from the provider's `openapi.yaml` **and**
  `fieldmap.lock` to stay wire-compatible; the spec alone is no longer sufficient.
- The lock lives in the contract directory alongside the spec; placing it in a gitignored
  output directory voids the stability guarantee.
- The committed mapping must be reviewed like any other source of truth: a changed existing entry
  is a wire-breaking change.
- `buf breaking` remains a complementary safety net for the changes the lock does not cover —
  type changes, renames, and removals.
