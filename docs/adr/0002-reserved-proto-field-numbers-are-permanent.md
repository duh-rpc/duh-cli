# 2. Reserved proto field numbers are permanent

Date: 2026-05-29

## Status

Accepted

## Context

`duh generate` pins JSON field and enum-variant names to stable Protobuf field numbers in a
checked-in `fieldmap.lock`. When a field or variant is removed from the OpenAPI schema, its
number must not be handed to a different field, or Protobuf consumers reading older wire data
silently decode it into the wrong field.

A natural request is to reclaim such a number once the field appears unused, to keep low
single-byte tags (1–15) available. The forces against it:

- Serialized wire data persists beyond the current schema — at rest, in queues, in caches, and
  in old clients outside our control. "No longer used in the current schema" never proves no
  wire data still encodes that number.
- Reusing a number causes a silent mis-decode, the exact failure this mechanism exists to
  prevent.
- The field-number space is effectively unlimited (~536M usable numbers per message), so there
  is no capacity pressure. The only cost of a retired low number is one extra tag byte.
- Protobuf itself provides no un-reserve operation for the same reason.

## Decision

We will treat a reserved field or enum-variant number as permanent. Once a number is removed
from active use it is retained as reserved forever and never reassigned. `duh` provides no
mechanism to retire or reclaim a reserved number.

## Consequences

- Wire compatibility is guaranteed across removals and re-additions without depending on a human
  judging whether old data still exists.
- Reserved numbers accumulate in long-lived locks, and single-byte tags consumed by removed
  fields stay consumed.
- A field re-added under a previously removed number is impossible; it receives a new number,
  which may push frequently-used fields past the single-byte tag boundary over time.
- Before a contract is published the lock carries no authority, so it may be discarded and
  regenerated from the current schema without violating this rule; after a contract ships, the
  lock only grows.
