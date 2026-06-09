# Lint Enum Sentinel Scope Blueprint

Linear: ENG-73 · ADRs: [0004](../../adr/0004-duh-generate-requires-unspecified-as-first-enum-variant.md), [0007](../../adr/0007-enum-variants-keyed-by-literal-openapi-value.md)

## Objective

`duh lint`'s `ENUM_UNSPECIFIED_VARIANT` rule raises a hard error on every enum whose first
variant name does not end in `UNSPECIFIED`, regardless of the enum's `type`. This is stale: it
predates the `openapi-schema.go` breaking change (commit `55eb0d7`) after which only **integer**
enums become Protobuf enums, while **string** enums generate plain proto `string` fields. As a
result the rule fires on cases the generator accepts and that have no zero-value concern at all:

- `type: string` + `enum` (the case in the ticket) — generates an open proto `string` field; there
  is no Protobuf enum and no number `0` to protect, yet the rule demands an `UNSPECIFIED` sentinel.
- bare numeric `type: integer` + `enum` (e.g. `[200, 404, 500]`) — a legitimate proto enum whose
  first value takes number `0`; the rule's name-suffix check is unsatisfiable because the variant
  literals are numbers.

Because the error cannot be satisfied in these cases, the de-facto workaround is to disable the
rule, which undermines its value. The goal is to make the rule fire **only** where the underlying
generator actually constrains the author, and to replace the proto-authoring error text with
guidance phrased in OpenAPI terms a contract author can act on.

This feature also adds a small advisory warning that targets the confusion behind the ticket: a
`type: string` enum written with proto-enum sentinel naming is almost always an author who meant a
closed proto enum and reached for the wrong `type`.

## Mental Model

DUH projects OpenAPI enums to Protobuf two different ways, and the `type` field is the switch:

| OpenAPI declaration | Generates | Sentinel relevant? |
|---|---|---|
| `type: string` + `enum: [earliest, latest]` | open proto `string` field; allowed values in a comment. Any string is valid on the wire (AIP-126 "may grow"). | No — there is no proto enum and no number `0`. |
| `type: integer` + `enum: [200, 404, 500]` | closed proto enum; first value → number `0` (`CODE_200 = 0`), any name. | No — a bare integer enum legitimately puts a real value at `0`. |
| `type: integer` + `enum: [CODE_UNSPECIFIED, CODE_OK, …]` (named values; legal OpenAPI) | closed proto enum keyed by literal value; the sentinel owns number `0`. | **Yes** — this is the only shape where an `*_UNSPECIFIED` sentinel applies. |

The two authoring modes the contract author chooses between:

- **Open string enum** — `type: string` + `enum`. The set may grow; the wire carries the literal
  string. No sentinel.
- **Closed proto enum** — `type: integer` + named variants with `*_UNSPECIFIED` first. Wire-safe:
  number `0` is reserved for "unset" so an absent field never decodes as a real value (ADR 0004).

## Correctness Constraints

### State Invariants

- **Lint ↔ generate parity.** `duh lint` must reject exactly the enum specs `duh generate` rejects,
  and accept the rest. `duh generate` (via `fieldmap.assertUnspecifiedFirst`) hard-errors only when
  an enum *declares* an `*_UNSPECIFIED` sentinel that is not its first declared variant; it accepts
  string enums (not locked) and sentinel-less integer enums (first value legitimately `0`).
  Preservation: the lint rule and `assertUnspecifiedFirst` share the same sentinel predicate
  (`strings.HasSuffix(value, "UNSPECIFIED")`) and the same first-position test, so a spec passes one
  iff it passes the other. Violation of parity is the bug this feature exists to remove.

### Behavioral Constraints

- The error-severity rule must **never** flag a spec that `duh generate` would successfully
  generate. (False positives are the defect; this constraint rules out the current behavior.)
- The new advisory rule must be **WARNING** severity only — it must never block generation, because
  a `type: string` enum with sentinel-style names is valid and generates correctly (an open string
  field). It is guidance, not a gate.

## Acceptance Criteria

Each is verifiable through `duh lint` exit code and stdout (surface tests, `RunCmd`):

1. A `type: string` enum without a sentinel (`[earliest, latest, at_offset]`) → no
   `ENUM_UNSPECIFIED_VARIANT` error (the ticket's repro is compliant).
2. A `type: string` enum with sentinel-style names (`[STATUS_UNSPECIFIED, STATUS_ACTIVE]`) → no
   `ENUM_UNSPECIFIED_VARIANT` error, **and** an `ENUM_STRING_SENTINEL_NAMES` warning is emitted.
3. A bare numeric `type: integer` enum (`[200, 404, 500]`) → no `ENUM_UNSPECIFIED_VARIANT` error.
4. A `type: integer` named enum with the sentinel first (`[CODE_UNSPECIFIED, CODE_OK]`) → compliant.
5. A `type: integer` named enum with the sentinel **not** first (`[CODE_OK, CODE_UNSPECIFIED]`) →
   `ENUM_UNSPECIFIED_VARIANT` error.
6. A `type: integer` named enum with no sentinel (`[CODE_OK, CODE_ERR]`) → compliant (the rule does
   not nudge; it mirrors `fieldmap`).
7. The same mis-ordered sentinel on an inline integer property → flagged.
8. A free-form `type: string` with no `enum` → never flagged by either rule.
9. The `ENUM_STRING_SENTINEL_NAMES` warning alone yields exit code 0 (warnings do not fail lint).

## Scope

### In Scope

- Re-scope `ENUM_UNSPECIFIED_VARIANT` to fire only on `type: integer` enums that declare an
  `*_UNSPECIFIED` sentinel not in first position; rewrite its message/suggestion in OpenAPI terms.
- Add the `ENUM_STRING_SENTINEL_NAMES` advisory rule (WARNING).
- Rewrite `internal/lint/rules/enum_unspecified_variant_test.go` to the model above.
- Remove the now-redundant `--disable ENUM_UNSPECIFIED_VARIANT` and `x-duh-lint-ignore` workarounds
  (and their now-inaccurate comments) from `internal/generate/duh/fieldmap_lock_test.go` helpers.
- Reconcile docs: the `docs/CONTEXT.md` "Seeding" entry (currently states the sentinel-first rule
  unconditionally) and ADRs 0004/0007 (already corrected in the working tree).

### Out of Scope / Non-Goals

- No change to `duh generate` or the `openapi-schema.go` library; lint is brought in line with the
  generator's existing behavior, not vice versa.
- No opt-in "closed string enum" feature. `type: string` + `enum` stays an open string field; the
  closed/wire-safe path remains `type: integer` with named variants. (Item #2 resolved: keep open,
  document the two modes.)
- The advisory warning's trigger is narrow — only a `type: string` enum containing a value ending in
  `UNSPECIFIED`. Broadening to all-`SCREAMING_SNAKE_CASE` values is explicitly not done (too noisy:
  some APIs legitimately use uppercase string values on the wire).
- The external `openapi-schema.go` `docs/enums.md` drift (its own code vs docs) is that repo's
  concern, not this one.
- The mono-repo's `services/demo` and `services/slip-stream` specs are downstream consumers; they
  resolve when they pick up the rule change, and are not edited here.

## Dependencies and Constraints

- `github.com/pb33f/libopenapi` — schema model (`base.Schema.Type`, `.Enum []*yaml.Node`).
- Behavioral coupling to `internal/fieldmap` (`isUnspecifiedSentinel`, `assertUnspecifiedFirst`,
  `checkEnumInvariant`) and to `openapi-schema.go` enum projection. The parity invariant depends on
  these continuing to define the sentinel as "name ends in `UNSPECIFIED`" and integer enums as the
  only locked proto enums. A change there is a change here.

---

## Functional

Two rules in `internal/lint/rules/`, registered with the validator like the existing rules.

### `ENUM_UNSPECIFIED_VARIANT` (ERROR) — re-scoped

Fires for a component-schema enum or an inline property enum when **all** hold:

- the schema is an integer enum: `len(schema.Type) > 0 && schema.Type[0] == "integer"`;
- some variant value is a sentinel (`strings.HasSuffix(value, "UNSPECIFIED")`);
- the first variant value is not a sentinel.

Otherwise it does not fire. (Empty `enum`, string enums, and sentinel-less integer enums are all
skipped.) Existing `$ref`-skipping and `x-duh-lint-ignore` handling are unchanged.

- **Message:** `Enum declares an *_UNSPECIFIED sentinel but its first variant is '<value>'`
- **Suggestion:** `Move the *_UNSPECIFIED variant to the first position so it owns Protobuf number 0 (the wire default for unset)`

### `ENUM_STRING_SENTINEL_NAMES` (WARNING) — new

Fires for a `type: string` enum (component schema or inline property) when any variant value ends
in `UNSPECIFIED`. Disjoint from `ENUM_UNSPECIFIED_VARIANT` by construction (that rule requires
`type: integer`), so the two never double-fire on one schema.

- **Message:** `String enum variant '<value>' uses proto-enum sentinel naming, but a string enum generates an open proto string field`
- **Suggestion:** `For a closed, wire-safe proto enum use type: integer with named variants; to keep an open string field, drop the *_UNSPECIFIED sentinel`

## Architecture

Both rules follow the existing `Rule` interface (`Name()`, `Validate(*v3.Document) []Violation`) and
the established iteration pattern in `enum_unspecified_variant.go`: walk `Components.Schemas`, check
each top-level enum schema, then each non-`$ref` property's inline enum. The sentinel predicate is a
local helper mirroring `fieldmap.isUnspecifiedSentinel`; the two definitions are intentionally
identical and the parity invariant documents the coupling. The new rule is registered in the same
place the other rules are wired into the validator.

## Testing

Testing follows the `surface-testing` skill.

Key surfaces:
- integration: `duh.RunCmd(ctx, &stdout, []string{"lint", specPath})` — assert exit code and stdout
  for each acceptance criterion above. Table-driven with `t.Run` per the repo's lint-test
  convention.
- The `internal/generate/duh/fieldmap_lock_test.go` enum tests run with all rules enabled after the
  workaround removal — they double as regression coverage that bare integer enums and sentinel-first
  integer enums pass un-disabled.

No new fakes, no async behavior, no clock — pure function of the parsed document.

## Limitations & Future Work

- The parity invariant is enforced by shared *convention* (identical sentinel predicate in two
  packages), not by a shared symbol. If someone changes one definition without the other, parity
  silently breaks. A future refactor could export a single sentinel predicate consumed by both
  `lint` and `fieldmap`.

## Open Questions

- None blocking. Rule name `ENUM_STRING_SENTINEL_NAMES` is the proposed name; trivially renamed
  before implementation if preferred.
