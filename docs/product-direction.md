# GoGym — Product Direction (post-v1)

> **Status: hypothesis, not a spec.** Nothing here is committed scope. This document exists to
> capture the shift in intent — from "a tool for one coach I know" to "a product other coaches
> pay for" — and to reason about what that changes, before v1 is finished. It deliberately does
> **not** supersede [spec.md](spec.md); the v1 loop described there is still the right
> core and should be built as written.
>
> Anything in here that must be decided *before* v1 ships has already been folded into
> [tasks.md](tasks.md) as the tasks marked `[A1]`–`[A4]`. Treat those tasks as actionable and
> this document as the argument behind them.

## The thesis

v1 as specced is a **plan-authoring tool**. A coach builds a plan in it and then screenshots it
into WhatsApp. Nothing comes back from the athlete, ever. The database only records what the
coach *intended*, never what happened.

That is a fine internal tool and a hard product to sell, because it competes with Word and Excel
on a feature ("a nicer way to type a plan") that Word and Excel already do for free. Coaches pay
monthly for two things:

1. **Time saved on repeat work** — they write the same plan with small variations dozens of times.
2. **Looking professional to their athletes** — the product is a visible part of their service.

Almost everything below is downstream of those two sentences.

## Gaps, by how much they affect willingness to pay

### 1. The athlete never sees the product

`Athlete` is a record, not a user, and there is no athlete-facing surface at all. This caps the
product's value and kills the cheapest growth channel available: every athlete who opens a plan
with the product's branding on it is a free impression, and some of them coach too.

**This does not require athlete accounts.** The cheap version is a signed share link per plan
(`/p/<token>`), mobile web, RTL, no login. The coach taps "share" and sends it on
WhatsApp/Telegram; the athlete opens it in the gym on their phone. That is days of work against
the athlete-auth system [spec.md](spec.md) rightly cut from v1 — and it is most of the
value. An athlete account then becomes an upsell rather than a prerequisite.

### 2. Nothing comes back from the athlete

There is no entity anywhere for "the athlete actually did this." Consequences:

- No retention loop — no reason to open the app between plan-writing sessions.
- No dashboard worth looking at (the spec even cuts per-athlete quick status, correctly, because
  with no inbound data there is nothing to put on it).
- Measurements are the only feedback signal, and they are manual and roughly monthly.

Workout logging (actual reps, actual load, RPE, date, done/skipped) is what turns the measurement
chart from a vanity graph into *"Ali has not trained in 9 days"* — the notification that makes a
coach keep the tab open.

**Design constraint this creates, noted here because it bites early:**
[er-diagram.md](er-diagram.md) records that `Day`/`Block`/`BlockMovement` carry no `updated_at`
because the plan builder deletes and recreates them. If logs foreign-key to `block_movement_id`,
every plan edit orphans training history. Logs should point at `movement_id` plus a denormalized
snapshot of what was prescribed (sets/reps/load as programmed), not at the plan structure. This
is recorded as a required comment in the Plan builder task in [tasks.md](tasks.md).

### 3. Templates — the highest-ROI feature for a paying coach

Every plan in v1 is built from scratch through the builder. Real coaches reuse relentlessly.
"Duplicate this plan for another athlete", "save this day as a template", "my template library"
is the feature that converts a trial into a subscription, because it is the one that visibly
gives a coach their Friday night back.

This is also the second reason to revisit delete-and-recreate plan editing: templates imply plans
are copied, versioned, and edited in place.

### 4. Data model holes that block real use

Reading [er-diagram.md](er-diagram.md) as a working coach would:

| Gap | Why it matters |
|---|---|
| `block_movement` has `reps` and `duration_seconds` but **no load** | Coaches program weight — kg, %1RM, or RIR. This is a hole, not a v2 feature. |
| No per-set variation | "3 sets of 12/10/8" is the normal case, not the exception. One `reps` int cannot express it. |
| No tempo | Standard programming vocabulary; cheap to carry. |
| `movement` has **no video/media** | "How do I do this one?" is the message every coach gets daily. A library without demo clips is not competitive. Pulls in object storage — an infra decision much cheaper to make before launch than after. |
| `movement` has only `category` | Muscle group + equipment filters are what make the builder *fast*. Category alone means scrolling. |
| No diet/nutrition plan | In the target market, training and diet plans are typically sold together. Even "a diet document attached to the athlete" is a large perceived-value jump for small effort. |
| No athlete status or contract dates | Active/paused/expired plus subscription start/end. This is what the planned accounting module needs *and* what powers "3 athletes expire this week". |
| No progress photos | The before/after is the coach's own marketing asset. Same check-in event as measurements. |

### 5. Two different billings, currently conflated

[spec.md](spec.md) plans a `billing` module for "the coach calculates monthly income."
That is **coach ↔ athlete** money. Selling the product introduces a second, unrelated one:
**you ↔ coach** subscription state (tier, expiry, athlete quota, invoices). Different lifecycles,
different owners, different failure modes — they should not share tables, and the roadmap's
modular-monolith boundaries should keep them apart from the start.

**Market assumption worth confirming:** the RTL/Persian design constraint implies an Iranian
market, where card-on-file recurring subscriptions are effectively unavailable. If so, model the
coach's subscription as *"an expiry date renewed by a manual gateway payment"* (Zarinpal/IDPay or
similar), not as the webhook-driven state machine you would build against Stripe. This changes the
data model enough to be worth settling before the module is designed.

**Metering:** athlete count is the natural lever the existing schema already provides — free up to
~3 athletes, paid beyond. It aligns price with value and lets a coach trial it on a real client.

### 6. Obligations that begin the moment the data is not yours

These were the original gap list. The first five are now scheduled in [tasks.md](tasks.md) as the
`[A1]`–`[A4]` tasks; the rest are still open and belong to the post-v1 milestone.

**Now covered by a task:**

- **Password reset and phone verification.** Phone is both the login identity and a unique key;
  without OTP you get typos you cannot fix and lockouts resolved by hand. → *Coach auth hardening*
- **Tenant isolation enforced per handler.** Repeating "scope by coach_id" in fifteen tasks means
  one forgotten `WHERE` is a cross-tenant leak; it belongs in one enforceable place. → *Coach auth*
- **No backups and no tested restore.** Losing a coach's athlete history once ends the business.
  → *Backup & restore drill*
- **No soft delete.** A coach who deletes an athlete by accident has no recourse. → *Schema amendments*
- **No rate limiting on auth.** → *Coach auth*

**Still open, nothing scheduled:**

- **No data export.** "You can leave with your data" is a trust argument in the sales conversation
  before it is ever a legal one.
- **No audit log, and no error tracking** (Sentry or equivalent).
- **No product metric.** "Plans built per coach per week" is the number that tells you whether
  anyone actually uses this.

### 7. Who you sell to changes the schema

> **Answered, 2026-09-20: not selling to gyms.** The product stays per-manager, so the org
> layer below is not planned work. The single-helper hedge was still built — `internal/tenant`,
> see [architecture.md](architecture.md) — but for leak-prevention, which stands on its own,
> not as a step toward `org_id`. The reasoning is kept because the market may say otherwise
> later; nothing should be designed around it in the meantime.

One freelance coach is a low-value customer. A gym with five coaches — shared movement library,
athlete transfer between coaches, owner-level reporting — is a far better one.

[spec.md](spec.md)'s claim that "multi-coach later is a permissions feature, not a
schema migration" is optimistic: a gym needs an org layer *above* coach, not more coaches beside
each other. The recommendation is **not** to add an `organization` table now, but to route every
tenancy check through a single helper so that the day tenancy becomes `org_id`, it is one file
instead of every repository. That is the cheap hedge.

## Sequencing

1. **Finish v1 as specced.** The loop in [spec.md](spec.md) is the core of any version
   of this product. Do not stop and redesign.
2. **Do the `[A1]`–`[A4]` tasks in [tasks.md](tasks.md)** — the things that are free now and expensive
   later. These attach to existing v1 tasks; they are not a separate phase.
3. **Then the "sellable" milestone**, in this order:
   1. Plan share link (athlete-visible, no login)
   2. Templates / duplicate plan
   3. Athlete workout logging
   4. Diet plan
   5. Athlete contract dates → coach accounting module
4. **Later, and only if the market says so:** athlete accounts, gym/org layer, subscription
   billing for the product itself, notifications.

A task list for step 3 gets written once v1 is running and the open questions below have real
answers — not before. Writing it now would produce a document that reads as authoritative while
being entirely untested.

## Open questions — validate before building any of the above

- Would the coach this was built for pay for it? At what price? What would they *stop* paying for?
- Find **two coaches who are not friends** and ask the same. Three data points beats this document.
- Is the market assumption in §5 (Iran / RTL / no recurring card payments) correct?
- Freelance coaches or gyms? §7's schema hedge is cheap; committing to the gym model is not.
- Does the athlete want an app, or does the athlete want a link? Assume link until proven otherwise.

> Everything above is reasoned from these documents, not from users. Treat the ordering as a
> hypothesis to test, not a plan to execute.
