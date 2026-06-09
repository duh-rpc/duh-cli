# duh-cli Context

Domain glossary for `duh-cli`, the tool that generates DUH-RPC Go code, Protobuf
definitions, and HTTP clients/servers from OpenAPI specifications.

## Language

**Contract author**:
The engineer who owns a service's `openapi.yaml` and runs `duh lint` / `duh generate`.
_Avoid_: API owner, spec owner

**Consumer-service developer**:
An engineer in a different service who regenerates types from a provider's contract.
_Avoid_: client developer, downstream user

**Fieldmap lock**:
The checked-in `fieldmap.lock` file that pins JSON field/variant names to stable Protobuf
field numbers per message and enum.
_Avoid_: lockfile, fieldmap file, number map

**Append-only assignment**:
The rule that a new field or enum variant receives the next available number regardless of
its position in the OpenAPI schema.
_Avoid_: positional numbering, ordered assignment

**Reserved number**:
A Protobuf field/variant number whose field was removed; retained permanently and never
reused. In `fieldmap.lock` a reserved number is represented as a **tombstone** — the original
name→number entry kept with `reserved: true` — so the removal history stays visible in a diff.
_Avoid_: freed number, retired number, deleted number

**Seeding**:
The first `duh generate` run for a spec with no existing lock; it creates `fieldmap.lock` from
current OpenAPI declaration order. Enforces the ADR 0004 precondition that *when an integer enum
declares an `*_UNSPECIFIED` sentinel, that variant is declared first* (string enums and
sentinel-less integer enums are exempt).
_Avoid_: bootstrapping, initialization

**Open string enum**:
A `type: string` + `enum` schema. Generates an open proto `string` field with allowed values in a
comment; any string is valid on the wire (AIP-126 "may grow"). Not locked, has no `*_UNSPECIFIED`
sentinel.
_Avoid_: closed enum, string constant

**Closed proto enum**:
A `type: integer` + `enum` schema. Generates a Protobuf enum locked in **Fieldmap lock**; the first
value owns number 0. When it declares an `*_UNSPECIFIED` sentinel that variant must be declared
first so it owns 0 (the wire default for unset).
_Avoid_: integer enum, numeric enum

**FieldNumbers**:
The structured input `duh generate` passes to the `openapi-schema.go` library to drive proto
field/variant numbering from the lock (instead of positional assignment). Plain numbers plus
reserved lists; the library emits the resulting `reserved` statements.
_Avoid_: number map, lock map

**Contract directory**:
The checked-in, version-keyed directory holding a contract's `openapi.yaml`, generated
`schema.proto`, and `fieldmap.lock` (e.g. `api/v1/`).
_Avoid_: spec dir, api dir

**Output directory**:
The `--output-dir` target for regenerated Go/proto code; typically gitignored. Distinct from
the contract directory.
_Avoid_: gen dir, build dir

**Docs serve**:
The long-running `duh docs serve` mode — a local web server that renders the spec with
ReDoc and auto-reloads the browser on every spec save. The contract author's live
authoring loop.
_Avoid_: preview server, dev server

**Docs export**:
The one-shot `duh docs export` mode that writes a single self-contained HTML file of the
rendered spec for sharing. Distinct from **`duh generate`** (which emits Go/proto code);
"export" always refers to the docs HTML artifact, never code generation.
_Avoid_: build, render, generate (for the HTML artifact)

**Self-contained export**:
The exported HTML with the spec and the ReDoc renderer inlined so it renders fully
offline with no build step and no network. Works only for a single-file spec; external
`$ref`s are not bundled.
_Avoid_: standalone file, bundle

## Relationships

- A **Contract author** maintains one `openapi.yaml` and its co-located **Fieldmap lock** per
  version, inside a **Contract directory**.
- A **Fieldmap lock** maps each message field name and enum variant name to exactly one Protobuf
  number; removed names leave a **Reserved number**.
- `duh generate` writes generated code to the **Output directory** but writes the **Fieldmap
  lock** to the **Contract directory**.
- A **Consumer-service developer** regenerates from the provider's `openapi.yaml` **and**
  **Fieldmap lock** to stay wire-compatible.

## Example dialogue

> **Dev:** "I inserted `displayName` between `name` and `email` in the schema — does `email` keep
> proto number 2?"
> **Domain expert:** "Yes. The **Fieldmap lock** pins `name=1`, `email=2`. **Append-only
> assignment** gives `displayName` the next number, 3 — its position in the schema is irrelevant."
> **Dev:** "And if I later delete `displayName`?"
> **Domain expert:** "3 becomes a **Reserved number**. It's never reused, even if you add a new
> field with the same name."

## Flagged ambiguities

- "lock file" could mean the **Fieldmap lock** or an OS file lock — in this project it always
  means the **Fieldmap lock** (`fieldmap.lock`).
- "output" was used for both the **Output directory** (gitignored generated code) and the
  **Contract directory** (checked-in `fieldmap.lock`) — resolved: these are distinct locations,
  and the lock never goes in the output directory.
