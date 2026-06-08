# 8. ReDoc is embedded, not loaded from a CDN

Date: 2026-06-08

## Status

Accepted

## Context

The `duh docs` subcommand renders OpenAPI specs as browsable documentation using
ReDoc, a browser-side renderer distributed as a standalone JavaScript bundle. The
renderer can be delivered two ways: referenced from a CDN with a `<script src>` tag, or
shipped inside the binary.

Two usage modes constrain the choice:

- The HTML export must render with no network available. Its recipients open the file
  directly, run no build step, and may be offline.
- The serve mode runs inside locked-down remote development environments (for example,
  Google Cloud Workstations) where outbound access to public CDNs is not guaranteed.

A CDN reference breaks both: an offline export shows nothing, and a network-restricted
workstation cannot load the renderer.

## Decision

We will vendor the ReDoc standalone bundle (~1 MB, MIT-licensed) into the repository at
a pinned version and embed it in the binary. Both serve and export deliver this embedded
copy; neither references an external CDN.

## Consequences

- The export is fully self-contained and renders offline.
- Serve works in network-restricted remote environments.
- A ~1 MB binary asset is checked into the repository, increasing repository and binary
  size.
- ReDoc upgrades require a manual vendor update; security and feature updates do not
  arrive automatically.
