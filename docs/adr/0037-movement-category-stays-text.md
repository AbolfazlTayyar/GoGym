# 0037 — A movement category stays a normalised name, not a table

- Status: accepted
- Date: 2026-11-16
- Deciders: Abolfazl Tayyar

## Context and Problem Statement

`movement.category` is free text, and the client builds its category chips from the names its
movements use. The coach wants to rename a category once and have every movement follow, and never
to see two categories with the same name, including ones that only look the same: «قدرتی» typed
with a Persian ی and «قدرتى» typed on an Arabic keyboard, or with an extra space. She doesn't need a
category with no movements in it. Built-in movements carry categories too, and a coach can't change
them ([0021](0021-universal-rows-forbidden.md)).

## Decision Drivers

- KISS: one coach, a handful of categories.
- Rename and delete must reach every movement in the category, including soft-deleted ones that old
  plans still show.

## Considered Options

1. A `movement_category` table (`coach_id NULL` for built-in rows) and `movement.category_id`
2. Keep the text column; normalise every name on write, and rename or clear a category with one
   tenant-scoped `UPDATE` over the coach's movements

## Decision Outcome

Chosen option: **2**. A category is the name on its movements, so two with the same name are one
category by definition, and normalisation (Arabic ي ى ك to Persian ی ک, runs of spaces collapsed,
spaces and ZWNJs at the ends dropped) makes look-alikes the same name. The client applies the same
rules, so it can warn before a rename merges two categories. `PUT` and `DELETE
/api/v1/movement-categories/*name` rename or clear a category; a name a built-in movement uses is a
403, since changing only the coach's own movements would split it in two. Option 1 is only needed
for a category with no movements, which the coach doesn't want. It would cost a foreign key, a
backfill and a change to the movement API.

### Consequences

- Good: no schema change, and the movement API is unchanged.
- Bad: a category disappears once no movement uses it, and it has no id, so the routes address it
  by its URL-encoded name. That route is a catch-all because a name can contain "/".
- Moving to option 1 later is a backfill from names that are already normalised. Do it when the
  coach wants to set up categories before she has movements in them, or wants to order them.
