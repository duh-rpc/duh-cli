# 11. Wrong-guess endpoints are spec-declared teaching 404s

Date: 2026-07-08

## Status

Accepted

## Context

Callers who guess a DUH endpoint path before reading the spec — most often LLM agents —
get a 404 with no guidance and no machine-readable way to recover. An empirical study of
agents guessing endpoints from capability descriptions showed non-CRUD operation names
miss consistently, and that a single corrective hint in the error recovers the caller in
one retry.

The generated server must answer predictable wrong guesses with such a hint. The forces:

- The dispatcher's fall-through on unmatched paths is load-bearing — the scaffold
  framework relies on it to chain handlers on one binding and route to a mux. Code that
  claims a path it does not own breaks that chaining and shadows mux routes.
- A runtime catch-all with request-time fuzzy matching would need prefix-scoping and
  per-spec opt-out machinery for composed deployments, and puts a tunable heuristic on
  the request path whose behavior changes are invisible in a diff.
- A wrong path that silently serves the intended operation becomes a second, undocumented
  API surface that callers depend on and that can never be removed.
- Only the contract author knows the semantic misses that matter (e.g. a caller guessing
  `/revisions.compare` for `/commits.diff`); mechanically derived guess tables make
  generated output unpredictable.

## Decision

We will teach wrong guesses at generate time, from author-declared paths only. A path
item lists its likely wrong guesses in an `x-duh-did-you-mean` extension; `duh generate`
turns each into a real, matched route that replies 404 with a `did_you_mean` detail
naming the single canonical path — and never serves the request.
Unmatched paths keep falling through untouched. A declared path that collides with a
canonical route or with another declaration fails generation and lint.

## Consequences

- Teaching coverage equals the author-declared table. Unpredicted guesses fall through to
  the deployment's terminal behavior, which remains the serving framework's concern.
- Canonical routes gain no request-time cost, and no new runtime dependency is introduced;
  the reply uses the existing DUH error-reply primitives.
- The spec remains the single source of truth: generated hints are emitted from the
  declaring operation's own route table and cannot drift from it.
- New wrong-guess coverage requires manually declaring the guess and regenerating;
  nothing detects undeclared misses automatically.
- Removing a declared guess and regenerating restores ordinary fall-through for that
  path; nothing persists at runtime.
