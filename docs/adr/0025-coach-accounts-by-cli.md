# 0025 — Coach accounts are created by a CLI, not a public signup endpoint

- Status: accepted
- Date: 2026-10-11
- Deciders: Abolfazl Tayyar

## Context and Problem Statement

`POST /api/v1/auth/signup` was open with no gate, because v1 has one coach. Once the API is
deployed, anyone who finds the URL can create a coach account. The planned fix, OTP phone
verification and password reset by SMS, needs an SMS provider, a new schema and anti-enumeration
work. It protects against a stranger signing up with a mistyped phone, which can't happen while
one coach uses the app. A forgotten password still needs some way back in other than a raw
`UPDATE`.

## Decision Drivers

- Nobody can create an account by finding the URL.
- A password reset doesn't need hand-written SQL.
- No SMS provider or schema change while there is one coach.

## Considered Options

1. Keep public signup and build OTP verification and password reset now
2. An `ALLOW_SIGNUP` env flag, off by default, plus a CLI for password resets
3. Remove the signup route, and create coaches and reset passwords with a `cmd/coachctl` CLI

## Decision Outcome

Chosen option: **3**. Option 1 pays for SMS, a schema and anti-enumeration work to solve a problem
one coach doesn't have. Option 2 still needs a CLI for resets, and the signup route stays in the
production binary with only a flag in front of it. One wrong env value reopens it, and nothing
fails when that happens. With option 3 there is no route to reopen. The CLI goes through
`coach.Service`, which now holds the phone and password rules the signup handler used to apply.
It reads the password from stdin so the password stays out of shell history and the process
list. It ships in the same image as the API, so it runs on the VPS with `docker compose exec`.

### Consequences

- Good: no unauthenticated way to create an account, and a password reset is one command.
- Good: when signup returns, it is a thin handler over `coach.Service.Create`, which already
  validates input.
- Bad: every new coach and every reset needs someone with shell access to the server. That is
  acceptable for one coach and is the trigger for the deferred *Coach auth hardening* task.
- Bad: a reset doesn't end existing sessions; tokens already issued stay valid until they expire
  (see [0010](0010-coach-jwt-auth.md)).
