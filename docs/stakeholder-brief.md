# GoGym — Stakeholder Brief (the coach)

> **What this is:** how the coach works and what they need from the app, the input for the
> *Frontend stack decision* in [tasks/backend.md](tasks/backend.md). §0 is the coach's own words.
> §1–4 fill it out from follow-up answers collected on 2026-10-17. §5 turns the needs into
> constraints for the frontend and raises questions. It doesn't answer them: the stack decision
> does.

## 0. In the coach's own words

> My way of working is a notebook where I've written the plans of private
> athletes, split with colorful sticks. It's very important for me to find an athlete's plan
> ASAP, because sometimes the gym is very crowded and I'm handling 15–20 athletes at a time, so
> I have to see the athlete's next movement very quickly.
>
> Another thing that frustrates me a lot is writing plans. Right now it takes me so long to
> write a plan. I need it to be as fast as possible too.

## 1. How I work today

I'm a personal trainer with around 20–30 athletes, in two groups:

- **Private athletes** train with me in the gym. Their plans live only in my notebook, and they
  don't get a copy. When they're on the floor, they ask me what's next and I look it up.
- **Public athletes** get one plan from me as a PDF that I make in Canva. After that I don't
  manage them, but they sometimes ask about a movement in their plan.

A plan takes me **about an hour** to write.

## 2. Where and how I use it

- **On my phone, for everything.** That includes building whole plans.
- **My internet is reliable.** It's fine if the app needs a connection.
- **I want it on my home screen** with an icon, so it opens like an app. I don't need an app
  store.
- **Persian, right-to-left, with Jalali dates and Persian digits** (۱۴۰۵/۰۷/۲۵, not
  2026-10-17).

## 3. What I need most

1. **Find an athlete's plan instantly in a crowded gym.** With 15–20 athletes on the floor, an
   athlete tells me which day of the plan they're on, and I need that day open right away so I
   can read the next movement. I don't need the app to track where they are in the session.
   - In the notebook, colored sticks let me jump straight to one athlete's plan. The app has to
     be at least that quick.
   - My home screen should show **only my private athletes**, the ones I'm handling in the gym.
   - Private and public athletes should be **clearly separated**.
   - I want to **search by name**. Public athletes are mostly looked up that way, when they
     message me about a movement.
2. **Write plans fast.** An hour per plan has to drop to **under 15 minutes**. The slow parts
   are:
   - **Writing every detail by hand:** movement name, sets, reps and rest, over and over.
   - **Rewriting similar plans:** most plans are close to one I've already written, and I still
     start from scratch.
3. **Edit a plan when it needs to change.** I have to be able to change an existing plan, not
   only write a new one.
4. **Move my notebook in on day one.** When I start, I'll type every private athlete's current
   plan into the app, roughly 20–30 plans back to back. If the builder is slow, that day decides
   whether I keep using it.

Measurements and the movement library matter too, but I use them less often.

## 4. What the athlete gets from me

- **Private athletes:** nothing to receive. I read their plan from my phone.
- **Public athletes:** for v1, a screenshot of the plan sent on WhatsApp/Telegram instead of the
  Canva PDF, so the plan screen has to look good enough to send. A shareable PDF would be nice,
  but only if it's easy to build.
- Athletes don't need their own login or link in v1.

## 5. What this means for the frontend

Derived from the sections above. These are the questions the stack decision has to answer. None
of them is answered here.

| Need | Constraint on the frontend |
|---|---|
| Plan lookup mid-session, in a crowd | The path from opening the app to a given day of a private athlete's plan is the most important flow. Target: open the app, tap the athlete, tap the day. Tapping an athlete should land on their **current plan** with day tabs, not on a profile or a plan list. The app is used one-handed between athletes, so it needs large touch targets and readable text at a glance. |
| Private-only home, public separate, name search | The API already supports this: `GET /api/v1/athletes` defaults to private athletes, takes `athlete_type=public\|all`, and searches by name with `q`. The UI needs the two groups visually distinct, with public athletes one tap away from home. |
| Plan in under 15 minutes, on a phone | That's roughly 25–50 movement rows (4–5 days × 6–10 movements) in under 15 minutes, so about 20–30 seconds per movement including search. The builder has to minimize typing: pick movements from search, use steppers or presets for sets/reps/rest instead of free typing, carry values over from the previous block, and avoid a full page reload per added movement. |
| Similar plans rewritten from scratch | Starting a plan from an existing one (duplicate, then edit) is what removes most of the time. **New scope:** [spec.md](spec.md) doesn't include it, and [product-direction.md](product-direction.md) puts templates after v1. It needs a backend endpoint and depends on plans being editable (see the gap below). |
| 20–30 plans entered on day one | Builder speed matters from the very first session, not after some use. It's a reason to test the builder with a real plan on a phone before building other screens. |
| Public athletes get a screenshot (a PDF only if easy) | One day of a plan should read clearly in a single phone screenshot, with supersets visibly grouped. It replaces a Canva PDF, so the bar is high. For a PDF, the easy route is a print stylesheet plus the browser's "Save as PDF". Generating the PDF in code is not easy: many PDF libraries (JS and Go) don't join Persian letters or lay out RTL text correctly. |
| Plans must be editable | The builder doubles as the editor: change sets/reps/rest, swap, reorder or remove a movement, and add or remove a block or day in an existing plan. Not supported by the API today (see the gap below). |
| Reliable connection | Online-only. No offline storage or sync. |
| Home-screen icon | PWA: web app manifest, icons, HTTPS. A service worker is only needed if the target browsers require one for install, not for offline caching. |
| RTL Persian | The layout is `dir="rtl"` from the start, with a font that renders Persian well. Charts (the measurement trend) and icons that imply direction need mirroring checked. |
| Jalali dates, Persian digits | The API speaks ISO Gregorian dates and Latin digits. Converting between them is the frontend's job: Jalali display and a Jalali date picker. **Gotcha:** a Persian phone keyboard types ۱۲ and not 12 in number fields, so inputs must normalize digits before sending, or the API rejects them. |
| Theme | **Open.** [spec.md](spec.md)'s design constraints say dark/neutral and high contrast for gym lighting. |

### Gaps found against the current API

- **Plans can only be added to.** `/plans`, `/plans/:id/days`, `/days/:id/blocks` and
  `/blocks/:id/movements` only accept `POST`. There is no edit, reorder or delete for days,
  blocks or block movements, and no plan update. The coach needs plans to be editable (§3), so
  v1 needs update, delete and reorder for each level. Duplicating a plan depends on this too:
  a copy only helps if it can then be edited.
- **No way to duplicate a plan.** See the table above.

## 6. Open questions for the coach

- Dark, light, or follow the phone's setting? (To be answered later.)
