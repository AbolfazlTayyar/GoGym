# 0034 — Prescribed load is a weight in kg

- Status: accepted
- Date: 2026-11-10
- Deciders: Abolfazl Tayyar
- Supersedes: [0009](0009-free-text-load.md)

## Context and Problem Statement

[0009](0009-free-text-load.md) stored `block_movement.load` as free text, before the plan builder
existed, so coaches could write kg, `%1RM`, RPE or "bodyweight". The builder now lets the coach set a
weight on each movement, and her plans only ever prescribe it in kg. The question was whether the
weight stays text or becomes a number.

## Decision Drivers

- The coach sets the weight with +/− and by typing a number; nothing else is offered.
- Workout logging and progress (volume, heaviest weight over time) will need arithmetic on it.
- The app is not published yet, so the column can change with no data to convert.

## Considered Options

1. Keep `load TEXT` and have the builder write and parse a fixed shape such as `"17.5 کیلو"`
2. Add a numeric `load_kg` beside `load TEXT`, which stays for RPE, `%1RM` or "bodyweight"
3. Replace `load` with a numeric `load_kg`

## Decision Outcome

Chosen option: **3**. `load_kg NUMERIC(5,2)`, nullable (no weight), 0.01–500 kg, checked by the
service like reps. Option 1 keeps a number inside text that every reader must parse, and any other
writer of the API can break the shape. Option 2 keeps a field no screen writes, and two places to read
one prescription from. The app is unpublished, so the migration drops `load` without converting it.

### Consequences

- Good: the weight can be summed, compared and charted, and the API validates it.
- Good: a superset that repeats a movement ("۱۰ with weight + ۱۰ without") still tells the two apart.
- Bad: a load that isn't kg (`%1RM`, RPE, "bodyweight") has no field. Add one with its own ADR if
  the coach starts prescribing that way.
- Bad: the API changed shape (`load` → `load_kg`), so a client written against `load` breaks.
