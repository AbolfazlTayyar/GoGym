# GoGym — Design concepts

The design direction for the frontend, built on [stakeholder-brief.md](../stakeholder-brief.md)
and on two of her real plans. The board shows the screens the brief's two main needs run through:
**home** (find the athlete), **day view** (read the next movement) and **builder** (write a plan
fast).

It is a static mockup to set the direction, not a spec. Athlete names are made up. The plan content
follows her real plans. The options it was chosen from are in [archive/](archive/README.md).

## Chosen direction

[concept-8-blush-day-buttons.png](concept-8-blush-day-buttons.png) is the design to build from:
concept 6's look, day view, new plan and builder, with concept 2's home (both in
[archive/](archive/README.md)). Why, and what was rejected, is in
[ADR 0029](../adr/0029-blush-theme-day-buttons-home.md).

What the home takes from concept 2, adapted to the blush look:

- **Every athlete row carries its plan's day letters.** Tapping one opens that day, so home → any
  day is 1 tap. Tapping the name opens the plan.
- **The rows share one white sheet** instead of concept 6's separate cards, so about nine athletes
  fit instead of seven. The avatar, the chevron and concept 2's per-athlete color bar are gone: the
  letters need the width.
- **Letters keep their columns.** A four-day plan leaves E's place empty. A row with no plan gets the
  yellow "بدون برنامه" pill and a "برنامه جدید" button where the letters go.

## What her plans showed

Two consecutive monthly plans for the same athlete, made in Canva:

- **Days are letters, A–E**, 4–5 days a plan, 6–9 numbered movements a day.
- **Each day opens and closes with timed cardio** (bike or treadmill, 10–15 minutes). These sit
  outside the numbering.
- **Sets × reps is one number for all sets** (`۱۵ تکرار × ۳ ست`), with no ranges or per-set pyramids.
  Timed work is in seconds or minutes (`۴۵ ثانیه × ۳`, `پلانک دو دقیقه × ۲`).
- **Per leg / per side is common** (`۱۵ تکرار هرپا`, `۲۰ تکرار هر سمت`).
- **Supersets and tri-sets are written with "+"**, and each part has its own reps or time
  (`۱۵ تکرار هرپا + ۴۵ ثانیه × ۳ ست`, `۱۲+۱۲+۱۲`).
- **The same movement can appear twice in a combo** ("۱۰ with weight + ۱۰ without").
- **No rest times and no loads.** The only rest instruction is a day-level note ("Day E with
  minimal rest"). Pauses are written into the movement name ("with a 2-second pause").
- **Next month's plan reuses this month's structure** with different movements, which is the
  "rewriting similar plans" pain from the brief.

Every builder follows these: day letters, presets based on her common values (12/15/20 × 3,
15 × 4, 30 seconds × 3), rest hidden, and a per-leg / per-side / seconds toggle.

## Gaps against the current API

Found while mapping her plans onto `block` / `block_movement`. Most of her plans fit: reps as a
whole number, sets per block, supersets as several movements in a block, duration in seconds. These
don't:

- **Per leg / per side has no field.** `block_movement` has `reps`, `duration_seconds` and `load`.
  "15 each leg" needs a marker of its own. Writing it into `load` would misuse that field.
- **A day has no note.** "Do day E with minimal rest" belongs to the day, but `day` only has a
  `label`.
- **Warm-up and cool-down cardio** fit as a one-movement block with `duration_seconds`. Showing
  them outside the numbering, as every design does, needs a rule: either their position (first
  or last block) or the movement's category.
- Plan editing and duplicating are already listed in the brief.

Already supported: the same movement twice in one block ("10 with weight + 10 without").
`block_movement` has no uniqueness on `(block_id, movement_id)`, and `load` tells the two apart.

## Regenerating the images

The boards are HTML in [src/](src/) and [archive/src/](archive/src/) (`src/data.js` holds the
shared mock data). Re-render one with:

```
google-chrome-stable --headless=new --hide-scrollbars --window-size=1860,1300 \
  --virtual-time-budget=8000 --screenshot=docs/design/concept-8-blush-day-buttons.png \
  "file://$PWD/docs/design/src/concept-8-blush-day-buttons.html"
```

Round 1 boards are 1440 wide (three phones), round 2 and 3 boards and concept 8 1860 (four phones).

The fonts (Vazirmatn, Inter, JetBrains Mono) load from Google Fonts, so rendering needs a network connection.
