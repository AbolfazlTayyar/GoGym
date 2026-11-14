# 0035 — A plan is shared through one stored, expiring token

- Status: accepted
- Date: 2026-11-14
- Deciders: Abolfazl Tayyar

## Context and Problem Statement

Coaches send plans to athletes on WhatsApp and Telegram, today as a Canva PDF or a screenshot.
Athletes have no accounts ([product-direction.md](../product-direction.md) §1), so a link has to
open the plan without a login: `/p/<token>`, read through `GET /api/v1/public/plans/:token`. The
link sits in chats we can't reach once it's sent, so how a token is made and ended can't change
later without breaking links already sent.

## Decision Drivers

- A link can be ended: it expires, and the coach can replace it.
- No athlete accounts, and no signing keys to rotate.
- An expired link still tells the athlete which coach to ask.
- The tenant rules stay intact everywhere except this one read.

## Considered Options

1. A random token stored on `plan` (`share_token`, `share_expires_at`), replaced on every share
2. A token signed with a server key (the plan id and an expiry), with nothing stored
3. A `plan_share` table with one row per share, so a plan can have several live links
4. A stored token that never expires

## Decision Outcome

Chosen option: **1**. `POST /plans/:id/share` writes a new 128-bit random token (base64url, 22
characters) and an expiry 30 days out, so the plan's previous link stops working at once. A plan
has at most one live link, and the coach re-shares to extend it. Option 2 can't be revoked short
of rotating the key, which would end every coach's links. Its expiry is fixed once the link is
sent, and the server-side rule for re-sharing would need a column anyway. Option 3 adds a table for
several live links per plan, which one coach sending one plan to one athlete doesn't need. Option 4
leaves a plan readable for good by whoever ever got the link.

`GET /public/plans/:token` has no coach to scope by, so it is the one read outside
`tenant.Scope`: `athlete.Repository.FindForSharedPlan`, used by this path only. An expired token
answers 200 with `status: "expired"`, the coach's name and phone, and null `athlete` and
`plan`. It isn't a 410, because an error envelope carries no data, and the page has to name the
coach. An unknown or replaced token, a deleted plan, and a soft-deleted athlete are all a 404.
The response carries only what the page shows. It leaves out the athlete's phone and notes,
and the source plan's title, which may belong to another athlete. The endpoint is rate-limited
per IP and sent with `X-Robots-Tag: noindex`.

### Consequences

- Good: a link can be ended, by time or by re-sharing, without keys or extra tables.
- Good: the bypass of tenant scoping is one named method on one path.
- Bad: re-sharing to a second person ends the first person's link.
- Bad: anyone holding a live link sees the coach's phone and the athlete's name; that is the
  point of the link, so the page carries nothing else personal.
