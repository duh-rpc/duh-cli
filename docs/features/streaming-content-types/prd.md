# Streaming Content Types & PATH_FORMAT Domain Segments PRD

## Problem

There are two independent linter gaps that prevent valid DUH-RPC specs from passing `duh lint`:

**Streaming content types.** The DUH-RPC streaming spec (`duh.go/docs/streaming.md`) is finalized and defines three streaming content types: `application/octet-stream`, `application/duh-stream+json`, and `application/duh-stream+protobuf`. The `duh lint` CONTENT_TYPE rule only accepts `application/json` and `application/protobuf`, rejecting all three streaming types. This means valid streaming endpoints fail linting, and users must either skip the rule or accept false errors. The `BYTES_FORMAT` rule also blocks `format: binary` with a note that it is "reserved for streaming" — now that streaming is defined, this restriction should be lifted.

**PATH_FORMAT domain segments.** The linter documentation (`docs/duh-linter-rules.md`) describes two valid path forms: `/{resource}.{method}` and `/{domain}/{resource}.{method}`. However, the `PATH_FORMAT` rule implementation only accepts the single-segment form. Paths like `/billing/invoices.create` or `/admin/jobs.types` — which are valid according to the documented spec — are rejected. The regex `^/[a-z][a-z0-9_-]{0,49}\.[a-z][a-z0-9_-]{0,49}$` has no provision for an optional `/{domain}/` prefix, and the error message generation logic assumes exactly two parts (resource and method).

## Users

API designers writing OpenAPI specs for DUH-RPC services that include streaming endpoints (file downloads, binary exports, server-to-client structured streams) or that organize endpoints under domain prefixes (e.g., `/billing/invoices.create`, `/admin/jobs.types`).

## Mental Model

Streaming content types are a response-only extension to the existing content negotiation model. A streaming endpoint looks like any other DUH-RPC endpoint — POST-only, JSON/protobuf request body, standard error responses — except its success response uses a streaming content type instead of `application/json` or `application/protobuf`. The linter validates that streaming types appear only where they belong (responses) and that the rest of the endpoint follows standard DUH-RPC rules.

## Core Design Principles

- **Streaming types are response-only.** They must not appear in request bodies. Request bodies for streaming endpoints use `application/json` or `application/protobuf` like any other endpoint.
- **No special-casing of other rules.** Existing rules (`ERROR_SCHEMA`, `RESPONSE_STANDARD_NAME`, `REQUEST_BODY_REQUIRED`, etc.) apply to streaming endpoints the same way they apply to all endpoints. The response schema for a streaming endpoint defines the payload type carried by data/final frames.
- **No wire protocol validation.** The linter validates OpenAPI declarations. It does not validate frame formats, sequence fields, or streaming wire protocol details.

## Scope

### In Scope

- **CONTENT_TYPE rule**: Add `application/octet-stream`, `application/duh-stream+json`, and `application/duh-stream+protobuf` to the accepted content types, but only in responses. If any of these three types appear in a request body, it is a violation.
- **CONTENT_TYPE rule**: The existing requirement that `application/json` must appear in every request body is unchanged. There is no linter requirement for `application/json` in responses; streaming content types in responses are accepted without any JSON co-requirement.
- **BYTES_FORMAT rule**: Permit `format: binary` on schema properties when used alongside a streaming content type (`application/octet-stream`). The restriction that `format: binary` is "reserved for streaming" is resolved — streaming is now defined.
- **PATH_FORMAT rule**: Update the path validation regex to accept both `/{resource}.{method}` and `/{domain}/{resource}.{method}` forms. The `{domain}` segment follows the same character rules as `{resource}` and `{method}` (lowercase letters, numbers, hyphens, underscores; must start with a lowercase letter). Only a single optional domain segment is supported — deeper nesting (e.g., `/{a}/{b}/{resource}.{method}`) is not valid. The error message generation must also handle the three-part form correctly.
- **Linter rules documentation**: Update `docs/duh-linter-rules.md` to reflect the new accepted content types, the response-only constraint, and the `BYTES_FORMAT` change. Remove or update the "Note on `application/octet-stream`" paragraph in the CONTENT_TYPE section and the "reserved for streaming" language in the BYTES_FORMAT section. Correct the CONTENT_TYPE section to remove the claim that `application/json` is required in responses — the linter enforces this only for request bodies.
- **Tests**: Update the existing `InvalidOctetStreamContentType` test case and add new test cases covering the streaming content types in both valid (response) and invalid (request body) positions. Add test cases for `PATH_FORMAT` covering valid domain-prefixed paths (e.g., `/billing/invoices.create`), invalid domain segments (uppercase, special characters), and paths with too many segments.

### Out of Scope / Non-Goals

- **AI cheatsheet update** (`duh.go/docs/ai-cheatsheet.md`) — separate task.
- **Wire protocol validation** — no validation of frame format, flag bytes, length-prefix encoding, or sequence fields.
- **Bidirectional streaming** — the spec marks this as "not yet defined."
- **`application/octet-stream` in request bodies** — not supported for now. May be revisited when upload semantics are defined.
- **New lint rules** — no new rule names. Changes are extensions of existing `CONTENT_TYPE`, `BYTES_FORMAT`, and `PATH_FORMAT` rules.
- **Deeply nested domain paths** — only a single `/{domain}/` prefix is supported. Paths like `/{a}/{b}/{resource}.{method}` are not valid DUH-RPC paths.

## Dependencies and Constraints

- The DUH-RPC streaming spec (`duh.go/docs/streaming.md`) is the authoritative source for streaming content type semantics.
- The `CONTENT_TYPE` rule implementation is in `internal/lint/rules/content_type.go`. The `BYTES_FORMAT` rule does not yet exist and must be created in the same rules package.
- The `PATH_FORMAT` rule implementation is in `internal/lint/rules/path_format.go`. The regex and error message generation need to be updated to support the `/{domain}/{resource}.{method}` form that is already documented.
- The linter rules documentation (`docs/duh-linter-rules.md`) must stay in sync with the implementation.

## Open Questions

None. All product-level questions have been resolved.
