# 0011 — Tenant isolation through one scoping helper

- Status: accepted
- Date: 2026-09-20
- Deciders: Abolfazl Tayyar

## Context and Problem Statement

Every coach-owned table carries `coach_id`. If every repository has to remember its own `WHERE
coach_id = ?`, a single forgotten one leaks one coach's athletes to another.

## Decision Drivers

- Make a cross-tenant query hard to write and easy to spot in review.
- Don't reveal that another coach's row exists.

## Considered Options

1. A hand-written `Where("coach_id = ?")` in each repository
2. A single helper, `tenant.Scope(db, coachID)`, that every tenant-owned query is built from
3. Postgres row-level security

## Decision Outcome

Chosen option: **2**. `internal/tenant` is the only place the scope is applied. A row outside the
coach's scope returns **404, not 403**. RLS was **deferred**, not rejected: while this helper is
the only path into tenant-owned tables, RLS would enforce the same rule a second time. Revisit it
once a background job or admin tool can query those tables outside this path.

### Consequences

- Good: an unscoped query stands out in review, and the scoping rule lives in one file.
- Bad: only reviewers enforce it, not the database, until RLS is added.
