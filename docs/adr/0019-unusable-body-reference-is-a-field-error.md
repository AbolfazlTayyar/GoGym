# 0019 — An unusable id in the request body is a field error, not a 404

- Status: accepted
- Date: 2026-10-07
- Deciders: Abolfazl Tayyar

## Context and Problem Statement

`POST /blocks/:id/movements` takes an array of entries, each naming a `movement_id` from the
coach's library. When one names a movement the coach can't use (unknown, another coach's, or
soft-deleted), the whole batch is rejected. The question is how to report it: as a 404 like an
out-of-scope path id ([0011](0011-tenant-isolation-single-helper.md)), or as a validation error.

## Decision Drivers

- The client must be able to tell "the block is gone" from "one of the picks is bad".
- The client should know which entry to fix.
- Nothing may reveal that another coach's movement exists.

## Considered Options

1. 404 `not_found` for the whole request
2. 400 `validation_failed`, with the bad entry's field set to `"not found"`, e.g. `"[1].movement_id"`

## Decision Outcome

Chosen option: **2**. Under option 1, a missing block and a bad pick give the same code, and the
client can't tell which entry failed. Option 2 uses the same message for unknown, foreign and
soft-deleted movements, so nothing leaks. It is reported alongside every other field error in
the batch.

The resource a request acts on still returns 404, even when its id arrives in the body: `POST
/plans` with another coach's `athlete_id` is a 404, because the plan would be created under that
athlete. The rule applies to ids that only *reference* another row, such as a library entry.

### Consequences

- Good: the client can highlight the exact entry, and a stale parent still reads as a 404.
- Bad: "not found" now shows up under two codes, so each new body reference has to decide whether
  it is a parent (404) or a reference (400).
