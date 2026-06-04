# Wire-Compatible oneOf (Style B) Blueprint

## Objective

DUH-RPC serves the same OpenAPI schema as both `application/json` and `application/protobuf`,
so it only accepts schema constructs that can be described truthfully on **both** wires. Today the
toolchain rejects every use of `oneOf`: openapi-schema.go errors on any `oneOf` without a
discriminator, and duh-cli's `PROHIBITED_ONEOF` rule fires on the keyword unconditionally. That
blanket ban is broader than wire-compatibility requires.

There is a `oneOf` form that **is** byte-compatible on both wires — proven end-to-end with the real
`protoc`/`protojson` toolchain (`docs/oneof-wire-proof/`, written up in
`docs/oneof-wire-compatibility.html`). This feature makes the toolchain accept that one form,
"style B," and generate a protobuf `oneof` for it. The flat/discriminated `oneOf` stays prohibited,
because protobuf cannot produce its flattened JSON.

This is the contract, stated once: **a `oneOf` of single-`required` branches over optional `$ref`
properties, with no discriminator, maps to a protobuf `oneof`.** The explainer document is the
normative description of the shape; this blueprint codifies the toolchain changes that honor it.

## Mental Model

A contract author expresses a tagged union as an object whose properties are the variants — one
optional `$ref` property per variant — plus a `oneOf` that declares exactly one of them must be
present:

```yaml
Event:
  type: object
  properties:
    cat_event: { $ref: '#/components/schemas/Cat' }
    dog_event: { $ref: '#/components/schemas/Dog' }
  oneOf:
    - required: [cat_event]
    - required: [dog_event]
```

The **key is the tag**. On the wire this is the nested, key-tagged shape
`{"cat_event": {"pet_name": "Whiskers"}}`, which is exactly what a protobuf `oneof` emits. There is
**no `discriminator`** — a discriminator is the mechanism of the *prohibited* flat form, where
variant fields are hoisted to the top level (`{"type": "cat", "pet_name": "Whiskers"}`), a shape no
protobuf serialization can produce.

The distinction the whole feature turns on is mechanical:

- `oneOf` **with a `discriminator`** (or with `$ref`/inline variant schemas) → flat form → **prohibited**.
- `oneOf` **of `required`** over named object properties, **no discriminator** → nested form → **supported (style B)**.

Style B differs from the already-supported "optional properties, no `oneOf`" form (style A) only in
that it declares the variants mutually exclusive — which is why it earns a real protobuf `oneof`
(style A generates plain optional fields).

## Correctness Constraints

### State Invariants

- **The generated proto for a style-B schema preserves the nested wire shape.** Marshaling a
  single-variant value via `protojson` (duh.go's marshaler, bare defaults) produces
  `{"<variant>": {...}}` with the variant's original snake_case property name as the key. Violated
  if generation drops the `json_name` annotation (protojson would then emit lowerCamelCase).
  Enforced structurally: openapi-schema.go already emits `[json_name = "<property>"]` verbatim for
  every field (`internal/proto/builder.go`), and the proof harness asserts it.

- **`oneof` members carry stable fieldmap-lock numbers, append-only.** Each variant property is a
  normal proto field with a number assigned by the existing positional/locked numbering. Adding a
  variant takes the next free number; removing one reserves its number permanently as a tombstone
  (ADR 0002, ADR 0006). A protobuf `oneof` grouping does not change numbering — it groups
  already-numbered fields. Violated if the `oneof` grouping is allowed to renumber or reuse a
  member's number.

- **The `oneof` group name is not wire-significant.** Only field numbers and `json_name` values
  appear on either wire; the `oneof` group identifier is cosmetic and may change freely without a
  wire break. It is therefore **not** recorded in `fieldmap.lock`.

### Behavioral Constraints

- **The toolchain MUST NEVER accept the flat/discriminated `oneOf`.** It cannot be served on both
  wires. This is the load-bearing constraint the whole prohibition exists to enforce; relaxing it
  for style B must not open a path for the flat form.

- **The toolchain MUST NEVER emit a `oneof` containing a `repeated` field.** proto3 forbids
  `repeated` inside a `oneof`. A style-B variant property whose schema is an array must therefore be
  rejected at validation time, not discovered by `protoc` downstream.

- **`duh generate` MUST NOT route a style-B schema to Go-struct codegen.** The existing discriminated
  path produces a flat-JSON Go union and is rejected by the duh-cli converter wrapper. Style B must
  be built as a protobuf message with a `oneof`, leaving `ConvertResult.Golang` empty.

- **Concurrency / partial failure:** generation is a pure, single-threaded transform with no shared
  state; there is no concurrency invariant. Generation is all-or-nothing per spec — `duh lint` must
  reject a malformed style-B schema before `duh generate` runs, so a half-generated artifact is never
  produced.

## Acceptance Criteria

Each is mechanically verifiable through a public surface (`duh.RunCmd` for duh-cli, `Convert` for
openapi-schema.go, `protojson` round-trip for duh.go).

1. A spec containing a valid style-B schema passes `duh lint` with exit code 0.
2. `duh generate` on that spec emits a `.proto` whose message contains a `oneof` group with one
   field per variant, each annotated `[json_name = "<property>"]`, and writes/updates `fieldmap.lock`
   with a stable number per variant field.
3. A spec containing a flat/discriminated `oneOf` (a top-level `oneOf` of `$ref`s with a
   `discriminator`) still fails `duh lint` with `[PROHIBITED_ONEOF]` and a non-zero exit code.
4. A malformed style-B schema fails `duh lint` with a clear message and non-zero exit for each of:
   a `oneOf` branch with more than one `required` entry; a branch naming a property absent from
   `properties`; a variant property whose schema is an array.
5. The four `DISCRIMINATOR_*` rules are absent from the registered rule set; a style-B schema does
   **not** produce any `DISCRIMINATOR_REQUIRED` (or other discriminator) violation.
6. A generated style-B protobuf message, marshaled with `protojson`, produces
   `{"<variant>": {...}}`; the bytes are semantically equal to the optional-fields encoding of the
   same value, and a client built from either form unmarshals the other (the proof-harness assertions,
   run against generated output).
7. Removing a variant from a style-B schema and regenerating produces a `reserved` entry for the
   removed field's number; the surviving variants keep their numbers.

## Scope

### In Scope

- openapi-schema.go: recognize the style-B pattern, build it as a protobuf message with a `oneof`,
  and add `oneof` rendering to the proto generator (new capability — it never emitted `oneof` before).
- duh-cli: narrow `PROHIBITED_ONEOF` to the flat form; delete the four `DISCRIMINATOR_*` rules,
  their tests, and testdata; update `docs/duh-linter-rules.md`.
- Style B at the **top level** of `components/schemas`, with **`$ref` (message-typed) variant
  properties**, optionally alongside regular always-present fields in the same message (proto3
  permits a `oneof` next to normal fields).
- A permanent wire-compatibility regression test that runs against **generated** output.

### Out of Scope / Non-Goals

- **Scalar-typed variants** (a `oneOf` whose variants are `string`/`integer` properties). proto3
  permits scalar `oneof` members, but the documented shape illustrates `$ref` variants only; defer
  until asked. *Confirmed boundary: a style-B object whose variant properties are scalars is not a
  v1 target and may be rejected.*
- **Nested `oneOf`** — the pattern appearing inside a property's schema or array `items` rather than
  a top-level component schema.
- **The flat/discriminated `oneOf`** — remains prohibited by design, not a gap.
- **Removing the existing discriminated-`oneOf` → Go-struct path** in openapi-schema.go. It stays
  for non-duh-cli library consumers; this feature only ensures style B is routed to proto instead.
  duh-cli continues to reject any `ConvertResult.Golang` output.
- **Any change to duh.go production code.** Proven unnecessary; duh.go's role here is a regression test.

## Dependencies and Constraints

- **Ordering:** openapi-schema.go must land first. Narrowing the duh-cli linter alone would let a
  style-B spec pass `lint` while `generate` still errors in the unpatched library — moving the
  failure rather than fixing it. duh-cli's library dependency must be bumped to the released
  openapi-schema.go version before the linter change ships.
- Honors ADR 0002 (reserved numbers permanent), ADR 0005 (field numbers injected via structured
  library option), ADR 0006 (tombstone entries) — `oneof` members are ordinary numbered fields and
  inherit all three with no special-casing.
- proto3 `oneof` restrictions (no `repeated` members, members are implicitly presence-tracked) are a
  hard constraint the validation must respect.

---

## Functional

Three surfaces change.

**openapi-schema.go — detection.** A new predicate classifies a schema carrying `oneOf` as one of:
*flat/discriminated* (existing behavior — discriminator present or variants are `$ref`/inline
schemas), or *style B* (no discriminator; every `oneOf` branch is a constraint object whose only
meaningful key is `required` with exactly one entry; each named entry is a declared property of the
same object). A schema that has `oneOf` but matches neither (e.g. a branch with two `required`
entries, or `required` naming an undeclared property) is a **validation error**, not a silent
passthrough.

**openapi-schema.go — build.** The two-pass builder
(`internal/proto/builder.go`) currently (1) marks any `oneOf` schema as a union and (2) skips union
schemas during message building. Style B must instead flow into `buildMessage`:

- First pass: classify; mark *flat/discriminated* schemas as Go unions as today; do **not** mark
  style-B schemas as unions.
- Second pass: do **not** `continue` past style-B schemas; build them as messages.
- `buildMessage`: number `schema.Properties` exactly as today (this already gives each variant
  property its locked number), and additionally record that the variant properties belong to one
  `oneof` group.

**openapi-schema.go — generate.** The proto generator gains `oneof` rendering: a message emits
`oneof <name> { <member fields> }` wrapping the variant fields, with non-variant fields rendered
normally outside the group. Member fields keep their `[json_name = "..."]` annotations and numbers.
`reserved` statements remain at message level.

**duh-cli — lint.** `PROHIBITED_ONEOF` is narrowed: it fires only when a `oneOf` schema is the flat
form (has a `discriminator`, or has `$ref`/inline variant schemas). A style-B `oneOf` passes. A
malformed style-B `oneOf` produces a clear violation (reusing `PROHIBITED_ONEOF` with a message
naming the specific problem, or a sibling rule — implementer's call, contract is "clear message,
non-zero exit"). The four `DISCRIMINATOR_*` rules are deleted from the registry and repo.

## Architecture

The classification predicate is the single source of truth for "is this style B," and must be
applied consistently in **both** repos — the linter's allow/deny decision and the library's
build/route decision must agree, or a spec could lint clean and fail to generate (or vice versa).
The predicate is simple and pure (it inspects `discriminator`, `oneOf` branch structure, and
`properties` of one schema); each repo implements it against its own OpenAPI datamodel
(`libopenapi` high-level model). The acceptance tests in each repo pin the boundaries so the two
implementations cannot drift silently.

## Data Design

The `ProtoMessage` model in openapi-schema.go gains a representation of `oneof` groups — a named
group referencing a subset of the message's fields. This is the minimal contract; the exact struct
shape (e.g. a `Oneofs []*ProtoOneof{ Name string; Fields []*ProtoField }` versus a group tag on each
`ProtoField`) is an implementation choice left to the build, constrained only by: the generator must
render a valid proto3 `oneof`, and field numbering/reserved handling must be untouched.

### Invariant Preservation

- *Stable numbers:* numbering happens in the existing `schema.Properties` loop before any `oneof`
  grouping is applied; the grouping references fields by identity, never reassigning numbers. The
  fieldmap-lock path (`FieldNumbers` structured option, ADR 0005) is unchanged.
- *Nested wire shape:* unchanged `json_name` emission preserves snake_case keys; proven by the
  harness against generated output.
- *No flat form:* the classifier routes anything with a discriminator or `$ref` variants to the
  prohibited/Go path, never to the `oneof` builder.

### Illegal State Analysis

- A `oneof` containing a `repeated` field is unrepresentable in valid proto3 → rejected at
  validation (array-typed variant property is an error) rather than represented and failing in
  `protoc`.
- A style-B branch with `required: [a, b]` cannot map to "exactly one of" → rejected at validation.
- `required` naming a property not in `properties` → rejected at validation (a `oneof` member with no
  backing field would be unrepresentable).

## Testing

Testing follows the `surface-testing` skill.

Key surfaces:
- **duh-cli** — `duh.RunCmd(...)` only (per project CLAUDE.md). Surface tests with testdata specs:
  style-B spec lints clean and `generate` emits a `.proto` containing `oneof`; flat/discriminated
  spec still fails `[PROHIBITED_ONEOF]`; each malformed style-B variant fails with non-zero exit;
  no `DISCRIMINATOR_*` violation is ever emitted.
- **openapi-schema.go** — `Convert(...)`. Golden tests asserting the emitted `.proto` text contains
  the expected `oneof` group and `json_name` annotations, and that `fieldmap.lock` numbers the
  variant fields append-only (including a remove-variant → `reserved` case).
- **duh.go / proof harness** — the wire-compatibility regression test (`docs/oneof-wire-proof/`)
  re-pointed at **generated** style-B output: marshal one variant via `protojson`, assert the nested
  key-tagged JSON, and cross-unmarshal between the `oneof` form and the optional-fields form (JSON
  and binary). This converts the one-off proof into a standing guard.
- **fakes needed:** none — the toolchain is a pure transform; tests use real specs and the real
  `protoc`/`protojson` toolchain (already available in the environment).

## Limitations & Future Work

- Scalar-typed and nested style-B unions are deferred (see Non-Goals); both are natural extensions
  of the same classifier and `oneof` renderer.
- The `oneof` group name is derived deterministically and is not configurable in v1; an
  `x-`extension override could be added if authors need to control it (it is not wire-significant,
  so this is purely ergonomic).

## Open Questions

- TODO: the exact derivation of the `oneof` group name (e.g. a constant token vs. derived from the
  schema name) — low-stakes because it is not wire-significant, but the two repos' golden tests must
  agree on whatever is chosen. Resolve at implementation time in openapi-schema.go; duh-cli only
  asserts a `oneof` is present, not its name.
