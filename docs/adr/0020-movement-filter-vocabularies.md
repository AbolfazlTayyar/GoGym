# 0020 — Movement muscle group and equipment are closed vocabularies

- Status: accepted
- Date: 2026-10-08
- Deciders: Abolfazl Tayyar

## Context and Problem Statement

`movement` only had `category`, and category alone means scrolling while building a plan. The
movement list needs muscle group and equipment filters. These columns were cut from the schema
amendments task, so this change adds them. The question is what a value in them can be.

## Decision Drivers

- A filter must not miss a row because a coach typed "Legs", "leg" or a Persian word instead of `legs`.
- The UI is Persian-first, so the client should be able to label values in its own language.
- The seeded library and coaches' own movements must use the same values.

## Considered Options

1. Free text, matched case-insensitively
2. One value per column from a closed set, enforced by a `CHECK` constraint and by the service
3. A closed set, but `muscle_group TEXT[]` so a compound lift can list several groups
4. Lookup tables with foreign keys

## Decision Outcome

Chosen option: **2**. `muscle_group` is one of `chest, back, shoulders, arms, legs, core, full_body`.
`equipment` is one of `bodyweight, barbell, dumbbell, kettlebell, machine, cable, band, other`. Both
are nullable. Option 1 makes every filter a guess at spelling. Option 3 is more accurate (a deadlift
trains back and legs), but it makes the filter an array match and the picker UI harder, before any
coach has asked for it. Option 4 only pays off once coaches can extend the list, which v1 doesn't do.
The coarse groups match how a coach scans ("leg day"), and `other` keeps unusual equipment from
being rejected. The `CHECK` follows `athlete_type`: filtering depends on the value always being one
of the set.

### Consequences

- Good: filters are exact matches, the client can show fixed chips, and the seed covers every value.
- Bad: a compound lift is filed under one group, and a new value needs a migration plus a Go constant.
  Revisit with option 3 if coaches ask for secondary muscle groups.
