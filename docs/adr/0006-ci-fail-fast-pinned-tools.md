# 0006 — Fail-fast CI with pinned tool versions

- Status: accepted
- Date: 2026-09-17
- Deciders: Abolfazl Tayyar

## Context and Problem Statement

CI should catch broken code on `main` quickly. The first pipeline used golangci-lint `latest`,
which resolved to a stale v1 binary that couldn't read the v2-schema `.golangci.yml`, so CI failed
for a reason that had nothing to do with the code.

## Decision Drivers

- A lint error shouldn't wait for a Postgres container to start.
- CI failures should come from the code, not from tool versions changing between runs.

## Considered Options

1. One job, floating tool versions
2. Two jobs (`lint-and-unit` → `integration`), every tool/action version pinned

## Decision Outcome

Chosen option: **2**. GitHub Actions runs `go vet`, golangci-lint (pinned, e.g. `v2.13.2`) and unit
tests first. The testcontainers suite runs only if that job passes. Version bumps are deliberate,
in their own commit.

### Consequences

- Good: fast feedback, and the same run gives the same result next week.
- Bad: tools don't upgrade themselves; someone has to bump them.
