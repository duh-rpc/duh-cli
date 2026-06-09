# 10. Open string enums and closed integer enums are the two enum modes

Date: 2026-06-08

## Status

Accepted

## Context

DUH projects an OpenAPI enum to Protobuf in one of two ways, switched by the schema's `type`:

- `type: string` + `enum` generates an open proto `string` field. The allowed values appear only as
  a comment; any string is valid on the wire, and the field is not pinned in `fieldmap.lock`.
- `type: integer` + `enum` generates a closed Protobuf enum whose first value owns wire number `0`.

A prior version of the conversion library generated all enums, string and integer alike, as
Protobuf enums. That broke JSON wire compatibility for string-valued enumerations, so the library
changed to generate string enums as string fields. Some style guides (e.g. Google
[AIP-126](https://google.aip.dev/126)) likewise model string-valued enumerations that may grow as
plain strings rather than closed enums.

The `ENUM_UNSPECIFIED_VARIANT` lint rule was written for the old model: it required every enum's
first variant to be an `*_UNSPECIFIED` sentinel regardless of `type`. After the library change it
raised unsatisfiable errors — on string enums (which have no Protobuf enum and no number `0`) and on
bare numeric integer enums (whose variant literals are numbers, never `*_UNSPECIFIED`) — so it was
disabled in practice. `duh generate`, by contrast, rejects an enum only when it declares an
`*_UNSPECIFIED` sentinel that is not the first declared variant; lint and generate had diverged.

## Decision

We will treat the two generation behaviors above as the only two enum authoring modes — the open
string enum (`type: string`) and the closed proto enum (`type: integer`) — and add no third. In
particular, no closed or value-validated string-enum mode. A closed proto enum that declares an
`*_UNSPECIFIED` sentinel must declare it first.

`duh lint` will mirror `duh generate`: the `ENUM_UNSPECIFIED_VARIANT` error fires only on a
`type: integer` enum whose `*_UNSPECIFIED` sentinel is not first, and a new advisory warning,
`ENUM_STRING_SENTINEL_NAMES`, flags a `type: string` enum that uses `*_UNSPECIFIED` naming as a
likely wrong-`type` mistake.

## Consequences

- String enums and bare numeric integer enums lint clean; the rule no longer needs disabling, which
  restores its value.
- Lint and generate accept and reject exactly the same enum specs, so a spec that lints clean seeds
  a lock without a generation-time error.
- Contract authors must choose `type` deliberately between an extensible string set and a wire-safe
  closed enum; the advisory warning catches the common wrong-`type` mistake.
- Parity rests on a shared sentinel definition ("name ends in `UNSPECIFIED`") duplicated across the
  lint and fieldmap packages; changing one without the other silently breaks parity.
- Authors who want generated code to reject a string field whose value is outside the listed set get
  no such enforcement; for string enums the listed values remain advisory.
