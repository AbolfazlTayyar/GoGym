# 0013 — Response envelope written by explicit helpers

- Status: accepted
- Date: 2026-09-23
- Deciders: Abolfazl Tayyar

## Context and Problem Statement

Each handler returned its own bare struct, so the response shape depended on the handler. It was
cheap to unify while only `internal/coach` returned JSON. It gets expensive once six modules and
a frontend depend on the current shapes.

## Decision Drivers

- One shape the frontend can parse: `success`, `data`, `error{code,message,fields}`, `meta`.
- Swagger annotations must describe the real body.
- No per-request buffering.

## Considered Options

1. Leave each handler with its own shape
2. A middleware that rewrites whatever the handler wrote into the envelope
3. Explicit `internal/httpx` helpers (`OK`, `Created`, `Error`, …) that handlers call

## Decision Outcome

Chosen option: **3**. Option 2 buffers every body, and it moves the real shape away from the
handler, so the swagger annotation stops matching what's returned. `NoRoute`/`NoMethod`/recovery
go through the same helpers. `/healthz` is exempt because probes match its bare
`{"status":"ok"}`. Swagger uses separate doc-only `SuccessEnvelope`/`ErrorEnvelope` types, checked
against the runtime type by a test.

### Consequences

- Good: one parseable shape, stable `error.code` constants, docs that match the real body.
- Bad: handlers must remember to use the helpers. A grep for `c.JSON` is the check.
