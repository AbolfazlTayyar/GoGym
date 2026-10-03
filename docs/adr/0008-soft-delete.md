# 0008 — Soft delete for coach-deletable tables

- Status: accepted
- Date: 2026-09-19
- Deciders: Abolfazl Tayyar

## Context and Problem Statement

Once other coaches use the app, deleting an athlete by accident wipes their whole history
(measurements, plans) through `ON DELETE CASCADE`, and there is no way to undo it. Adding this
later means a migration against live data.

## Decision Drivers

- Accidental deletes should be recoverable.
- Cheap now, with no production data yet.

## Considered Options

1. Hard delete only
2. `deleted_at` + GORM's `gorm.DeletedAt` default scope on athlete and movement
3. Archive tables / audit log

## Decision Outcome

Chosen option: **2**, on `athlete` and `movement`, the tables a coach can delete from the UI.
Child tables (measurement, plan, …) are not soft-deleted themselves. They become unreachable
because their parent lookup filters out the deleted parent (see [0015](0015-child-ownership-via-parent-lookup.md)).

### Consequences

- Good: a soft-deleted athlete returns 404 everywhere without any extra code in the handlers.
- Bad: raw SQL and `Unscoped()` queries must remember the filter. Unique constraints still see deleted rows.
