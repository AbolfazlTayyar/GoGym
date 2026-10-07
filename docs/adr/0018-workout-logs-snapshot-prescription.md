# 0018 — Workout logs snapshot the prescription, not the plan rows

- Status: accepted
- Date: 2026-10-07
- Deciders: Abolfazl Tayyar

## Context and Problem Statement

`day`, `block` and `block_movement` carry `created_at` only, because the plan builder deletes and
recreates them instead of editing in place. Workout logging (what the athlete actually did) doesn't
exist yet, but when it does, each log row has to say what it was a log *of*. Training history is
the hardest data in the product to replace, so the choice has to be made before the first log table.

## Decision Drivers

- No plan edit may orphan or silently rewrite an athlete's training history.
- A log must show what was prescribed on the day it was done, not what the plan says now.
- Plan rows should stay cheap to rebuild.

## Considered Options

1. A log foreign-keys to `block_movement_id`
2. Keep plan rows stable (edit in place, soft delete) so a log can foreign-key to them safely
3. A log references `movement_id` plus a denormalized snapshot of the prescription (sets, reps,
   duration, load as programmed)

## Decision Outcome

Chosen option: **3**. Option 1 orphans every log on the next plan edit. Option 2 keeps the foreign
key valid, but an edited row then reports the new prescription for a session done under the old
one, so it needs a snapshot anyway, and it gives up delete-and-recreate for nothing. The rule is
also written next to the models in `internal/plan/model.go`.

### Consequences

- Good: logs survive any plan edit, and history reads exactly as it was programmed.
- Good: the plan builder keeps deleting and recreating rows freely.
- Bad: the prescription is stored twice, and a log can't point back at the exact block it came from.
- Neutral: plan templates and duplication would copy a plan and edit the copy in place, which works
  against delete-and-recreate. If templates are coming, `updated_at` on `day`, `block` and
  `block_movement` is cheap to add before there is data to backfill. That is not decided here.
