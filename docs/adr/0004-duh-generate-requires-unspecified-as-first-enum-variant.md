# 4. duh generate requires *_UNSPECIFIED as the first enum variant

Date: 2026-05-29

## Status

Accepted

## Context

`duh generate` derives Protobuf enums from OpenAPI string enums and pins each variant name to a
stable Protobuf number in `fieldmap.lock`. Protobuf constrains the number `0` in ways that JSON does
not:

- Proto3 mandates that an enum's first value is `0`; `protoc` rejects any enum whose first declared
  value is non-zero.
- `0` is the wire default. An enum field absent from the wire decodes as the `0` variant, and proto3
  has no separate "unset" state for enum scalars.

These two facts make `0` load-bearing for the wire safety the lock exists to protect. If a real
business value occupies `0` (for example `ACTIVE = 0`), a decoder cannot distinguish "the field was
never set" from "the field was deliberately `ACTIVE`" — a silent default-value confusion, the same
class of silent mis-decode the lock is built to prevent. Reserving `0` for an `*_UNSPECIFIED`
sentinel keeps the two distinguishable.

The lock seeds numbers from OpenAPI declaration order on first run, so an enum whose `*_UNSPECIFIED`
variant is not declared first would receive a non-zero seed number. The `ENUM_UNSPECIFIED_VARIANT`
linter rule enforces `*_UNSPECIFIED`-first at lint time, but `duh generate` is a separate command that
can be run without `duh lint`, so a non-conforming spec can still reach seeding unguarded unless
`generate` enforces the ordering itself.

Several resolutions were weighed:

- **Silently force `*_UNSPECIFIED` to `0` during seeding regardless of declaration position.**
  Rejected: silently reorders the lock's enum numbers away from declaration order, violating the
  fail-loud principle.
- **Rely on the `ENUM_UNSPECIFIED_VARIANT` linter rule alone.** Rejected: `duh generate` can be run
  without `duh lint`, leaving seeding unguarded.
- **Add an integrity hash or snapshot to detect the bad lock after the fact.** Rejected: detection
  after a violating lock is written is strictly worse than refusing to write one.

## Decision

We will require that every enum's first declared variant in the OpenAPI schema is its
`*_UNSPECIFIED` variant, enforced as a first-run seeding precondition. When `duh generate` seeds a
lock and an enum's first declared variant is not `*_UNSPECIFIED`, it hard-errors with a non-zero exit
and writes no files.

## Consequences

- A seeded lock always satisfies `*_UNSPECIFIED=0`; because the variant is declared first,
  declaration-order seeding assigns it `0` with no special case, and the invariant holds by
  construction rather than by post-hoc validation.
- An enum that lists `*_UNSPECIFIED` non-first cannot be seeded; the author must reorder the variant
  before running generation.
- Once seeded, a variant number is pinned by name, so the precondition is reachable only at first-run
  seeding — later reordering or appending variants cannot trigger it.
