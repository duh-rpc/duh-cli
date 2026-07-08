# Did-You-Mean Teaching 404s Blueprint

Linear: [ENG-133](https://linear.app/kapetan-io/issue/ENG-133/duh-cli-unknown-endpoint-reply)

## Objective

A caller that guesses a wrong endpoint path on a DUH service today gets a 404 with no
usable guidance. The callers most likely to guess are LLM agents calling a service before
reading its spec, and the [ENG-132 API-intuition study] showed exactly where they fail:
CRUD-shaped names are guessed correctly 10/10, while non-CRUD names (`diffs.get`,
`revisions.compare`, `tree.list`) miss consistently — and a single `did_you_mean` hint in
the error recovers the caller in one retry.

This feature lets a contract author declare those likely wrong guesses in the OpenAPI spec.
`duh generate` turns each declared guess into a real, matched route in the generated
dispatch switch whose only job is to reply with a **teaching 404**: a DUH Reply naming the
canonical endpoint. One wrong guess costs exactly one retry — for every guess the author
predicted.

An earlier framing of this ticket proposed a runtime catch-all (a `default:` arm claiming
all unmatched paths under the service base path, with request-time fuzzy matching). That
was rejected: it breaks the fall-through contract that scaffold's RPC-handler chaining
depends on, drags in prefix-scoping and opt-out machinery, and puts a tunable heuristic on
the request path. Teaching at generate time from author-declared paths has none of those
problems. What scaffold itself replies for genuinely unmatched paths is a separate,
scaffold-layer concern — tracked as [ENG-134].

[ENG-132 API-intuition study]: https://linear.app/kapetan-io/issue/ENG-132/git-server-graph-api
[ENG-134]: https://linear.app/kapetan-io/issue/ENG-134/scaffold-duh-reply-on-unmatched-requests

## Mental Model

A did-you-mean path is a **real route that teaches**. It is matched by the generated
switch exactly like a canonical route — but instead of dispatching to a handler, it
replies 404 with a hint naming the one canonical path the author mapped it to. It is not
an alias, not a redirect, and never serves the operation.

Everything the switch does not match — canonical or teaching — falls through
(`return false`) exactly as today. Teaching coverage equals the declared table, nothing
more: an unpredicted guess is out of this feature's hands by design.

## Core Design Principles

1. **Correct the caller, never accommodate the guess.** A declared wrong-guess path always
   replies 404 + hint and never serves the request. The moment a wrong path silently
   works, it becomes a second, undocumented API surface that can never be removed. One
   canonical path per operation.
2. **The fall-through contract is inviolate.** Generated `ServeHTTP` returns `false` for
   any unmatched path, unconditionally. Scaffold chains multiple RPCHandlers on one
   binding and falls through to a mux; that composition depends on this contract, and no
   teaching mechanism may touch it.
3. **Teaching is spec-declared and deterministic.** No heuristics at generate time or
   request time. The generated output is a pure function of the spec; the author owns the
   guess table and maintains it like any other part of the contract (observe 404s →
   declare → regenerate).
4. **Zero cost to real traffic.** Canonical routes gain no new code on their dispatch
   path; teaching paths are additional `case` arms in the existing switch.

## Correctness Constraints

### State Invariants

These are invariants of the generated artifact (the feature has no runtime state):

- **Every `case` path in the generated switch is unique** — canonical routes and teaching
  paths are disjoint sets, and teaching paths are pairwise distinct across the spec.
  Violated by: an author declaring a did-you-mean path equal to a real route, or the same
  did-you-mean path declared more than once — within a single path item's list or across
  path items. Enforcement: `duh generate` fails with an
  error naming the colliding paths (and `duh lint` reports it earlier). Structural
  backstop: duplicate `case` literals in a Go switch do not compile, and generated output
  is compile-verified in tests.
- **Every `did_you_mean` value names a canonical route of the same spec.** Enforcement is
  structural: the generated hint is the declaring operation's own route-path constant,
  emitted from the same parsed operation record — there is no second source to drift from.

### Behavioral Constraints

- **A teaching path never serves.** Its `case` arm contains only the reply and
  `return true`; there is no dispatch to any handler.
- **An unmatched path never gets a reply from generated code.** The switch's miss behavior
  (`return false`) is byte-identical to today.
- **A spec that does not use the extension generates the same server code as before this
  feature.** Adoption is strictly opt-in per spec.

Reversibility: removing a did-you-mean entry and regenerating returns that path to
ordinary fall-through. Nothing persists.

## Acceptance Criteria

- Given a spec where path `/commits.diff` declares `x-duh-did-you-mean: [/diffs.get]` and
  `servers[0].url` has path `/v1`, a request to `/v1/diffs.get` on the generated handler
  returns HTTP 404 with a DUH Reply: `code` 404, `message` `no such endpoint:
  /v1/diffs.get`, `details` exactly `{"did_you_mean": "/v1/commits.diff"}` — and
  `ServeHTTP` returns `true`.
- The teaching reply is returned for any HTTP method on the teaching path (a guessed
  endpoint does not exist; method is irrelevant).
- A request to a path matching no case still causes `ServeHTTP` to return `false` and
  writes nothing to the response.
- Canonical routes behave exactly as before, including their method check and dispatch.
- A spec without any `x-duh-did-you-mean` generates server code with no teaching arms and
  no other new constructs.
- `duh generate` exits non-zero, naming both paths, when a did-you-mean path equals a
  canonical path in the spec.
- `duh generate` exits non-zero, naming the path and its declaring operation(s), when the
  same did-you-mean path is declared more than once — within a single path item's list or
  across path items.
- `duh generate` exits non-zero, naming the offending path, when a did-you-mean entry is
  malformed (non-string, not starting with `/`, or containing a character that cannot appear in
  a path such as a quote, backslash, or newline).
- `duh generate` exits non-zero, naming the path, when `x-duh-did-you-mean` is declared on a
  path item that has no `post:` operation for it to teach (its entries would otherwise be
  silently dropped from the switch).
- `duh lint` reports the same two collision cases as `ERROR` violations, and reports a
  malformed entry (non-string, or not starting with `/`) as an `ERROR`.
- Generated code containing teaching arms compiles against the pinned duh.go runtime
  (existing compile-verify harness).

## Scope

### In Scope

- The `x-duh-did-you-mean` path-item extension: parsing, validation, and code generation
  of teaching `case` arms in `server.go`.
- Collision and format validation in `duh generate` (hard error) and `duh lint` (new rule).
- Documentation: extension reference in `docs/duh-openapi-reference.md`, rule entry in
  `docs/duh-linter-rules.md`.

### Out of Scope / Non-Goals

- **Any catch-all or default-arm behavior.** Unmatched paths fall through untouched. What
  the binding ultimately replies is scaffold's concern ([ENG-134]).
- **Mechanical guess derivation** (plural/singular, verb synonyms, subject/action
  inversion). Considered and rejected for now: only the author knows the semantic misses
  that matter, and derived tables make generated output harder to predict. Revisit only
  with evidence.
- **Runtime fuzzy matching** in any form, in generated code or in duh.go.
- An `endpoints` listing in the reply, an exported operation-path table, per-service
  opt-out extensions, and teaching-404 observability counters — all removed from the
  original ticket's proposal during design; access logs already cover the author's
  observe-and-declare loop.
- Client-side changes: generated clients ignore the extension entirely.

## Dependencies and Constraints

- No new dependencies. `duh.ReplyWithCode(w, r, code, details, msg)` and
  `duh.CodeNotFound` exist in duh.go v2.4.0, already the pinned runtime.
- No scaffold changes required or permitted by this feature.

---

## Functional

### The extension

`x-duh-did-you-mean` sits at the **path-item level**, beside the path's `post:` operation.
Its value is a sequence of strings, each a spec-relative path beginning with `/` (the same
form as the spec's own path keys; the server base path from `servers[0].url` is prepended
exactly as for canonical routes):

```yaml
paths:
  /commits.diff:
    x-duh-did-you-mean:
      - /diffs.get
      - /revisions.compare
    post:
      ...
```

Teaching paths are guesses, not contract paths — they are exempt from the naming lint
rules that govern canonical paths. They must only be well-formed (a string beginning with `/`,
free of characters that cannot appear in a path), sit beside a `post:` operation, and be
collision-free (see Correctness Constraints).

The extension affects `server.go` generation only. Clients, protobuf/fieldmap outputs, and
`duh docs` rendering are unaffected (ReDoc ignores unknown `x-` extensions).

### The generated arm

For each declared path, the switch gains one arm (shown for the example above; teaching
arms follow the canonical arms in the switch):

```go
case "/v1/diffs.get":
	duh.ReplyWithCode(w, r, duh.CodeNotFound,
		map[string]string{"did_you_mean": RPCCommitsDiff},
		"no such endpoint: /v1/diffs.get")
	return true
```

The hint references the declaring operation's existing route constant, not a second string
literal. There is no method check: any method on a teaching path receives the teaching
404.

### The wire reply

`duh.ReplyWithCode` performs the same content negotiation as every other generated reply
(JSON default, protobuf on request). JSON form:

```json
{
  "code": 404,
  "message": "no such endpoint: /v1/diffs.get",
  "details": {"did_you_mean": "/v1/commits.diff"}
}
```

`did_you_mean` is always exactly one path. The reply never enumerates other endpoints.

## Architecture

Three touch points in the existing generate pipeline, described as contracts (internal
signatures are the implementor's):

- **Parser** (`internal/generate/duh/parser.go`): extracts the extension from each path
  item into the operation record alongside the existing route-path data, resolving each
  entry to its full route path (base + declared path). Malformed entries (non-string,
  missing leading `/`) are generate errors.
- **Validation**: before any output is written, the collected teaching paths are checked
  for the two collision classes (equal to a canonical route; the same teaching path
  declared more than once, whether within a single path item's list or across path items).
  Errors name every offending path so the author can fix the spec in one pass.
- **Template** (`internal/generate/duh/templates/server.go.tmpl`): emits one teaching arm
  per declared path after the canonical arms. When no operation declares the extension,
  the template's output is unchanged from today.

`duh lint` gains one rule, `DID_YOU_MEAN_COLLISION` (`ERROR`), covering both collision
classes and entry format, so authors get feedback before generating. Generate does not
depend on lint having run; it enforces independently.

### Invariant Preservation

- *Switch-path uniqueness*: enforced by generate-time validation; backstopped structurally
  by the Go compiler (duplicate `case` literals are a compile error) and the existing
  compile-verify tests. This invariant relies on application logic plus a structural
  backstop, so both collision classes carry explicit acceptance tests.
- *Hint names a real route*: structural — emitted from the declaring operation's own
  record; illegal state unrepresentable in the template's inputs.
- *Fall-through untouched*: structural — the template's miss path is not modified by this
  feature. The no-extension acceptance test pins that no teaching constructs appear in the
  generated text; runtime fall-through itself (`return false`, nothing written) is pinned by
  the unmatched-path acceptance test against the compiled handler.

## Security

Teaching replies disclose only paths the author deliberately wrote into the spec — the
same information the spec itself publishes. The reply message embeds the matched `case`
literal (a spec-derived constant), never attacker-controlled request input, so there is no
reflection or injection surface.

## PII

None. No request data is stored, logged, or echoed beyond the spec-derived path literal.

## Scale

Dispatch is a Go string switch; teaching arms add cases but no per-request allocation or
new code on canonical-route dispatch. Realistic tables (tens of entries) are negligible.

## Testing

Testing follows the `surface-testing` skill. All tests drive `duh.RunCmd` and, for runtime
behavior, the compile-verified generated handler (the established
`server_test.go` / `full_test.go` harness). No fakes, clocks, or async observability are
needed — `ServeHTTP`'s boolean return makes fall-through directly assertable.

Key surfaces:

- [integration: `duh generate` on a spec with `x-duh-did-you-mean` → generated handler
  answers the teaching path with the exact Reply JSON and `true`; unmatched path writes
  nothing and returns `false`; canonical routes unaffected]
- [integration: `duh generate` on collision specs (path equals canonical; duplicate across
  path items; malformed entry) → non-zero exit, error names the paths]
- [integration: `duh generate` on a spec without the extension → no teaching arms in
  output]
- [integration: `duh lint` on the same collision/malformed specs → `DID_YOU_MEAN_COLLISION`
  ERROR violations (rule-test style of `internal/lint/lint_test.go`)]
- [integration: compile-verify generated teaching code against the pinned duh.go runtime]

## Limitations & Future Work

- Teaching coverage equals the declared table. Unpredicted guesses fall through to
  whatever the deployment's terminal behavior is; making that terminal response a proper
  DUH Reply is scaffold's job ([ENG-134]).
- The author's maintenance loop (observe 404s → declare → regenerate) is manual by
  design. If evidence accumulates that authors want help, tooling that suggests
  candidates from access logs could be layered on without changing the extension's shape.
