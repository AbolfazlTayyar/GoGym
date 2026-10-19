# 0033 — The builder saves a whole day by replacing its blocks

- Status: accepted
- Date: 2026-10-19
- Deciders: Abolfazl Tayyar

## Context and Problem Statement

The builder edits a whole day in the browser and saves it with one ذخیره button: sets and reps
change, lines are deleted, moved up and down, and turned into supersets. The API could only add
(`POST /days/:id/blocks`, `POST /blocks/:id/movements`), with no edit, delete or reorder.

## Decision Drivers

- One save is one request and one transaction: a failed save leaves the day as it was.
- A fixed number of queries per save, however long the day.
- The builder's draft maps onto the request without diffing.

## Considered Options

1. `PUT /days/:id/blocks` replaces the day's blocks with the full list
2. Separate update, delete and reorder endpoints for blocks and block movements (about ten), with the
   client sending one request per change
3. A JSON-patch style list of operations on the day

## Decision Outcome

Chosen option: **1**. The body is the day's blocks in order; positions become `order_index` and
`order_in_block`, so a client can't send gaps or duplicates. In one transaction the day row is
locked, its blocks are deleted (the cascade removes their movements), and the new ones are inserted
in one batch per table. The response is the saved day in the plan detail's shape, so the builder
needn't refetch. Errors are reported together, keyed by position (`[2].movements[1].movement_id`).

This is the delete-and-recreate that [er-diagram.md](../er-diagram.md) already assumes for `day`,
`block` and `block_movement`, which is why they carry no `updated_at`. Option 2 is many endpoints and
a save that can fail halfway through. Option 3 needs the client to diff its draft against the server
and the server to apply operations in order, for no gain at 6–9 lines a day.

A movement the day already uses is accepted even after it was deleted from the library, since the plan
keeps naming it ([0022](0022-delete-movement-in-use.md)); otherwise deleting a movement would make every
day that uses it unsaveable. Any other unusable movement is a field error
([0019](0019-unusable-body-reference-is-a-field-error.md)).

`POST /days/:id/blocks` and `POST /blocks/:id/movements` stay. The builder no longer uses them, but
they are the simplest way for a script or a test to append to a day, and removing them is a breaking
change that buys nothing. Remove them if their validation ever drifts from the save's.

### Consequences

- Good: one request, one transaction and a fixed number of queries per save.
- Good: the builder sends its draft as is.
- Bad: block and block movement ids change on every save, so nothing outside the day's tree may
  reference them (workout logs snapshot the prescription instead, [0018](0018-workout-logs-snapshot-prescription.md)).
- Bad: two saves of the same day at once: the later one wins whole and the earlier one's edits are
  lost. The day-row lock makes that the worst case; without it the day could hold both saves' lines.
