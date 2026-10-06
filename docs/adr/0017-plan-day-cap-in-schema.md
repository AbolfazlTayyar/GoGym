# 0017 — A plan has at most 7 days, enforced by the schema

- Status: accepted
- Date: 2026-10-06
- Deciders: Abolfazl Tayyar

## Context and Problem Statement

A plan is one training week that repeats, so it should never hold more than 7 days. The spec
didn't cap it, and `day` had nothing stopping an 8th row or two days sharing an `order_index`.
There is no create-day endpoint yet, so the cap has to hold for whatever writes `day` later.

## Decision Drivers

- The cap must hold under concurrent inserts, not only for one request at a time.
- No per-insert locking or trigger to maintain.
- Day order should be unambiguous in the plan detail response.

## Considered Options

1. Count existing days in the service before inserting
2. A trigger that counts rows on insert
3. `UNIQUE (plan_id, order_index)` plus `CHECK (order_index BETWEEN 0 AND 6)`

## Decision Outcome

Chosen option: **3**. Seven slots, at most one day per slot, means at most seven days. Postgres
enforces it atomically. Option 1 lets two concurrent inserts both see six days and both succeed.
Option 2 has the same race unless it locks the plan row.

### Consequences

- Good: no race, no trigger. An over-cap insert fails as `gorm.ErrCheckConstraintViolated` and a
  taken slot as `gorm.ErrDuplicatedKey`, so the plan builder can map both to a 400.
- Good: days have a total order, so the detail endpoint never has to break a tie between days.
- Bad: `order_index` must be a dense 0–6 slot, not a sparse sort key like 10, 20, 30. Swapping
  two days in place needs a temporary slot, or a deferrable constraint added later.
- Neutral: the spec's day labels grew from A–F / day1–day6 to A–G / day1–day7 so every slot can be
  named.
