# 0038 — Session attendance lives in its own opt-in tables, not on `plan`

- Status: accepted
- Date: 2026-11-16
- Deciders: Abolfazl Tayyar

## Context and Problem Statement

Athletes who train every day pay for a month, and the coach can tell when that ends from the plan's
start date. Athletes who train every other day (یک روز در میان) pay for 12 sessions, so when they
renew depends on how often they actually come. The coach wants to tap once per visit and see when the
12 are used. A new plan is the renewal. Many coaches won't use this at all, so it must cost them
nothing and leave the plan code alone.

## Decision Drivers

- Coaches who don't track sessions see no change, and no plan code depends on the feature.
- One visit per day and at most 12 per plan hold even under a double tap.
- No setting to turn the feature on.

## Considered Options

1. A `session_plan` row marks a tracked plan, and `attendance (plan_id, date)` holds its visits,
   in their own module (`internal/membership`) that reads plans through `plan.Service`
2. A `tracks_sessions` column on `plan` and the visits in a table the plan module owns
3. A per-coach setting that turns the feature on, with a session count on every plan

## Decision Outcome

Chosen option: **1**. No `session_plan` row means an untracked plan, so the feature is opt-in per plan
with no setting. The module imports `plan` for ownership and for the current plan; `plan` never
imports it, so the plan responses and code are unchanged. The client turns tracking on with a second
request after creating or copying a plan (`PUT /plans/:id/attendance`). `GET /attendance/current`
gives the home rows their counts in one request, the way `GET /plans/current` gives their plans
([0032](0032-current-plans-endpoint.md)), and "current" is the plan module's own definition
([0016](0016-current-plan-latest-started.md)).

A visit's primary key is `(plan_id, date)`, so a second visit on one day is refused by the schema.
Adding a visit locks the plan's `session_plan` row before counting, so two taps at once can't make a
13th. `attendance` references `session_plan`, so turning tracking off deletes the visits by cascade,
and deleting the plan deletes both. 12 is a constant (`SessionsPerPlan`), since every gym sells 12;
the API returns it as `sessions` so the client never hard-codes it.

The visit date comes from the phone (its today), and the server accepts up to its own UTC today plus
one day, like measurements, since the phone's date runs ahead of UTC.

Option 2 puts a column most coaches never use on the table every plan query reads, and ties the plan
create and copy code to the feature. Option 3 adds a setting and a per-plan count no gym needs.

### Consequences

- Good: removing the feature is dropping two tables and one module.
- Bad: a plan created and tracked in two requests can end up untracked if the second fails; the
  coach sees the toggle off and turns it on again, which is cheaper than coupling the plan module.
- Bad: tracking from plan create is two requests, not one.
- Bad: if a gym ever sells another number of sessions, the constant becomes a column on `session_plan`.
