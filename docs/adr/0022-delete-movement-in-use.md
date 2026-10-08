# 0022 — Deleting a movement a plan uses is allowed, as a soft delete

- Status: accepted
- Date: 2026-10-08
- Deciders: Abolfazl Tayyar

## Context and Problem Statement

`DELETE /movements/:id` can name a movement that `block_movement` rows in existing plans still
reference. The spec doesn't say what should happen. Silently breaking an existing plan is not acceptable.

## Decision Drivers

- An existing plan must read exactly as it did before the delete.
- A coach should be able to clean up their library. It is a picker, and clutter makes it slow.

## Considered Options

1. Refuse with 409 `conflict` while any `block_movement` references the movement
2. Soft delete regardless: plans keep the reference, and the library and new picks lose it
3. Hard delete, cascading to or nulling `block_movement.movement_id`

## Decision Outcome

Chosen option: **2**. [0008](0008-soft-delete.md) made `movement` soft-deletable so that references
stay valid, and plan detail already loads movements `Unscoped`. A deleted movement keeps its name in
every plan that used it. Option 1 protects plans, but they are already safe, and a movement used once
in an old plan could then never leave the library. Option 3 changes plans the coach didn't touch.

### Consequences

- Good: a delete never changes a plan, and the library only holds what the coach wants to pick from.
- Bad: a plan can show a movement that is no longer in the library, and there is no undelete yet.
- Bad: [0019](0019-unusable-body-reference-is-a-field-error.md) rejects a soft-deleted movement in a
  new block. Once the builder edits blocks by deleting and recreating them, re-saving a block that
  holds a deleted movement would fail. That flow has to allow movements the block already had.
