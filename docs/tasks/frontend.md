# GoGym — Frontend tasks (v1)

The React app in `web/`, from scaffolding to every v1 screen and serving it from Caddy. Why this stack is in [ADR 0030](../adr/0030-react-spa-same-origin.md). The screens are the three concept 8 boards in [design/](../design/README.md), and the task format is explained in [README.md](README.md).

Several screens need a backend task from Phase 2 of [backend.md](backend.md) first; each task's status line names it.

> **The boards are the spec for how the app looks.** The app looks like [concept-8-1](../design/concept-8-1-home-and-plans.png), [concept-8-2](../design/concept-8-2-athlete.png) and [concept-8-3](../design/concept-8-3-search-movements-sign-in.png). Each board is HTML in [design/src/](../design/src/): its `concept-8-*.html`, the shared `blush.css`, and the SVG icons in `data.js`. Screens are built by porting that markup, CSS and those icons, not by drawing them again from the PNG. Every screen's checkpoint is a side-by-side check at 390×840, the boards' screen size: same layout, colors, fonts, spacing and icons. The only differences allowed are listed below.

**Where the app differs from the boards, on purpose:**

| Board | What's drawn | What the app does | Why |
|---|---|---|---|
| 8-1 day view, builder | Warm-up and cool-down as cream strips outside the numbering | Every block is numbered | The rule for spotting them is undecided; see the deferred *Warm-up and cool-down strips* task |
| 8-1 builder, day view | The هرپا / هرسمت options, "۱۵ تکرار هرپا" | Reps or seconds only | The API has no per-side field yet; see the deferred *Builder: per-side reps…* task |
| 8-1 new plan | Start date ۱ آبان, the first of next month | Start date is today and the title names this month | Asked for |
| 8-2 profile | Older plans show only their start date | Older plans also show their day count and size | *Plan summaries* added that data; the board predates it |
| Not drawn | Deleting an athlete, deleting a plan (a swipe on its profile row) or a day, editing a plan's title and date (the same swipe), the sign-out sheet, empty and error states | Built from the boards' existing parts (sheet, pink pill, white sheet, yellow strip) | The backend supports them and the coach needs them |
| Not drawn | Swipe actions | Plan, check-in, movement and athlete rows swipe right to left, as in iOS Messages, to show حذف and ویرایش; built-in movements don't swipe | Asked for. A plan is edited and deleted only this way; the other lists' tap and sheet still do both |

> **Mobile first.** The coach does everything on a phone, building whole plans included ([stakeholder-brief.md](../stakeholder-brief.md) §2). Every screen is designed and checked at 360–430px wide first, on a real phone before it's called done. Wider screens get the same phone-width column, centered, not a desktop layout.

> **One API module.** Every call goes through `web/src/api/`, which unwraps the response envelope and throws `ApiError`. Screens never call `fetch` directly. This is also where offline caching would plug in later (ADR 0030).

> **Persian everywhere.** UI text is written in Persian in the components, with no i18n library. Dates show as Jalali with Persian digits, and every number a coach types goes through `toLatinDigits` before it's sent.

---

## Phase 0 — Foundation

### Scaffold the web app

✅ **Done**

**Description:** An empty React app in `web/` that builds, lints and tests, and talks to the local API through Vite's proxy. In development the browser then sees one origin, the same as behind Caddy in production, so the auth cookie works without CORS.

**Prompt:**
```
Create the frontend app in web/ at the repo root (docs/adr/0030).

- Vite + React + TypeScript (strict), npm, with package-lock.json committed. Pin the Node
  version in web/.nvmrc and in package.json "engines"; CI reads the same file.
- React Router v7 in library (SPA) mode and TanStack Query, nothing else yet. Each later
  task adds the libraries it needs and says why.
- ESLint (the Vite template's config with the react-hooks rules), Prettier, and Vitest.
  npm scripts: dev, build, preview, lint, typecheck, test.
- index.html: lang="fa" dir="rtl", the viewport meta (width=device-width,
  initial-scale=1, viewport-fit=cover), a theme-color, and the title GoGym.
- Fonts: the boards' two fonts, self-hosted: Vazirmatn (@fontsource-variable/vazirmatn)
  for the text, and Inter (@fontsource-variable/inter) where docs/design/src/blush.css
  uses it (day letters, counts, the GoGym wordmark). Not Google Fonts, which is
  unreliable from Iran and would also block offline use later.
- vite.config.ts proxies /api and /healthz to http://localhost:8080. `npm run dev --
  --host` serves the app on the LAN, for testing on a real phone.
- Makefile targets web-dev, web-build and web-test that run in web/; the Go targets stay
  as they are. Ignore web/node_modules and web/dist in .gitignore.
- A placeholder page that calls GET /healthz through the proxy and shows the result. The
  sign-in task replaces it.

Add a "Frontend conventions" section to CLAUDE.md with what's settled so far (the app
lives in web/, mobile first, the dev proxy) and a line for web/ in docs/architecture.md's
package layout. Later frontend tasks add to that section instead of starting new ones.
```

**Checkpoint:** With the API running (`docker compose up`), `make web-dev` serves the placeholder and it shows `ok`. The same page opens on a phone on the same Wi-Fi at `http://<pc-ip>:5173`. `npm run lint`, `npm run typecheck`, `npm test` and `npm run build` all pass in `web/`.

**Expected output:**
```
$ make web-build
vite v8.3.4 building client environment for production...
✓ built in 221ms
dist/index.html
dist/assets/index-….css
dist/assets/index-….js
```

**In short:** there's an empty app on the phone that can reach the API.

---

### Design tokens, base styles and the app frame

✅ **Done**

**Description:** The concept 8 look in code: the colors, shadows and background from `docs/design/src/blush.css`, RTL-safe base styles, and the frame every screen sits in, a header plus the bottom bar with the app's two top-level pages.

**Prompt:**
```
Port the concept 8 look into the app and build the frame every screen sits in.

- src/styles/tokens.css: the custom properties from docs/design/src/blush.css (--pink,
  --pinkText, --blush, --line, --glow, --soft and the rest) on :root, plus the page's
  background gradient. Keep the board's names so the two stay comparable. Small pink text
  uses --pinkText (#D6004A); #FF0057 on white is under 4.5:1 (docs/adr/0029).
- src/styles/base.css: a reset, Vazirmatn, a line height that suits Persian, and logical
  properties only (margin-inline-start, padding-inline, inset-inline-end), so nothing
  hard-codes left or right. Icons that point (back chevrons) are mirrored for RTL.
- Mobile first. The base styles are the phone layout, 360-430px wide, with no media
  queries. Wider screens keep a phone-width column (max-width around 480px), centered.
  Touch targets are at least 44px. Inputs use at least a 16px font, or iOS Safari zooms
  the page on focus. Use 100dvh, not 100vh, and pad fixed bars with
  env(safe-area-inset-bottom) so they clear the home indicator once installed.
- The frame: a header (title, subtitle, round buttons), a scrolling content area, and the
  bottom bar with شاگردها and حرکات (docs/design/README.md, "How they connect"). Both
  routes exist and show an empty state for now.
- Plain CSS: the global token and base files, plus a CSS module per component. No
  Tailwind and no component library; the design is custom and they would fight it.
- src/components/icons.tsx: the SVG icons from docs/design/src/data.js (ICON), as
  components, unchanged. Screens use these, not an icon library, so they match the boards.
- The boards are the spec for how the app looks (frontend.md, top). Port values from
  blush.css and the board HTML (sizes, radii, shadows, font weights), don't eyeball the PNG.

Add the rules above to CLAUDE.md's Frontend conventions: tokens keep blush.css's names,
icons come from data.js, screens are ported from the board HTML and checked side by side
at 390x840, logical properties only, and the mobile-first rules.
```

**Checkpoint:** In devtools at 390×840, the header, background and bottom bar match the boards side by side. At 360px and on a real phone, switching between the two pages works. No horizontal scroll at 360px. At 1280px the app is a centered column. `grep -rnE "(margin|padding)-(left|right)|[^-](left|right):" web/src --include=*.css` finds nothing.

**In short:** the app looks like the boards and has the frame every screen will fill.

---

### Frontend CI job

✅ **Done**

**Description:** CI runs lint, the type check, tests and the build for `web/` on every push, so a broken frontend fails the run the same way broken Go does.

**Prompt:**
```
Add a web job to .github/workflows/ci.yml, on the same triggers as the Go jobs:
checkout, actions/setup-node with node-version-file: web/.nvmrc and npm caching on
web/package-lock.json, then npm ci, lint, typecheck, test and build, all in web/.

- It runs in parallel with the Go jobs; it doesn't wait for the integration tests.
- Pin every action version, per CLAUDE.md's CI conventions.
- Only skip it when web/ didn't change if that stays a simple paths filter; say which you
  chose.
```

**Checkpoint:** Push a branch with a deliberate type error in `web/`: the web job fails at typecheck while the Go jobs pass. Fix it and the whole run is green.

**In short:** a broken frontend can't sit on `main` unnoticed.

---

### API client

✅ **Done**

**Description:** The one module every API call goes through. It unwraps the `{success, data, error, meta}` envelope, turns failures into a typed error with Persian messages, and sends the coach to sign-in on a 401.

**Prompt:**
```
Add web/src/api/, the only place the app calls the API from.

- client.ts: a fetch wrapper for /api/v1 paths. It sends and reads JSON, uses
  credentials: "same-origin" for the auth cookie, and unwraps the envelope described in
  CLAUDE.md's API response conventions. On success it returns data, and meta when the
  caller asks for it. On failure it throws an ApiError with status, code, message and
  fields. A body that isn't an envelope (a proxy's 502, a dropped connection) becomes an
  ApiError with a code of its own, never an uncaught JSON parse error.
- codes.ts: the error codes the client branches on, mirroring the httpx.Code* constants,
  each with the Persian message to show. Never show the server's error.message; it's
  English.
- types.ts: hand-written types for the payloads, matching the json tags in
  internal/*/handler.go (docs/adr/0030 says why they aren't generated). Nullable fields
  are `T | null`, not optional.
- One file per resource (athletes.ts, plans.ts, ...) of plain async functions, each added
  by the screen task that first needs it.
- A 401 from any call other than login clears the query cache and sends the app to
  sign-in, remembering the page to come back to.
- One QueryClient: no refetch on every window focus, retries for network errors only (a
  4xx won't change on retry). Query keys come from one function per resource, so a save
  can invalidate the right lists.
- A helper that maps ApiError.fields onto form fields by json name, for the forms later
  tasks build.

Vitest tests with a mocked fetch: success returns data and meta; validation_failed throws
with fields; a 401 sends the app to sign-in; a non-JSON 502 becomes an ApiError. Add to
CLAUDE.md's Frontend conventions: calls go through src/api, failures are ApiError, and
messages come from codes.ts.
```

**Checkpoint:** `npm test` passes. On the placeholder page, a deliberate call to an unknown `/api/v1` path shows the Persian not-found message, not English text or a blank screen.

**In short:** every screen talks to the API the same way and shows Persian errors.

---

### Persian digits and Jalali dates

✅ **Done**

**Description:** A Persian phone keyboard types ۱۲, not 12, and the API rejects it. The coach reads dates as ۱ مهر ۱۴۰۵, while the API speaks `2026-09-23`. These helpers are used by every form and every date on screen ([stakeholder-brief.md](../stakeholder-brief.md) §5).

**Prompt:**
```
Add web/src/lib/digits.ts and web/src/lib/jalali.ts, and a Jalali date field.

- digits.ts: toLatinDigits converts Persian (۰-۹) and Arabic-Indic (٠-٩) digits to Latin,
  and the Persian decimal separator "٫" to "."; every number and phone a coach types
  passes through it before sending. toPersianDigits for display. parseNumber returns a
  number, or null for an empty or invalid field.
- jalali.ts: show an ISO date as Jalali with Persian digits through
  Intl.DateTimeFormat("fa-IR-u-ca-persian"), with no library. Convert between Jalali
  year/month/day and ISO with jalaali-js for the date field. Also: today as an ISO date
  in the phone's local time, not UTC (a check-in logged at 01:00 in Tehran must not land
  on yesterday), and the Jalali month names.
- A Jalali date field: shows the date and opens a picker sheet for year, month and day.
  Keep it simple. Only two forms use it, and both pre-fill the date.

Unit tests: Persian, Arabic and mixed input; "۶۸٫۵" → 68.5; 1405/01/01 = 2026-03-21 and
1403/12/30 = 2025-03-20 (1403 is a leap year); "today" just after midnight in
Asia/Tehran.
```

**Checkpoint:** `npm test` passes. On a phone with a Persian keyboard, typing ۶۸٫۵ into a test field shows `68.5` as the parsed value.

**In short:** the app reads Persian input and shows Jalali dates.

---

### Sign in and sign out

✅ **Done**

**Description:** The one screen before home: mobile and password ([concept-8-3](../design/concept-8-3-search-movements-sign-in.png), fourth phone). There's no signup, since coach accounts are made with `coachctl`. The coach stays signed in, and signing out is the one item in a small sheet behind the home greeting.

**Prompt:**
```
Build sign-in and sign-out, and the bottom sheet later screens reuse. Sign-in matches
the fourth phone of docs/design/concept-8-3-search-movements-sign-in.png: port its markup
and styles from docs/design/src/concept-8-3-search-movements-sign-in.html and blush.css.
The sheet copies the measurement sheet on concept-8-2's third phone.

- /login: the logo, mobile and password with the eye toggle, and the pink ورود button.
  The mobile field is type="tel" inputmode="numeric" and goes through toLatinDigits.
  POST /api/v1/auth/login sets the auth cookie; the app never reads or stores a token.
  invalid_credentials and rate_limited show their Persian message under the form, and
  field errors under their field.
- Startup: GET /api/v1/coaches/me decides between the app and /login. Show only the logo until it answers, so a signed-in coach never sees /login
  flash. A signed-in coach who opens /login goes home.
- After login, go back to the page the 401 redirect remembered, or home.
- Sign-out: a sheet behind the greeting on home with one item. POST /api/v1/auth/logout,
  clear the query cache, go to /login.
- No signup and no "forgot password": accounts and resets go through coachctl
  (docs/adr/0025).
- The bottom sheet component: a backdrop, slides up, closes on a backdrop tap, a swipe
  down and the phone's back button (that needs a history entry while it's open). A native
  <dialog> may be enough; if you pick a library, check it's still maintained and say why.
```

**Checkpoint:** Side by side with concept-8-3's fourth phone at 390×840, sign-in matches it. On a phone: a wrong password shows the Persian invalid-credentials message, and a mobile typed in Persian digits signs in. Close the browser and reopen it: you're still signed in and land on home without seeing `/login`. Sign out from the greeting sheet, and then opening a deep link like `/movements` goes to `/login` and back to `/movements` after signing in. The Android back button closes the sheet instead of leaving the page.

**In short:** the coach signs in once and stays signed in, and the app never touches the token.

---

## Phase 1 — Screens

Ordered so the plan builder can be tried with a real plan on a phone as early as possible ([stakeholder-brief.md](../stakeholder-brief.md) §5: builder speed decides the first day). Home, the athlete form and the day view come first because the builder needs an athlete and a plan to open.

### Home: athletes with their day letters

✅ **Done**

**Description:** The screen the app opens on, and the most-used one: every private athlete with their current plan's day letters, so the day an athlete names is one tap away ([concept-8-1](../design/concept-8-1-home-and-plans.png), first phone). Typing in search finds private and public athletes together ([concept-8-3](../design/concept-8-3-search-movements-sign-in.png), first phone).

**Prompt:**
```
Build home to match the first phone of docs/design/concept-8-1-home-and-plans.png, and
search to match the first phone of concept-8-3-search-movements-sign-in.png. Port the
markup and styles from those boards' HTML in docs/design/src/ and blush.css. Home takes
two requests: GET /api/v1/athletes?limit=100 (the default is 20 and
the coach has 20-30 athletes; 100 is the max, docs/adr/0014) and GET /api/v1/plans/current,
matched by athlete_id. Never one request per athlete.

- Header: the greeting سلام مربی as drawn (opens the sign-out sheet), شاگردهای من, and the pink "+" that
  opens the new athlete form with the open tab's type.
- The خصوصی / عمومی switch with counts from meta.total; public is athlete_type=public.
  Private is the default.
- Rows share one white sheet: the name, the plan title under it, and the day letters as
  buttons. A letter opens the day view on that day. The name opens the current plan on its
  first day, or the profile when there's no plan. Letters keep their columns by
  order_index, so a four-day plan leaves the fifth place empty. If any plan has 6 or 7
  days, check the row still fits at 360px and tell me what you did.
- No current plan: the yellow بدون برنامه pill, and a "+ برنامه جدید" button where the
  letters go, opening the new-plan screen for that athlete.
- Search: typing queries athlete_type=all&q= (debounced around 250ms) and shows private
  and public results as two groups with the match highlighted. The API does the matching;
  don't filter on the client. Clearing the search brings the switch back.
- Loading, empty (no athletes yet) and error states.

Mobile: the list scrolls between a fixed header and the bottom bar, and every letter has a
44px tap area. Unit test the function that places letters in columns. A component test: a
row with no plan shows the pill, and a row with a four-day plan leaves the fifth place
empty.
```

**Checkpoint:** Side by side at 390×840, home matches concept-8-1's first phone and search matches concept-8-3's first. On a phone with real data, the athlete list and every row's day letters arrive in two requests (devtools network tab). Tapping a letter opens that day, and back returns to home at the same scroll position. Searching for a fragment of a public athlete's name finds them under عمومی while the switch is hidden.

**In short:** opening the app and tapping a letter shows the day an athlete is on.

---

### Athlete form: new and edit

✅ **Done**

**Description:** One form for a new athlete and for editing one ([concept-8-2](../design/concept-8-2-athlete.png), fourth phone). Name, surname and mobile are required. Everything else is optional, and type and level are buttons, so the only typing is names, phone, goal, height and injuries.

**Prompt:**
```
Build the athlete form, one component for new and edit, to match the fourth phone of
docs/design/concept-8-2-athlete.png. Port the markup and styles from
docs/design/src/concept-8-2-athlete.html and blush.css.

- /athletes/new?type=private|public and /athletes/:id/edit. First name, last name and
  mobile are required. Type is two buttons, level three (the API's beginner, intermediate,
  advanced, with Persian labels). Goal, height in centimeters and injuries are optional,
  marked اختیاری. The ذخیره pill in the header saves; × asks before discarding changes.
- New: POST /api/v1/athletes, then the new athlete's profile. Edit: PUT /api/v1/athletes/:id
  with every field, since an omitted optional field is cleared. Edit also has a delete
  button with a confirmation: DELETE /api/v1/athletes/:id, then home.
- The mobile field is type="tel" inputmode="numeric" and height is inputmode="decimal",
  both through toLatinDigits. API field errors show under their field.
- After a save or delete, invalidate the athlete lists and that athlete's queries.

Component tests: a 400 with fields.phone shows the message under mobile; Persian digits in
mobile and height are sent as Latin.
```

**Checkpoint:** Side by side with concept-8-2's fourth phone at 390×840, the form matches it. On a phone, add an athlete using only the Persian keyboard, and they appear on home under the right tab. Edit them to public and they move to the عمومی tab. Delete them and they're gone from both.

**In short:** the coach can add, fix and remove athletes from the phone.

---

### Day view

✅ **Done**

**Description:** One day of a plan as one white sheet: numbered lines, the prescription under each in the coach's own wording, supersets as one line with a pink edge ([concept-8-1](../design/concept-8-1-home-and-plans.png), second phone). The coach reads it mid-session, and public athletes get it as a screenshot, so one day must fit one screen.

**Prompt:**
```
Build the day view to match the second phone of docs/design/concept-8-1-home-and-plans.png.
Port the markup and styles from docs/design/src/concept-8-1-home-and-plans.html and
blush.css.

- /plans/:planId?day=<label>, opening on the first day when day is missing or unknown.
  GET /api/v1/plans/:id, and GET /api/v1/athletes/:athleteId for the name in the header.
- Header: the athlete's name, "برنامه <title> · شروع <Jalali start date>", the person icon
  (profile), the pencil (builder on this plan and day), and back.
- Day letters as pills, the open one pink. Switching days replaces the URL instead of
  pushing it, so back leaves the plan in one step.
- The sheet: "روز B · ۹ حرکت", then each block as a numbered line: the pink number, the
  movement name, and the prescription under it in Persian ("۱۵ تکرار × ۳ ست",
  "۴۵ ثانیه × ۳ ست", whole minutes as دقیقه), with load after it when set. A block with
  several movements is one numbered line with each movement on its own row joined by "+",
  the سوپرست tag and the pink edge (spec.md: a superset reads as one unit).
- Warm-up and cool-down strips: not yet. Their rule is undecided (docs/design/README.md,
  "Gaps"), so every block is numbered for now.
- A 404 (a deleted plan or athlete) shows a not-found state with a way home.

Unit test the prescription text: reps, seconds, minutes, load, and a three-movement block.
```

**Checkpoint:** Enter the board's day B (the same movements and prescriptions) and compare it with concept-8-1's second phone side by side at 390×840: it matches apart from the warm-up and cool-down strips. A nine-line day, the coach's longest, fits a 390×844 screen without scrolling, and a screenshot of it reads clearly with the superset visibly grouped. If it doesn't fit, say what would have to give. Switching days and pressing back once returns to home.

**In short:** any day of any plan reads like the coach's notebook and fits one screenshot.

---

### New plan: copy first

✅ **Done**

**Description:** Most plans are close to one the coach already wrote, so a new plan starts by choosing where it starts from: this athlete's last plan (suggested), another athlete's plan, or nothing ([concept-8-1](../design/concept-8-1-home-and-plans.png), third phone).

**Prompt:**
```
Build the new-plan screen to match the third phone of
docs/design/concept-8-1-home-and-plans.png. Port the markup and styles from
docs/design/src/concept-8-1-home-and-plans.html and blush.css.

- /athletes/:id/plans/new, opened from the profile's "+ برنامه جدید" and home's no-plan
  button. Three choices:
  - a copy of this athlete's latest plan (the first in GET /athletes/:id/plans), chosen
    by default and marked پیشنهادی, showing its day letters and the line count per day
    from GET /plans/:id;
  - a copy of another athlete's plan: pick the athlete with the same search as home,
    then one of their plans;
  - an empty plan.
- Title pre-filled "برنامه <this Jalali month name>", start date today, both editable.
- ادامه: POST /api/v1/plans, with source_plan_id for a copy. For an empty plan, also POST
  /plans/:id/days with label A and order_index 0, so the builder has a day to open. Then
  open the builder on the first day, replacing this screen in history so back doesn't
  return to it.
- Invalidate the athlete's plan list and home's current plans.
```

**Checkpoint:** Side by side with concept-8-1's third phone at 390×840, the screen matches it. Copy an athlete's current plan, and the builder opens with the same days and lines. Changing a line there leaves the source plan unchanged in its day view. Copy a plan from a different athlete. Start an empty plan and the builder opens on an empty day A. In each case, back from the builder goes to where the coach came from, not to this screen.

**In short:** next month's plan starts as this month's, a tap away.

---

### Builder: edit a day

✅ **Done**

**Description:** The day view's sheet, editable, saved in one request ([concept-8-1](../design/concept-8-1-home-and-plans.png), fourth phone). This screen decides the coach's first day with the app: 20–30 notebook plans typed in back to back, each in under 15 minutes, on a phone ([stakeholder-brief.md](../stakeholder-brief.md) §3–5).

**Prompt:**
```
Build the builder to match the fourth phone of docs/design/concept-8-1-home-and-plans.png.
Port the markup and styles from docs/design/src/concept-8-1-home-and-plans.html and
blush.css, and the presets from PRESETS in data.js. The day being edited is a draft held in the browser: a reducer over
plain block and movement objects, with a pure function for every edit (add a line, set
reps or seconds, set sets, make or break a superset, move up or down, delete). ذخیره sends
the whole draft with PUT /api/v1/days/:id/blocks, then replaces it with the response.

- /plans/:planId/edit?day=<label>. Header: "برنامه <title>", the athlete's name, "کپی از
  <source_plan_title>" when GET /plans/:id returns one, the ذخیره pill, and back. A plan's
  title, start date and deletion are on the profile's plan rows, not here.
- Day letters with "+", which adds the next free label (A-G) at the next order_index
  through POST /plans/:id/days. Deleting a day (DELETE /days/:id) sits behind a long press
  or a small menu, with a confirmation.
- Lines look like the day view. Tapping one opens it in place: type تکرار or ثانیه (reps or
  duration_seconds), +/- steppers for that and for sets, preset chips from the coach's
  plans (۱۲ × ۳, ۱۵ × ۳, ۲۰ × ۳, ۱۵ × ۴, ۳۰ ثانیه × ۳), "+ سوپرست" (the next movement joins
  this block), بالا / پایین, and delete. Rest is hidden (docs/design/README.md). The board's
  هرپا / هرسمت options wait for per-side reps (the deferred Plan prescription detail task
  in backend.md); leave them out.
- The search box pinned at the bottom: typing searches GET /movements?q=, and picking a
  result adds a line that copies the previous line's sets and reps. When nothing matches,
  the last suggestion is "+ افزودن «<text>»": POST /movements with that name, then add it.
- Unsaved changes: switching day, going back or closing the tab with a changed draft asks
  first. A failed save keeps the draft and shows the field errors on their lines; the API
  keys them "[i].movements[j].<field>".
- Mobile: the pinned search box stays above the on-screen keyboard on iOS Safari and
  Android Chrome (visualViewport if needed), steppers have 44px targets, and nothing in a
  line needs typing; only the search does.

Unit tests for the reducer: every edit, making and breaking a superset, moving a line past
a superset, a new line copying the previous one, and turning the draft into the request
body.
```

**Checkpoint:** Side by side with concept-8-1's fourth phone at 390×840, with a line open, the builder matches it apart from the هرپا / هرسمت options. On a phone, type in one real day from the coach's notebook (a nine-line day with a superset) and time it. Report the time. The goal is about 20–30 seconds a line, so under 5 minutes for the day. Save, and the day view shows exactly what was typed. Break the connection and save: the draft is still there with an error, and saving again once back online works. Leave with unsaved changes and the app asks first.

**In short:** a whole day is written with taps, not typing, and saved in one go.

---

### Profile: plans

✅ **Done**

**Description:** Everything about one athlete except the plan itself: who they are, what to check before writing a plan (level, height, goal, injuries), and every plan they've had ([concept-8-2](../design/concept-8-2-athlete.png), first phone).

**Prompt:**
```
Build the athlete profile and its plans tab to match the first phone of
docs/design/concept-8-2-athlete.png. Port the markup and styles from
docs/design/src/concept-8-2-athlete.html and blush.css.

- /athletes/:id. Header: the name, the phone as a tel: link, the pencil (athlete form) and
  back. A card with type, level, height and goal chips, and the injuries in the yellow
  strip when set.
- Tabs برنامه‌ها <count> and اندازه‌ها <count>, with the open tab in the URL (?tab=) so back
  returns to it. The measurements tab is the next task; show an empty state for now.
- Plans from GET /athletes/:id/plans: a plan that hasn't started is marked بعدی; the
  current plan is outlined in pink, with its day letters as buttons into the day view and
  each day's line count (from GET /plans/:id, the same query the day view uses); older
  plans sit under برنامه‌های قبلی with their start date, day count and block_count. A plan
  opens its day view.
- Swiping a plan right to left, as in iOS Messages, shows حذف and ویرایش behind its right
  edge. ویرایش opens a sheet to change the title and start date (PUT /plans/:id); حذف asks
  first, then DELETE /plans/:id.
- "+ برنامه جدید" pinned at the bottom opens the new-plan screen.
```

**Checkpoint:** Side by side with concept-8-2's first phone at 390×840, the profile matches it apart from the older plans' added sizes. For an athlete with an upcoming, a current and two older plans, each appears in its place with the right tag. A letter on the current plan opens that day. The phone number dials when tapped on a phone. Back from a day view returns to the plans tab.

**In short:** one screen shows who an athlete is and every plan they've had.

---

### Profile: measurements and the trend chart

✅ **Done**

**Description:** One measurement at a time as a trend line, the history newest first, and a sheet to log or fix a check-in, which shows each field's last value while the coach types ([concept-8-2](../design/concept-8-2-athlete.png), second and third phones).

**Prompt:**
```
Build the measurements tab, its chart and the measurement sheet to match the second and
third phones of docs/design/concept-8-2-athlete.png. Port the markup and styles from
docs/design/src/concept-8-2-athlete.html and blush.css, the chart's SVG included.

- GET /api/v1/athletes/:id/measurements. Chips for وزن, سینه, کمر, بازو, ران, باسن
  (weight, chest, waist, arm, thigh, hip), weight first. For the chosen one: the latest
  value with its unit (کیلو or سانت), the change since the first check-in that has it,
  and a line chart over the check-ins that have it. A check-in without that value is
  skipped, not drawn as 0.
- The chart is a hand-written SVG (docs/adr/0030): the line, the points, the soft fill and
  Jalali date labels. Time runs right to left, like the text and the board; the design
  README says to check this with the coach, so keep the direction in one constant.
- History newest first, each row showing the date and the values it has. Tapping a row
  opens the sheet to edit it (PUT /measurements/:id), with a delete button (DELETE
  /measurements/:id).
- "+ ثبت اندازه" opens the sheet empty: the Jalali date (today) and six fields, each
  showing the athlete's last value under it, with the decimal keyboard and Persian digits
  converted. Only filled fields are sent. The API needs at least one, and its field errors
  show under their field.
- Invalidate the measurement list after any change.

Unit tests for the chart's scaling (one point, all values equal, a falling line) and for
finding each field's last value.
```

**Checkpoint:** Side by side with concept-8-2's second and third phones at 390×840, the tab and the sheet match them. Log a check-in with only weight and waist, typed in Persian digits with "٫". Both charts get a new point, and the other four don't. Edit and then delete a check-in, and the chart follows. An athlete with one check-in shows a single point, not an error.

**In short:** an athlete's progress is a glance, and logging a check-in takes seconds.

---

### Movement library

✅ **Done**

**Description:** The coach's movements and the built-in ones in one alphabetical list, with search and category chips. Built-in movements are read-only ([concept-8-3](../design/concept-8-3-search-movements-sign-in.png), second and third phones). Most movements are created from the builder's search, so this screen is mainly for tidying up.

**Prompt:**
```
Build the movement library and the movement sheet to match the second and third phones
of docs/design/concept-8-3-search-movements-sign-in.png. Port the markup and styles from
docs/design/src/concept-8-3-search-movements-sign-in.html and blush.css.

- /movements (bottom bar). Header: حرکات and "<n> حرکت شما · <m> پیش‌فرض". GET /movements
  returns the coach's own and the universal ones together, ordered by name. A search box
  (q, debounced) and category chips: همه plus the categories in the list, computed on
  the client.
- Universal movements (coach_id null) carry the lock tag پیش‌فرض and open the sheet
  read-only, with no save or delete.
- The coach's own open the sheet to edit: name, category chips (those in use plus
  "+ دسته جدید" for a new one), description, ذخیره (PUT /movements/:id) and حذف (DELETE,
  with a confirmation saying plans that use it keep showing it, docs/adr/0022). "+" in the
  header opens the same sheet empty (POST /movements).
- The UI never offers to edit a universal movement. If a 403 comes back anyway, show its
  message and refresh the list.
- After any change, invalidate the movement list. The builder's search uses the same
  query, so it sees the change too.
```

**Checkpoint:** Side by side with concept-8-3's second and third phones at 390×840, the list and the sheet match them. A built-in movement opens with no save or delete button. Add a movement with a new category, and its chip appears. Rename it, and the builder's search finds the new name. Delete one that a plan uses, and that plan's day view still shows it.

**In short:** the coach can tidy their exercise list without touching the built-in one.

---

## Phase 2 — Ship

### Serve the app from Caddy

⬜ **Not started** — needs *VPS deployment & production migrations* in [deployment.md](deployment.md).

**Description:** The built app is static files. Caddy already sits in front of the API, so it serves them too, on the same origin (ADR 0030): `/api/*` goes to Go, everything else is the app.

**Prompt:**
```
Ship the frontend behind the Caddy from the VPS deployment task (docs/adr/0030).

- A multi-stage image: a node stage runs npm ci and npm run build in web/, and the final
  stage is the caddy image with web/dist copied in. Pin both image versions.
- Caddyfile: /api/* and /healthz reverse_proxy to the api as today; everything else is a
  file_server over the built files with try_files {path} /index.html, so reloading
  /plans/<id> works. Files under /assets have hashed names and get a year-long immutable
  Cache-Control. index.html and the manifest get no-cache, so a deploy shows on the next
  open.
- The fallback must not catch the API: an unknown /api/v1 path is still the API's
  enveloped 404, never index.html. /swagger stays dev-only (CLAUDE.md).
- Runbook: deploy the API, with its migrations, before the frontend; a new screen may call
  an endpoint the old API doesn't have.
- If the deployment task added a CD job, extend it to build and ship this image too.
```

**Checkpoint:** `https://<your-domain>/` shows sign-in on a phone, and after signing in, reloading a deep link like `/plans/<id>?day=B` opens that day. The auth cookie is `Secure`. The commands below show the right headers.

**Expected output:**
```
$ curl -s https://<your-domain>/api/v1/nope
{"success":false,"data":null,"error":{"code":"not_found","message":"not found"}}

$ curl -sI https://<your-domain>/assets/index-<hash>.js | grep -i cache-control
cache-control: public, max-age=31536000, immutable

$ curl -sI https://<your-domain>/plans/123 | grep -iE "^(content-type|cache-control)"
content-type: text/html; charset=utf-8
cache-control: no-cache
```

**In short:** the app is live on the same address as the API.

---

### Home-screen install

⬜ **Not started** — after *Serve the app from Caddy*: installing needs HTTPS.

**Description:** The coach wants an icon on the home screen that opens like an app, with no app store ([stakeholder-brief.md](../stakeholder-brief.md) §2). That needs a web app manifest and icons. A service worker is only needed if the coach's browser won't offer install without one.

**Prompt:**
```
Make the app installable on the home screen.

- web/public/manifest.webmanifest: name GoGym, a short_name, lang fa, dir rtl, start_url /,
  display standalone, theme and background colors from the tokens, and icons at 192 and
  512 plus a maskable one, made from the logo on the sign-in board.
- An apple-touch-icon and the apple-mobile-web-app meta tags for iOS, whose Add to Home
  Screen ignores most of the manifest.
- No service worker unless the coach's phone's browser won't offer install without one.
  Check on the actual phone and tell me which it was. Offline caching is the deferred
  Offline reading task, not this one.
- Installed, the app has no browser back button. Check every screen has its own way back,
  and that the safe-area padding from the frame task clears the home indicator.
```

**Checkpoint:** On the coach's phone, the app installs from the browser and its icon opens it full screen, signed in, on home. Every screen can be left without a browser back button. The bottom bar isn't covered by the home indicator.

**In short:** GoGym is an icon on the coach's home screen.

---

## Deferred

Not scheduled. Each task says what should trigger it.

### Offline reading

⏸ **Deferred** — start when a plan fails to load in the gym because of the connection.

**Description:** The brief says the connection is reliable ([stakeholder-brief.md](../stakeholder-brief.md) §2), so v1 is online only. If that turns out false in the gym, the useful part is reading: the athletes and plans already loaded should open without a connection. ADR 0030 kept this cheap: the app's files can be cached by a service worker, and the API data by TanStack Query's persister, both plugged in at `src/api/` and the app entry.

**Prompt:**
```
Let the app open and show already-loaded athletes and plans without a connection.

- vite-plugin-pwa (Workbox) precaches the built app, so it starts offline. Check it
  doesn't serve a stale index.html after a deploy: an update applies on the next open.
- TanStack Query's IndexedDB persister keeps the athlete list, current plans and plan
  details. Choose how long they stay and tell me.
- Offline, a banner says so, and every save button is disabled with a reason. Editing
  offline is a separate decision: it needs a save queue and conflict rules on the
  backend.
- Sign-out clears the persisted data.
```

**Checkpoint:** Open home and a few plans, switch the phone to airplane mode, and close and reopen the app: home and those plans open, with the offline banner, and saving is disabled. Sign out and back in: nothing from before is shown.

**In short:** a weak signal in the gym doesn't stop the coach reading a plan.

---

### Print or save a day as PDF

⏸ **Deferred** — start when a public athlete needs more than a screenshot.

**Description:** The brief asks for a PDF only if it's easy ([stakeholder-brief.md](../stakeholder-brief.md) §4–5). The easy route is a print stylesheet plus the browser's "Save as PDF". Generating the PDF in code is not easy, because many PDF libraries don't join Persian letters or lay out RTL text.

**Prompt:**
```
Add a print stylesheet so the browser's print / Save as PDF gives a clean plan: the whole
plan with every day, one after another, with no app frame, buttons or bottom bar, and the
athlete's name and plan title on top. Add a "PDF" action to the day view that opens the
print dialog. Check the Persian text is joined and right to left in the saved PDF on the
coach's phone; if mobile Save as PDF can't do it, say so before going further.
```

**Checkpoint:** A PDF saved from the coach's phone reads correctly, with joined Persian letters, every day of the plan and supersets grouped.

**In short:** a public athlete can get a proper PDF, made by the browser.

---

### Builder: per-side reps, rep schemes, tempo and day notes

⏸ **Deferred** — start when *Plan prescription detail* in [backend.md](backend.md) is done.

**Description:** The board's builder offers هرپا / هرسمت next to تکرار and ثانیه, but the API has no per-side field yet, so the builder leaves them out. Once the backend task adds `per_side`, `tempo`, a rep scheme and day notes, the builder and the day view show and edit them.

**Prompt:**
```
Add the fields from the Plan prescription detail task to the builder and the day view:
the هرپا / هرسمت options on a line (per_side), a rep scheme shown in place of reps when
set, tempo under the prescription, and the day's note at the top of the day. Read the ADR
that task records for the rep-scheme shape first. Update the reducer and its tests, and
the prescription text tests.
```

**Checkpoint:** Day A of the coach's plan shows "۱۵ تکرار هرپا" on the hip thrust only, and day E shows its "با حداقل استراحت" note, both entered in the builder.

**In short:** the builder says "per leg" and "10+10+10" the way the coach writes them.

---

### Warm-up and cool-down strips

⏸ **Deferred** — start when the rule for telling a warm-up or cool-down block apart is decided.

**Description:** Every board shows the opening and closing cardio as cream strips outside the numbering. The rule for spotting them is open ([design/README.md](../design/README.md), "Gaps"): the same bike movement opens and closes a day, so category alone can't tell a warm-up from a cool-down. Position, or a `kind` on `block`, are the candidates. If it's a column, it's a backend task first.

**Prompt:**
```
Show warm-up and cool-down blocks as strips outside the numbering in the day view and the
builder, using the rule decided for it (read its ADR first). Numbering skips them. In the
builder, a line can be marked or unmarked as warm-up / cool-down if the rule is a stored
field.
```

**Checkpoint:** A day that starts and ends with a bike block shows two cream strips and numbers the rest from ۱.

**In short:** the day reads like the coach's sheets, cardio outside the count.

---

### Pre-rendered public pages

⏸ **Deferred** — start when there's a landing page to sell GoGym to other coaches, or when plan share links (`/p/<token>`, [product-direction.md](../product-direction.md)) need previews in WhatsApp and Telegram.

**Description:** The app sits behind a login, so search engines can't see it, and it doesn't need SEO. Public pages are different. React Router v7's framework mode can pre-render chosen routes to static HTML with `ssr: false`, so Caddy still serves plain files with no Node server (ADR 0030).

**Prompt:**
```
Move the app from React Router's library mode to framework mode with ssr: false, and
pre-render the public pages (list them) to static HTML with their title, description and
Open Graph tags. The logged-in app stays a client-side SPA. Check the Caddy fallback
still serves the app's routes. Share-link previews need per-plan tags, which pre-rendering
can't produce; if those are the trigger, propose the server side for them first.
```

**Checkpoint:** `curl` on a public page returns its content and meta tags in the HTML, without JavaScript. The app's routes behave as before.

**In short:** public pages can be found and previewed, without running a Node server.

---

### End-to-end tests for the builder

⏸ **Deferred** — start when a builder bug reaches the coach that the reducer tests didn't catch.

**Description:** The builder's logic is in a reducer with unit tests. What those tests can't catch is the browser: the keyboard covering the search box, a save that doesn't reach the API, a dialog that never closes. Playwright with a phone viewport against the Compose stack would.

**Prompt:**
```
Add Playwright to web/ with one mobile viewport (iPhone and Android profiles), running
against docker compose: create a coach with coachctl, sign in, add an athlete, create an
empty plan, build a day with a superset in the builder, save, and check the day view.
Add it to CI as its own job after the existing ones, with pinned versions.
```

**Checkpoint:** The test passes locally and in CI, and fails if the builder's save request is broken on purpose.

**In short:** the most important flow is tested the way the coach uses it.

---

### Dark theme

⏸ **Deferred** — start if the coach asks for dark (the open question in [stakeholder-brief.md](../stakeholder-brief.md) §6), or if the light theme is hard to read under gym lighting.

**Description:** ADR 0029 chose one light theme and drew no dark variant; the spec's original dark line was there for gym lighting. Since every color comes from the tokens, a dark theme is a second set of token values plus a design pass.

**Prompt:**
```
Design a dark variant of the concept 8 tokens first (a board in docs/design/src, checked
against 4.5:1 for text), and ask me to approve it. Then add it to tokens.css under
prefers-color-scheme: dark, plus a setting to force light or dark. Check every screen, the
chart and the sheets at 360px.
```

**Checkpoint:** With the phone in dark mode, every screen is dark and every text color passes 4.5:1.

**In short:** the app can follow the phone into dark mode.
