# 0014 — Offset pagination and private-by-default athlete list

- Status: accepted
- Date: 2026-10-02
- Deciders: Abolfazl Tayyar

## Context and Problem Statement

`GET /athletes` is the coach's dashboard. A coach's list can grow over time. `public` athletes
are one-off plan-link deliveries the coach doesn't manage ongoing, so mixing them in would bury
the coach's actual working list.

## Decision Drivers

- Bounded response size.
- The default view should be the athletes the coach actively works with.
- Simple for a frontend that hasn't been chosen yet.

## Considered Options

1. Flat, unpaginated list
2. `limit`/`offset` with `total` in `meta`
3. Cursor pagination

## Decision Outcome

Chosen option: **2**, with default limit 20 and max 100. The applied window is echoed in `meta`.
The list defaults to `athlete_type = private`, and `public` athletes are reachable through an
explicit filter rather than hidden. Results are newest first. Cursor pagination is overkill at
this size.

### Consequences

- Good: page numbers and totals are easy to build in any UI.
- Bad: offset paging can skip or repeat rows if athletes change between pages, which is acceptable
  at one coach's scale. `listMeta` moves to `httpx` once a second endpoint paginates.
