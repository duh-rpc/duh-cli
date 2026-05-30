# 7. Enum variants are keyed in fieldmap.lock by their literal OpenAPI value

Date: 2026-05-29

## Status

Accepted

## Context

`fieldmap.lock` pins each enum variant to a stable Protobuf number. A DUH spec declares the
`*_UNSPECIFIED` sentinel first (ADR 0004):

```yaml
Status:
  type: string
  enum: [STATUS_UNSPECIFIED, active, inactive, suspended]
```

The lock must choose an identifier to key each variant on. Two candidates:

**Literal OpenAPI value** — the exact string from the `enum:` list:

```yaml
enums:
  Status:
    variants:
      STATUS_UNSPECIFIED: {number: 0}
      active:             {number: 1}
      inactive:           {number: 2}
      suspended:          {number: 3}
```

**Derived proto constant** — the generated constant name (`ToEnumValueName("Status", "active")`):

```yaml
enums:
  Status:
    variants:
      STATUS_UNSPECIFIED: {number: 0}
      STATUS_ACTIVE:      {number: 1}
      STATUS_INACTIVE:    {number: 2}
      STATUS_SUSPENDED:   {number: 3}
```

The literal value is the JSON-wire identifier — on the JSON wire the variant name (`"active"`) is
what is transmitted, and the lock exists to map the JSON world to proto numbers. The proto constant,
by contrast, is *derived* from the schema name plus the value, so it moves whenever the derivation
input changes. Renaming the schema `Status` → `UserStatus` shifts every constant:

```
# literal-value key — survives the rename, numbers intact
  active: {number: 1}              →   active: {number: 1}            ✓

# constant key — every variant looks removed + re-added
  STATUS_ACTIVE: {number: 1}
    → tombstoned as reserved, and:
  USER_STATUS_ACTIVE: {number: 5}  ← silent renumber of a live variant
```

A shifted key reads as "old variant removed, new variant added," which tombstones a live number and
hands the live variant a new one — exactly the silent renumbering this whole mechanism exists to
prevent. Keying on the constant would also force lint and reconciliation to re-run the library's
naming logic just to match entries.

## Decision

We will key each enum variant in `fieldmap.lock` by its literal OpenAPI `enum` value (the string as
written in the spec), not by the derived proto constant name. Reconciliation and lint read the
literal directly from the schema's enum values.

## Consequences

- The lock key is the JSON-wire identifier, consistent with the lock's purpose of mapping the JSON
  world to proto numbers.
- Renaming an enum schema, or changing the proto-constant derivation rule, does not disturb variant
  numbers — the literal value is unaffected.
- Lint and reconcile match variants without re-deriving proto constant names.
- The lock keys (`active`) differ visually from the constants that appear in the generated proto
  (`STATUS_ACTIVE`), which a reader must keep in mind; the ADRs' use of `*_UNSPECIFIED` in
  constant form does not imply the lock keys on constants.
