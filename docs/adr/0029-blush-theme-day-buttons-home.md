# 0029 — Light blush theme, with day buttons on the home screen

- Status: accepted
- Date: 2026-10-19
- Deciders: Abolfazl Tayyar

## Context and Problem Statement

The frontend needs one design direction before its screens are built.
[design/archive/](../design/archive/README.md) holds seven concepts drawn from the [stakeholder brief](../stakeholder-brief.md) and two of the
coach's real plans. [spec.md](../spec.md) asked for a dark/neutral, high-contrast theme for gym
lighting. The brief left the theme open.

## Decision Drivers

- Home → the day an athlete names is the most-used flow, done one-handed in a crowded gym.
- One day of a plan has to read well in a single screenshot sent to a public athlete.
- The builder has to be fast enough to write 20–30 plans on a phone.

## Considered Options

1. Concept 2 as drawn: dark and lime, day buttons on home, a table day view and a table builder
   with its own number pad
2. Concept 6 as drawn: light blush, home → day in 2 taps, the day as one white sheet, copy-first
   builder
3. Concept 7: dark sports-tracker look, the closest to the spec's dark/high-contrast line
4. Concept 6 with concept 2's home (drawn as concept 8)

## Decision Outcome

Chosen option: **4**. Concept 2's home is the only one where any day is 1 tap from opening the app,
and its dense rows fit more athletes than concept 6's cards. The rest is concept 6's: the day view
reads like her paper plans and a whole day fits one screenshot, which concept 2's table doesn't
match in polish. The theme goes against the spec's dark/neutral line, which now points here.

### Consequences

- Good: home → any day is 1 tap, and about nine athletes fit on home.
- Good: the day view, new plan and builder are drawn and unchanged from concept 6.
- Bad: light and pink haven't been tried under gym lighting, which is what the spec's dark line
  was protecting. Small pink text has to use `#D6004A`: `#FF0057` on white is under 4.5:1. "No plan"
  is yellow, since red gets lost next to the pink.
- Bad: one theme only. No dark variant is designed.
- Bad: the day letters take the width the avatar used, so a row is told apart by its name alone.
