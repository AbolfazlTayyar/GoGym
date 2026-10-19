# 0032 — Each athlete's current plan comes from its own endpoint, not the athlete list

- Status: accepted
- Date: 2026-10-19
- Deciders: Abolfazl Tayyar

## Context and Problem Statement

Every row on the home screen shows its athlete's current plan: the title under the name and the
day letters as buttons ([0029](0029-blush-theme-day-buttons-home.md)). `GET /athletes` returns no
plan, so the client would need a plan list and a plan detail per athlete: 40–60 requests for one
screen. The athlete list can't simply add the plan itself: `internal/plan` imports
`internal/athlete` for ownership checks ([0015](0015-child-ownership-via-parent-lookup.md)), so
the reverse import is a cycle.

## Decision Drivers

- Home loads in a fixed number of requests and queries, whatever the number of athletes.
- "Current" keeps one definition ([0016](0016-current-plan-latest-started.md)).
- No new cross-module wiring for one screen.

## Considered Options

1. `GET /plans/current` in `internal/plan`, matched to the athlete list by `athlete_id`
2. A `current_plan` field embedded in `GET /athletes`, through an interface the plan module
   implements and `internal/server/server.go` wires into the athlete module
3. The client fetches each athlete's plans and current plan detail

## Decision Outcome

Chosen option: **1**. It returns the plan list's summary shape (title, `day_labels`,
`block_count`) for every one of the coach's athletes that has a current plan, private and public,
computed by the same `currentPlanID` the per-athlete list uses. It is unpaginated and unfiltered:
one coach's 20–30 athletes make one small response, so search results reuse it too. Home takes two
requests, `GET /athletes` and this one, and the endpoint runs four queries however many athletes
there are.

Option 2 changes the shape of the athlete list every client depends on, and needs an interface with
a single implementation plus wiring in `server.go` to dodge the import cycle. It would also make
every athlete-list query, including paginated search, load plans. Option 3 is the 40–60 requests.

### Consequences

- Good: the athlete list keeps its shape, and the plan module keeps the plan logic.
- Good: one response covers both tabs and search, since it is not filtered by athlete type.
- Bad: two requests instead of one, and the client joins them by `athlete_id`.
- Bad: all of the coach's plans are read to pick the current ones. That is a few hundred rows at
  this scale; a `DISTINCT ON` query would be the fix if it ever isn't.
