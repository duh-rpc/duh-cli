# 5. Field numbers are injected into proto generation via a structured library option

Date: 2026-05-29

## Status

Accepted

## Context

`duh generate` does not assign Protobuf field numbers itself. It hands the OpenAPI bytes to the
proto-conversion library (`github.com/duh-rpc/openapi-schema.go`, formerly `openapi-proto.go`),
which builds the proto. Numbering lives inside that library's non-importable `internal` package and
is purely positional:

```go
// inside the library's builder — before this decision
fieldNumber := 1
for propName := range schema.Properties.FromOldest() {
    field.Number = fieldNumber // position decides the number
    fieldNumber++
}
```

`fieldmap.lock` (ADR 0003) must make the *lock*, not declaration order, decide field numbers. The
library exposes no injection point, so something has to change. Three approaches were weighed:

- **Write `x-proto-number` into the spec.** The library already reads an `x-proto-number` extension
  on each property. duh-cli could inject numbers into the spec before conversion. Rejected: it
  pollutes the contract author's spec with numbers the lock is meant to own; the library enforces an
  all-or-nothing rule per schema; it has no enum support and emits no `reserved`; and because
  `Convert()` takes raw bytes, duh-cli would have to re-serialize mangled YAML.
- **Post-process the generated proto text.** Let the library emit positional proto, then rewrite
  each `= N` using the `json_name` annotation as a join key. Rejected: a fragile text-manipulation
  layer over an external output format that breaks whenever proto formatting shifts, and re-derives
  information the library already has.
- **Pass numbers as structured data to the library.** Add an optional, typed field-number map to
  the library's options that overrides positional assignment when present. Chosen.

The library is owned by the same organization, so extending it is viable rather than an upstream
dependency we cannot change.

## Decision

We will extend `openapi-schema.go` with an optional `FieldNumbers` field on `ConvertOptions`. When
non-nil it fully overrides positional assignment and the generator emits `reserved` statements; when
nil the library behaves exactly as before, preserving backward compatibility for other consumers.

```go
type ConvertOptions struct {
    PackageName   string
    PackagePath   string
    GoPackagePath string
    FieldNumbers  *FieldNumbers // nil ⇒ positional numbering
}

type FieldNumbers struct {
    Messages map[string]MessageNumbers // key: OpenAPI schema name
    Enums    map[string]EnumNumbers    // key: OpenAPI schema name
}
type MessageNumbers struct {
    Fields   map[string]int // key: JSON field name → proto number
    Reserved []int          // rendered as `reserved N;`
}
```

The three map keys are OpenAPI-native and join to fields the library already tracks: the message key
to `ProtoMessage.OriginalSchema`, the field key to `ProtoField.JSONName`, and the enum-variant key
to the enum's literal `value.Value`. `duh generate` computes `FieldNumbers` from `fieldmap.lock` and
passes it on every run; the library numbers strictly by name and ignores declaration order.

Adopting this requires migrating duh-cli off `openapi-proto.go` v0.2.0 onto `openapi-schema.go` and
its `*ConvertResult` API.

## Consequences

- The lock-to-proto contract is structured data, not text, so it is unit-testable and immune to
  proto formatting changes.
- Messages, enums, and reserved numbers are handled through one uniform mechanism.
- The author's `openapi.yaml` stays free of proto field numbers; numbers live only in the lock.
- The library keeps working unchanged for non-duh consumers (nil `FieldNumbers`), but duh-cli now
  depends on a library release that carries this API, the `reserved` rendering, and the enum
  numbering correction.
- duh-cli must migrate to the renamed library and adapt its converter wrapper to `*ConvertResult`.
