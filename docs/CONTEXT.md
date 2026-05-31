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
current OpenAPI declaration order. Enforces the ADR 0004 precondition that each enum's first
declared variant is its `*_UNSPECIFIED`.
_Avoid_: bootstrapping, initialization

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
