# 0012 — No organization layer above coach

- Status: accepted
- Date: 2026-09-20
- Deciders: Abolfazl Tayyar

## Context and Problem Statement

`product-direction.md` §7 considered selling to gyms, which would need an `organization` above
coach (shared library, athlete transfer, owner reporting). It suggested a cheap hedge so that
tenancy could later switch to `org_id`.

## Decision Drivers

- The target customer is the freelance coach, not the gym.
- Avoid designing for a market that hasn't been validated.

## Considered Options

1. Add an `organization` table now
2. Don't add the table, but design around a future `org_id`
3. Not planned: the app stays per-coach

## Decision Outcome

Chosen option: **3**. This supersedes the hedge in `product-direction.md` §7. `tenant.Scope` is
kept for leak prevention ([0011](0011-tenant-isolation-single-helper.md)). That it would also make
an `org_id` switch a one-file change is a side benefit, not a reason to build anything.

### Consequences

- Good: simpler schema and permissions, and no speculative design work.
- Bad: if gyms turn out to be the market, the org layer becomes a real migration (a new ADR).
