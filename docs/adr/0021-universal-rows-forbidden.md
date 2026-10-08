# 0021 — Changing a universal row is a 403; another coach's row stays a 404

- Status: accepted
- Date: 2026-10-08
- Deciders: Abolfazl Tayyar

## Context and Problem Statement

`movement` has universal rows (`coach_id IS NULL`) that every coach reads and none may change.
[0011](0011-tenant-isolation-single-helper.md) answers a row outside the coach's scope with 404. A
universal row is not outside it: it shows up in the coach's own list. `PUT` and `DELETE` on one need
an answer, and so does the same request against another coach's movement.

## Decision Drivers

- Nothing may reveal that another coach's row exists.
- The client should be able to tell "you can't change this" from "this is gone".

## Considered Options

1. 404 for both: anything the coach can't change doesn't exist for writing
2. 403 for both
3. 403 `forbidden` for a universal row, 404 for another coach's

## Decision Outcome

Chosen option: **3**. The coach already sees a universal movement, so a 403 tells them nothing new.
Option 1 would answer a row that's on screen with "not found", which reads like a stale list.
Option 2 confirms that another coach's id exists, which 0011 rules out. The check runs before body
validation, so a bad body doesn't change the answer. The write itself still goes through
`tenant.Scope`, so a universal row can't change even if the check is skipped. `httpx.CodeForbidden`
is new for this.

### Consequences

- Good: no leak, and the client can tell a read-only row apart from a missing one.
- Bad: the API now has both 403 and 404 for "can't change this". Each new table with shared rows has
  to make the same visible-or-not call.
