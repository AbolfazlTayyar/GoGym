# 0007 — swaggo docs, dev-only UI, root basePath

- Status: accepted
- Date: 2026-09-17
- Deciders: Abolfazl Tayyar

## Context and Problem Statement

With no frontend yet, endpoints need a browsable way to be exercised by hand. The API is versioned
under `/api/v1`, but `/healthz` is intentionally unversioned, and Swagger 2.0 only allows one
global `basePath`.

## Decision Drivers

- Docs generated from the handlers, so they stay close to the code.
- Don't expose interactive docs in prod.
- "Try it out" must hit the right URL for every route.

## Considered Options

1. Hand-written OpenAPI file
2. swaggo annotations with `@BasePath /api/v1`
3. swaggo annotations with `@BasePath /`, full path in every `@Router`

## Decision Outcome

Chosen option: **3**. The `/api/v1` basePath broke `/healthz` in Swagger UI. `/swagger/*any` is
mounted only in the dev environment. Generated output in `docs/swagger/` is committed because the
binary imports it. Annotations use typed structs, never `gin.H` or maps.

### Consequences

- Good: every endpoint is testable from the browser with real bearer auth.
- Bad: `make swagger` must be rerun after every annotated change, or stale docs ship.
