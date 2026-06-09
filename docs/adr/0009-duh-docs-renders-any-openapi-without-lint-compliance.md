# 9. duh docs renders any OpenAPI without requiring lint compliance

Date: 2026-06-08

## Status

Accepted

## Context

The `duh docs` subcommand renders an OpenAPI spec as browsable documentation. Every
other command in the tool — `lint`, `generate`, `add` — enforces or assumes DUH-RPC
conventions, so a reader would reasonably expect `docs` to gate rendering on DUH-RPC
compliance as well.

Gating rendering on compliance would prevent `docs` from displaying third-party specs,
in-progress specs still being authored, and any OpenAPI document that does not follow
DUH-RPC conventions — the documents a viewer is most often pointed at.

## Decision

We will render any OpenAPI document that parses, with no DUH-RPC compliance check. The
sole requirement is that the spec parses as OpenAPI. Compliance validation remains the
responsibility of `duh lint`.

## Consequences

- `docs` works on third-party, in-progress, and non-DUH-RPC specs.
- The rendered output is not a signal of DUH-RPC compliance; a spec can render cleanly
  yet fail `duh lint`.
- A spec that does not parse as OpenAPI cannot be rendered and is reported as an error.
