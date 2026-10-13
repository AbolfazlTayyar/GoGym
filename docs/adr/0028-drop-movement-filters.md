# 0028 — Movement has no muscle group or equipment until the library outgrows search

- Status: accepted
- Date: 2026-10-13
- Deciders: Abolfazl Tayyar
- Supersedes: [0020](0020-movement-filter-vocabularies.md)

## Context and Problem Statement

[0020](0020-movement-filter-vocabularies.md) added `muscle_group` and `equipment` to `movement` as
closed-set filter columns. The app serves one coach, and their library is the 20 seeded movements
plus whatever they add. Name search and `category` already find a movement in a list that short.
The two filters cost two `CHECK` constraints, Go constants that must match them, validation,
swagger enums, a seed that has to cover every value, and filter chips in a client that doesn't
exist yet. Nothing has shipped, so no production database or client depends on them.

## Decision Drivers

- KISS: build for one coach's library, not a large one.
- Adding the columns back later must stay cheap.

## Considered Options

1. Keep them, as 0020 decided
2. Keep the columns but stop exposing them in the API
3. Drop them with a new migration on top of the ones that added them
4. Remove them from the migration history: delete the migration that added them and take them out of the seed

## Decision Outcome

Chosen option: **4**. The column migration is gone, the seed is renumbered to `000005` and inserts
no muscle group or equipment, and the API neither accepts nor returns them. Option 2 leaves dead
columns that someone has to keep consistent. Option 3 is what a deployed schema would need, but
with nothing shipped it only adds an add-then-drop pair that every fresh database runs for nothing.
Renumbering the seed, rather than leaving a gap at `000005`, makes a local database that ran the old
files fail on `migrate up` instead of silently keeping the stale columns.

Adding them back is cheap: the response gains fields, which is not a breaking change, and the
columns are nullable, so only the seeded rows need a backfill. The *Movement muscle group &
equipment filters* task in `docs/tasks/backend.md` holds that work.

### Consequences

- Good: less code and schema to keep in sync while the library is small.
- Bad: a local dev database migrated before this change has to be recreated.
- Bad: this only works because nothing has shipped. After the first production deploy, a schema
  change is always a new migration.
