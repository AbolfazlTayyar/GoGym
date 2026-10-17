# GoGym — Design concepts

Layout options for the coach to choose from, built on [stakeholder-brief.md](../stakeholder-brief.md)
and on two of her real plans. Each board shows the screens the brief's two main needs run through:
**home** (find the athlete), **day view** (read the next movement) and **builder** (write a plan
fast).

These are static mockups to pick a direction, not specs. Athlete names are made up. The plan content
follows her real plans.

## Round 2 — current options

Asked for after round 1: one design close to the earlier `gym-trainer-app.jsx`, and one simpler
design. Neither uses per-athlete colors.

| | 4 — Classic | 5 — Simple |
|---|---|---|
| Board | [concept-4-classic.png](concept-4-classic.png) | [concept-5-simple.png](concept-5-simple.png) |
| Idea | The earlier design kept almost as it is, changed only where the brief needs it | As little on screen as possible; the plan reads like her paper plans |
| Theme | Dark, the original charcoal + amber | Light, black on white, no other colors (red only for "no plan" and delete) |
| Home | The original athlete cards (avatar, goal · level, status pill), ~5 on screen; search field and private/public chips on top | Plain list: name and current plan, ~10 on screen |
| App → the day she needs | 2 taps (card, day) | 2 taps (name, day) |
| Day view | The original block cards and day pills; a nine-movement day scrolls, so a screenshot shows about five | Numbered lines with sets × reps underneath; the whole day fits one screen and one screenshot |
| Builder | The original editor, one day at a time behind A–E pills; a small sets stepper per block; the original movement sheet, where the picked movement opens to set type, reps, sets and presets | The day view made editable: tap a line to open it (type, +/− for reps and sets, superset, move); type a name at the bottom and tap a suggestion |
| Start from an earlier plan | Amber "copied from" card at the top of the builder | A first step that asks where to start, her last plan picked by default |
| Trade-off | Familiar and polished, but the most scrolling, both on home and in the day view | Fastest to read and to screenshot, but plain: nothing on home says which athletes need attention |

What concept 4 drops from the original, and why:

- **Week-progress plates ("هفته ۳ از ۸").** A plan has a start date but no length, so there is
  nothing to count to. Her plans are monthly. If she wants this back, the plan needs a length or end
  date, which is a schema change.
- **"Last update" and the "needs follow-up" status.** Nothing records when an athlete was last
  updated. "No active plan" stays, because the API knows that.
- **Profile as the first stop.** Tapping an athlete opens the plan. The profile moved to the person
  icon in the plan header.

## Round 1

The first three options. Kept for reference. Concept 1 is built on per-athlete colors, which round 2
drops.

| | 1 — Notebook tabs | 2 — Day buttons on home | 3 — Refined cards |
|---|---|---|---|
| Board | [concept-1-notebook.png](concept-1-notebook.png) | [concept-2-day-buttons.png](concept-2-day-buttons.png) | [concept-3-cards.png](concept-3-cards.png) |
| Idea | Her notebook made digital: each private athlete is a colored stick | Each athlete row carries its day letters | The earlier `gym-trainer-app.jsx`, reworked for the brief |
| Theme | Light, paper-like | Dark, high contrast (lime) | Dark (charcoal + amber) |
| Home | 3-column grid of colored sticks, 15 on screen | Dense rows with A–E buttons, ~10 on screen | 2-column cards, ~10 on screen |
| App → the day she needs | 2 taps (stick, day) | **1 tap** (the day on the row) | 2 taps (card, day) |
| Switching athlete mid-session | **Colored rail on the day view**, 1 tap | Back to home, 1 tap on the day | Back to home, 2 taps |
| Day view | Numbered list like her handwriting; the whole day fits one screenshot | Table with large reps/sets columns | One card per movement; a nine-movement day just fits |
| Builder | Inline rows; the open row has steppers, per-leg/per-side and presets; search at the bottom | Table with its own number pad and quick values; copy a single day | Bottom sheet per movement: search, type, steppers, presets |
| Start from an earlier plan | Copy of the whole plan | Copy day by day, from any plan | Copy of the whole plan |
| Trade-off | Light theme may glare less but stands out less in a dark gym; colors only help if she keeps them consistent | Most efficient but most "tool-like"; the table looks less polished in a screenshot sent to a public athlete | Familiar and polished, but cards take space and it's the slowest to scan |

Parts can be mixed across both rounds: e.g. concept 2's day buttons on concept 5's plain list, or
concept 2's number pad in any builder.

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

The boards are HTML in [src/](src/) (`data.js` holds the shared mock data). Re-render one with:

```
google-chrome-stable --headless=new --hide-scrollbars --window-size=1860,1300 \
  --virtual-time-budget=8000 --screenshot=docs/design/concept-4-classic.png \
  "file://$PWD/docs/design/src/concept-4-classic.html"
```

Round 1 boards are 1440 wide (three phones), round 2 boards 1860 (four phones).

The fonts (Vazirmatn, Inter, JetBrains Mono) load from Google Fonts, so rendering needs a network connection.
