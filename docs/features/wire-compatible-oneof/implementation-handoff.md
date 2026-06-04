# Implementation Handoff — Wire-Compatible oneOf (Style B)

This brief gives a downstream implementation session everything it needs to pick up one of the
tickets under **Linear ENG-60** without the discovery conversation that produced the design. Read
the canonical references below first; they are authoritative and the decisions in them are **not**
to be re-litigated.

## What is being built

The DUH toolchain currently rejects all `oneOf`. There is one `oneOf` form that is provably
wire-compatible across the JSON and protobuf wires — **style B** — and this feature makes the
toolchain accept it and generate a protobuf `oneof` for it. The flat/discriminated `oneOf` stays
prohibited.

**Style B** — an object with one optional `$ref` property per variant, plus a `oneOf` of
single-`required` branches and **no `discriminator`**:

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

Wire shape: `{"cat_event": {"pet_name": "Whiskers"}}` — nested, key-tagged, byte-identical to a
protobuf `oneof`. The **mechanical rule**: `oneOf` + `discriminator` (or `$ref`/inline variants) =
flat = **prohibited**; `oneOf` of `required` over named properties, no discriminator = nested =
**supported**.

## Canonical references (read these first — they live across THREE repos)

A session rooted in one repo will not see the others by default. Use absolute paths.

| File | Repo | Purpose |
|---|---|---|
| `/Users/thrawn/Development/duh-cli/docs/features/wire-compatible-oneof/blueprint.md` | duh-cli | **Authoritative spec.** Objective, correctness constraints, acceptance criteria, per-repo functional design. |
| `/Users/thrawn/Development/duh.go/docs/adr/0002-oneof-permitted-only-as-nested-key-tagged-union.md` | duh.go | The decision + rationale (why nested-only, why flat is impossible). |
| `/Users/thrawn/Development/duh-cli/docs/oneof-wire-compatibility.html` | duh-cli | Human explainer of the wire formats and the two supported styles. |
| `/Users/thrawn/Development/duh-cli/docs/oneof-wire-proof/` | duh-cli | The proven round-trip harness (`union.proto`, `proof_test.go`) — model the ENG-63 regression test on it. |

openapi-schema.go has **no** CONTEXT.md or ADRs of its own; its domain context is the files above.

## Ticket map and ordering

```
ENG-60  Support wire-compatible oneOf (style B)            [parent, In Progress]
  ├─ ENG-61  openapi-schema.go — emit a protobuf oneof     [do FIRST — blocks the others]
  ├─ ENG-62  duh-cli — narrow PROHIBITED_ONEOF + spec.md   [after ENG-61 is released]
  └─ ENG-63  duh.go — standing wire-compat regression test [after ENG-61]
```

- https://linear.app/kapetan-io/issue/ENG-61
- https://linear.app/kapetan-io/issue/ENG-62
- https://linear.app/kapetan-io/issue/ENG-63

**ENG-61 must land and be released first.** Narrowing the duh-cli linter (ENG-62) before the library
generates a `oneof` would let a style-B spec pass `lint` but fail `generate` — moving the failure,
not fixing it. ENG-62 must bump its openapi-schema.go dependency to the released version before
shipping. ENG-62 and ENG-63 are independent of each other.

## Decisions locked (do not reopen)

- Style B maps to a **real protobuf `oneof`**, not plain optional fields. (Plain optional fields
  would make style A and B generate identical proto and contradict the explainer.)
- **Scope v1:** top-level component schemas with `$ref` (message-typed) variants only. Scalar-typed
  variants and nested `oneOf` are **out of scope** — defer, don't implement.
- Linter: **narrow** `PROHIBITED_ONEOF` to the flat form and **delete** the four `DISCRIMINATOR_*`
  rules (they would false-positive on style B; docs already mark them removed).
- duh.go needs **no production change**; its ticket is a regression test.
- The `oneof` group **name is not wire-significant** and is **not** recorded in `fieldmap.lock`.

## Invariants the implementation must not break

- **Nested wire shape preserved.** Generated proto, marshaled via `protojson` (duh.go's marshaler,
  bare defaults), must emit `{"<variant>": {...}}` with the variant's snake_case property name as the
  key. This relies on the existing `[json_name = "<property>"]` emission — do not drop it, and do not
  enable lowerCamelCase output.
- **`oneof` members are ordinary numbered fields.** They get fieldmap-lock numbers from the normal
  property loop; append-only assignment and reserved tombstones apply unchanged. The `oneof`
  grouping must never renumber or reuse a member's number. Removing a variant → `reserved`.
- **No `repeated` inside a `oneof`** (proto3 forbids it) — a variant property whose schema is an
  array must be rejected at validation, not discovered by `protoc`.
- **Never accept the flat/discriminated `oneOf`,** and never route style B to the Go-struct path
  (which produces flat JSON the duh-cli wrapper rejects).

---

## Per-session kickoff prompts

### ENG-61 — start a session in `openapi-schema.go`

> Implement Linear ENG-61: emit a protobuf `oneof` for style-B unions. It is the first of three
> tickets under ENG-60 and blocks the others.
>
> Read first (authoritative, cross-repo absolute paths):
> - Blueprint: `/Users/thrawn/Development/duh-cli/docs/features/wire-compatible-oneof/blueprint.md`
> - Decision/rationale: `/Users/thrawn/Development/duh.go/docs/adr/0002-oneof-permitted-only-as-nested-key-tagged-union.md`
> - Proof harness to model tests on: `/Users/thrawn/Development/duh-cli/docs/oneof-wire-proof/`
>
> The change is in THIS repo, `internal/proto/builder.go`: classify the style-B pattern (no
> discriminator; every `oneOf` branch is `{required: [oneProperty]}` over a declared property),
> route it into `buildMessage` instead of marking it a Go union, and add `oneof` rendering to the
> proto generator (it never emitted `oneof` before). Reject malformed style B (>1 `required` per
> branch, `required` naming an undeclared property, array-typed variant). `oneof` members keep their
> normal field numbers and `json_name`.
>
> Use the `plan-from-spec` skill with the blueprint as the authoritative spec, then implement.
> Tests are golden tests on `Convert`: emitted `.proto` contains the `oneof` group + `json_name`;
> `fieldmap.lock` numbers variants append-only including a remove-variant → `reserved` case.
> Out of scope: scalar variants, nested `oneOf`, removing the existing discriminated→Go path.

### ENG-62 — start a session in `duh-cli` (also edits `duh.go` docs)

> Implement Linear ENG-62: allow style B in the linter and reconcile the spec. **Prerequisite:**
> ENG-61 must be released; bump the openapi-schema.go dependency to that version first so `lint` and
> `generate` agree.
>
> Read first: `/Users/thrawn/Development/duh-cli/docs/features/wire-compatible-oneof/blueprint.md`
> and `/Users/thrawn/Development/duh.go/docs/adr/0002-oneof-permitted-only-as-nested-key-tagged-union.md`.
>
> In duh-cli: narrow `PROHIBITED_ONEOF` (`internal/lint/rules/rpc_prohibited_oneof_and_allof.go`) to
> fire only on the flat form (has a `discriminator`, or `$ref`/inline variants); let style B pass;
> give malformed style B a clear non-zero-exit violation. Delete the four `DISCRIMINATOR_*` rules
> (rule files, tests, testdata, `validator.go` registration). Rewrite the `PROHIBITED_ONEOF` section
> of `docs/duh-linter-rules.md`.
>
> In duh.go (cross-repo): rewrite `docs/spec.md`'s "oneOf — Discriminated Unions Only" section to
> permit only style B and prohibit the discriminated/flat form; keep `allOf`/`anyOf` prohibited.
> Update `docs/ai-cheatsheet.md` and `docs/design-spec-alignment.md` where they restate the rule.
>
> Tests via `duh.RunCmd` only (per CLAUDE.md): style-B spec lints clean and generates a `.proto`
> with `oneof`; flat spec still fails `[PROHIBITED_ONEOF]`; each malformed style-B variant fails;
> no `DISCRIMINATOR_*` violation ever emitted.

### ENG-63 — start a session in `duh.go` (or openapi-schema.go)

> Implement Linear ENG-63: a standing wire-compatibility regression test for style B. **Prerequisite:**
> ENG-61 generates the proto. No production code change — duh.go's bare `protojson` defaults are
> already proven compatible; this guards against regressions.
>
> Model it on `/Users/thrawn/Development/duh-cli/docs/oneof-wire-proof/` (`union.proto`,
> `proof_test.go`), re-pointed at GENERATED style-B output. Assert: marshaled style-B value →
> nested `{"<variant>": {...}}` with snake_case keys; bytes semantically equal to the optional-fields
> encoding; cross-unmarshal between the `oneof` and optional forms (JSON and binary); a real `oneof`
> rejects two-variant input. Land it as a permanent test (duh.go, or alongside the openapi-schema.go
> golden tests — implementer's call).

## Definition of done (whole feature)

A spec containing a valid style-B schema lints clean and `duh generate` emits a `.proto` with a
`oneof`; a flat/discriminated `oneOf` still fails lint; malformed style B fails with a clear message;
the four `DISCRIMINATOR_*` rules are gone; `spec.md` matches ADR-0002; and a standing test proves the
generated output is wire-compatible on both wires. Full acceptance criteria are in the blueprint.
