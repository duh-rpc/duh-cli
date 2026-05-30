# 6. Removed fields are recorded in fieldmap.lock as tombstone entries

Date: 2026-05-29

## Status

Accepted

## Context

Reserved proto field numbers are permanent (ADR 0002): once a number is used it is never
reassigned, because old wire data may still carry it. `fieldmap.lock` must record that a number is
reserved. The question this decision settles is purely *how the lock file represents a removed
field* — there are two shapes.

Given a message that starts with three fields, then has `nickname` removed and `displayName` added:

**Tombstone entry** — keep the removed field's entry, flagged dead:

```yaml
messages:
  CreateUserRequest:
    fields:
      name:        {number: 1}
      email:       {number: 2}
      nickname:    {number: 3, reserved: true}   # removed; number retained
      displayName: {number: 4}
```

**Bare reserved list** — drop the entry, move the number into a list:

```yaml
messages:
  CreateUserRequest:
    fields:
      name:        {number: 1}
      email:       {number: 2}
      displayName: {number: 4}
    reserved: [3]
```

Both retain the number `3` and both let lint enforce that `3` is never reused. They differ in what a
human reviewer sees in a diff. The PRD's primary guard against a bad hand-edit is the reviewable
lock diff — "treat a surprising lock diff like a surprising `go.sum` diff." The tombstone diff
explains itself; the reserved-list diff requires the reviewer to remember which field owned the
number:

```
# tombstone diff — self-explanatory
-     nickname: {number: 3}
+     nickname: {number: 3, reserved: true}

# reserved-list diff — needs archaeology
-     nickname: {number: 3}
+   reserved: [3]   # reviewer must recall that 3 was nickname
```

The same property makes a rename (modeled as remove + add) legible — the lock tells the whole story
in place:

```yaml
      email:    {number: 2}
      nickname: {number: 3, reserved: true}   # was here
      handle:   {number: 4}                    # renamed to here
```

## Decision

We will record a removed field or enum variant as a tombstone: its original name→number entry is
retained in the lock with `reserved: true`, rather than moving the bare number into a separate
`reserved` list.

The conversion library still receives a flat `Reserved: [3]` list inside `FieldNumbers`; duh-cli
flattens tombstones into that list at conversion time. The tombstone is the *file* representation;
the flat list is only the wire to the library.

## Consequences

- Removal history is preserved: which field held a reserved number stays visible forever, so the
  lock diff a reviewer reads is self-explanatory.
- A rename reads naturally as an old name going reserved and a new name appearing.
- The lint reserved-retention check reads tombstones directly as authoritative, with no separate
  list to cross-reference.
- The file is slightly busier than a bare list, and it does not mirror Protobuf's own
  `reserved [3];` syntax — a deliberate trade of compactness and proto-symmetry for
  diff-reviewability.
