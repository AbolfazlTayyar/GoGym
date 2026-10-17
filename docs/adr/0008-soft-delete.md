# 0008 — Soft delete for athletes and movements, hard delete for measurements

- Status: accepted
- Date: 2026-09-19
- Amended: 2026-10-17, measurements made deletable, and hard-deleted
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

A coach can also delete a single measurement, to take a mistyped check-in off the progress chart.
That delete is hard. Soft delete protects an athlete because deleting one cascades to their whole
history, and a movement because plans reference it ([0022](0022-delete-movement-in-use.md)). Neither
applies to a measurement: nothing references it, and a wrong delete costs one check-in the coach
can enter again. With no undelete endpoint, a soft-deleted measurement could only come back by hand
in SQL, so a `deleted_at` column would be a migration for an undo nobody can use.

### Consequences

- Good: a soft-deleted athlete returns 404 everywhere without any extra code in the handlers.
- Bad: raw SQL and `Unscoped()` queries must remember the filter. Unique constraints still see deleted rows.
- Bad: a deleted measurement is gone, even by hand. A later undo or audit trail would need to add
  `deleted_at` first, and couldn't bring back measurements deleted before that.
