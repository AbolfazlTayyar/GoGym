# GoGym — Design concepts

The frontend design, built on [stakeholder-brief.md](../stakeholder-brief.md) and on two of her
real plans. Three boards cover every v1 screen: [spec.md](../spec.md)'s six features plus sign-in.

They are static mockups to set the direction, not specs. Athlete names are made up. The plan content
follows her real plans. The options this design was chosen from are in [archive/](archive/README.md).

## Chosen direction

Concept 8: concept 6's look, day view, new plan and builder, with concept 2's home (both in
[archive/](archive/README.md)). Why, and what was rejected, is in
[ADR 0029](../adr/0029-blush-theme-day-buttons-home.md).

| Board | Screens |
|---|---|
| [concept-8-1-home-and-plans.png](concept-8-1-home-and-plans.png) | Home, day view, new plan (copy first), builder |
| [concept-8-2-athlete.png](concept-8-2-athlete.png) | Profile with plans, profile with measurements, add measurement, new or edit athlete |
| [concept-8-3-search-movements-sign-in.png](concept-8-3-search-movements-sign-in.png) | Search, movement library, add or edit movement, sign in |

How they connect:

- **Home:** a day letter opens that day, the name opens the current plan, "+" adds an athlete, and
  typing in search shows private and public results together. Sign-out sits in a small sheet behind
  the greeting.
- **Day view:** the person icon opens the profile, the pencil opens the builder on this plan.
- **Profile:** a plan opens its day view, "+ برنامه جدید" starts the copy-first flow into the
  builder, "+ ثبت اندازه" opens the measurement sheet, and the pencil opens the athlete form.
- **Bottom bar:** athletes and movements, the only two top-level pages.

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
- Plan editing and duplicating are already listed in the brief. They are the *Plan update and
  delete*, *Save a day* and *New plan from a copy* tasks in [tasks/backend.md](../tasks/backend.md).

Already supported: the same movement twice in one block ("10 with weight + 10 without").
`block_movement` has no uniqueness on `(block_id, movement_id)`, and `load` tells the two apart.

Found while drawing the other screens. Each one is a task in
[tasks/backend.md](../tasks/backend.md), Phase 2:

- **Home needs each athlete's current plan.** `GET /athletes` returns no plan, so the plan title and
  day letters on every row would take a plan-list and a plan-detail request per athlete. *Current
  plans for the athlete list* adds one request that returns them all.
- **Athletes can't be edited or deleted.** There is no `PUT` or `DELETE /athletes/{id}`. The
  profile's pencil needs it, and so does moving an athlete between private and public. *Athlete
  update and delete* adds both.
- **Plan summaries have no size.** `GET /athletes/{id}/plans` has no day letters or counts, so
  until *Plan summaries: day letters and size* is done the profile shows them only for the
  current plan, whose detail it loads anyway.

Search across both groups (`athlete_type=all` with `q`), editing and deleting check-ins and
movements are already supported, and sign-out is part of the *Browser auth via httpOnly cookie*
task.

## Regenerating the images

The boards are HTML in [src/](src/) and [archive/src/](archive/src/). `src/data.js` holds the
shared mock data and icons, and `src/blush.css` the styles every concept 8 board shares. Re-render
one with:

```
google-chrome-stable --headless=new --hide-scrollbars --window-size=1860,1300 \
  --virtual-time-budget=8000 --screenshot=docs/design/concept-8-1-home-and-plans.png \
  "file://$PWD/docs/design/src/concept-8-1-home-and-plans.html"
```

On Windows, `msedge.exe` takes the same flags. Round 1 boards are 1440 wide (three phones), round 2
and 3 boards and concept 8 1860 (four phones).

The fonts (Vazirmatn, Inter, JetBrains Mono) load from Google Fonts, so rendering needs a network connection.
