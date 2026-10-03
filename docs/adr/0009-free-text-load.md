# 0009 — Prescribed load is free text

- Status: accepted
- Date: 2026-09-19
- Deciders: Abolfazl Tayyar

## Context and Problem Statement

`block_movement` had `reps` and `duration_seconds` but no load, and coaches program load. The
question was how to store it.

## Decision Drivers

- Coaches write load in mixed notations: kg, `%1RM`, RPE, "bodyweight".
- Don't lose information when the coach enters it.

## Considered Options

1. Numeric kg column
2. Structured value + unit columns
3. Free-text `load TEXT`

## Decision Outcome

Chosen option: **3**. A numeric column would force a lossy conversion at entry time, and a
structured model would be premature before the plan builder UX exists.

### Consequences

- Good: any notation a coach uses fits as written.
- Bad: no arithmetic or reporting on load (volume, progression). Revisit with a new ADR once
  reporting or workout logging needs a number.
