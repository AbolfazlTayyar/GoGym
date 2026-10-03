# 0015 — Child tables prove ownership through the parent lookup

- Status: accepted
- Date: 2026-10-03
- Deciders: Abolfazl Tayyar

## Context and Problem Statement

`athlete_measurement` (and later `plan`, `day`, `block`) has no `coach_id`, so `tenant.Scope`
can't filter it directly. Trusting an `athlete_id` from the URL would let a coach write to
another coach's athlete.

## Decision Drivers

- Keep [0011](0011-tenant-isolation-single-helper.md)'s single enforcement point.
- Don't add `coach_id` to every child table.
- Soft-deleted, foreign, missing and malformed parents should all return the same 404.

## Considered Options

1. Add a denormalized `coach_id` to every child table
2. Join to the parent inside each child query
3. Load the parent through its tenant-scoped service (`athlete.Service.Get`) first, then touch the child

## Decision Outcome

Chosen option: **3**. Child repositories trust the parent id they're given and say so in a
comment. The service layer guarantees that parent id was already checked.

### Consequences

- Good: one ownership path, and the soft-delete and 404 behavior come with it for free.
- Bad: one extra query per request. A child repository called directly, without going through
  its service, has no protection.
