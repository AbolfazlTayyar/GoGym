# GoGym — Archived design concepts

The options the coach chose from, in the three rounds they were drawn in. The chosen direction,
concept 8, is in [../README.md](../README.md#chosen-direction). Kept for the reasoning behind it;
don't build from these.

The boards' HTML is in [src/](src/). They load `board.css` and `data.js` from
[../src/](../src/), shared with concept 8.

## Round 3

Asked for after round 2: two looks taken from outside references. Both keep round 2's rules: no
per-athlete colors, tapping an athlete opens the plan, and a new plan starts from a copy.

| | 6 — Blush | 7 — Sport |
|---|---|---|
| Board | [concept-6-blush.png](concept-6-blush.png) | [concept-7-sport.png](concept-7-sport.png) |
| Reference | [appetit.fit](https://appetit.fit/): its palette (hot pink `#FF0057`, blush `#FFEFF2`–`#FFD3DB`, cream `#F7F3EF`), pill buttons with a pink glow, large radii | [openGym](https://github.com/DuarteSantos8/openGym): true black, gray iOS-style cards, one green (`#30D158`), big titles, week strip, set steppers, floating centre button |
| Idea | Concept 5's flow in a soft, light, pink look | A sports-tracker look made for a dark gym: big type, high contrast |
| Theme | Light, blush pink; yellow for "no plan", since red would get lost next to the pink | Dark, black and green; orange for "no plan", blue for cardio |
| Home | Pink header with a greeting, white athlete cards (initial, name, plan), ~7 on screen | Big title with today's date, two tiles (private, no plan), grouped list, ~6 on screen; a green centre button starts a new plan |
| App → the day she needs | 2 taps (card, day) | 2 taps (row, day) |
| Day view | One white sheet: pink numbers, sets × reps under each movement; the whole day fits one screenshot | The week strip with A–E and movement counts, then one grouped list; the whole day fits one screen |
| Builder | The day view made editable: a tapped line opens in blush with type, +/− and presets | Written movements as short rows, the open one as a big card with large steppers; values carry over from the previous movement |
| Adding a movement | Type at the bottom, tap a suggestion; the new line copies the previous line's values | A picker where one tap adds a result with the values chosen above the list ("مثل قبلی" by default) |
| Start from an earlier plan | A first step that asks where to start, with a preview of the copied days | Starts as a copy (card at the top of the builder) |
| Trade-off | Warm and friendly, and good-looking in a screenshot, but light and pink go against the spec's dark/neutral, high-contrast default for gym light; small pink text only passes contrast in the deeper shade (`#D6004A`) | Easiest to read in a dark gym, but the tiles and big title leave the fewest names on home, and it looks more like an athlete's tracker than a coach's notebook |

## Round 2

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

Parts can be mixed across rounds: e.g. concept 2's day buttons on concept 5's plain list, concept
2's number pad in any builder, or concept 7's one-tap picker in concept 6's look.
