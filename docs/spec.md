# GoGym — MVP One-Page Spec v1

Audience: a personal trainer (coach) managing their own athletes. Not athlete-facing.

## Users

- **Coach** — the only login/account in v1. Manages athletes, builds plans, tracks measurements.
- **Athlete** — a record the coach manages, not a login. No athlete-facing app in v1.

> Single coach for v1, but every athlete/plan-owning table carries `coach_id` from day 1 so multi-coach later is a permissions feature, not a schema migration.

## Core entities (as defined)

- `Coach` — id, first_name, last_name, phone, password_hash *(added for auth; not in the original data model but needed to log in)*
- `Athlete` — id, coach_id, first_name, last_name, phone, experience_level (beginner/intermediate/advanced), injuries, goal, height
- `AthleteMeasurement` — id, athlete_id, date, weight, chest, waist, arm, thigh, hip — one row per check-in, drives the progress chart
- `Plan` — id, athlete_id, start_date, title/note — an athlete can have several plans over time
- `Day` — id, plan_id, label (A/B/C/D/E/F or day1/day2/day3/day4/day5/day6), order_index  — a plan is made of several days
- `Block` — id, day_id, order_index, sets, rest_seconds, notes — a day is made of several blocks
- `BlockMovement` — id, block_id, movement_id, reps, duration_seconds, order_in_block — a block can hold one or several movements (supersets/combo movements grouped under one block)
- `Movement` — id, name, category (warmup/strength/cardio/…), coach_id (nullable), description — a coach-editable library, not per-athlete. `coach_id = NULL` marks a universal, system-seeded movement, visible to every coach but not editable or deletable by them; a non-null `coach_id` is a coach's own custom addition.

## MVP feature list (in scope for v1)

1. Coach dashboard — athlete list + "add athlete" action + search by athlete first/last name
2. Athlete profile — base info + measurement trend chart + "add measurement" action
3. Athlete's plan list — current plan highlighted, past plans as history
4. Plan detail — tabbed by Day (A/B/C/D…), each day lists its blocks; movements within a block are grouped visually (supersets read as one unit, not separate rows)
5. Plan builder — add day → add block to a day → add movement(s) to a block, with search over the movement library
6. Movement library management — list, search, add/edit/delete for the coach's own movements; system-seeded (universal) movements are visible but read-only

This is the v1 loop: **add athlete → build plan (days → blocks → movements) → log measurements over time.**

## Explicitly out of scope for v1~

- Messaging / notifications
- Dashboard quick status per athlete
- Multi-coach accounts, admin roles, permissions UI
- Athlete-facing login/app
- Scheduling / calendar / session check-in flow
- Analytics beyond the measurement trend chart

## Planned post-v1: coach accounting module

Not in the v1 feature list, but confirmed as a real part of the system: a `billing`/accounting module where the coach calculates monthly income and does related accounting actions, scoped to their own athletes/sessions. Treated as its own module (own entities, service, repository) from the architecture side — see roadmap step 2 — even though it ships after the core v1 loop. Entities and exact scope TBD; revisit once v1's athlete/plan/measurement loop is stable.

## Design constraints carried from the UI brief

- Mobile-first, dark/neutral professional theme, RTL (Persian content)
- High contrast, large touch targets — legible under gym lighting
- Athlete and plan cards must stay dense-but-uncluttered on a touch UI

## Next step

Per the roadmap: turn the entity list above into an ER diagram, then write the first migration — `Coach → Athlete → AthleteMeasurement` and `Athlete → Plan → Day → Block → BlockMovement → Movement` as the two main chains.
