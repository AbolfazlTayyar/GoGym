# 0002 — Modular monolith with layered modules

- Status: accepted
- Date: 2026-09-14
- Deciders: Abolfazl Tayyar

## Context and Problem Statement

The app has domains that must not get tangled: athlete/plan management now, and a coach
accounting module later. It is built by one person, ships as one product, and is partly a
vehicle for learning idiomatic Go backend structure.

## Decision Drivers

- Keep separate concerns (e.g. billing vs. plans) from leaking into shared service/repository layers.
- One deployable, so infra stays simple.
- Practice real layering instead of one big `main.go`.

## Considered Options

1. Layer-first layout (`handlers/`, `services/`, `repositories/`)
2. Modular monolith: module-first folders, handler → service → repository → model inside each
3. Microservices

## Decision Outcome

Chosen option: **2**. Each feature is `internal/<module>/{model,repository,service,handler,routes}.go`,
wired only in `internal/server/server.go`. `internal/coach` is the reference shape. See
[architecture.md](../architecture.md).

### Consequences

- Good: a module can be read, tested and changed on its own. Billing can arrive later without touching plans.
- Bad: shared models were parked in `internal/models` until their owning module existed, so they
  have to be moved over one module at a time.
