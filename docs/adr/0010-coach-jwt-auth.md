# 0010 — Coach-only JWT auth with in-process rate limiting

- Status: accepted
- Date: 2026-09-20
- Deciders: Abolfazl Tayyar

## Context and Problem Statement

v1 has one kind of login: the coach. Athletes are records, not users. Login endpoints that
accept a password must not be open to unlimited guessing.

## Decision Drivers

- Stateless auth that's simple to run on one instance.
- Never store or log plaintext passwords. Don't reveal whether the phone or the password was wrong.
- No extra infra (Redis) for v1.

## Considered Options

1. Server-side sessions
2. JWT (bcrypt-hashed passwords) + in-process sliding-window limiter on `/auth/*`
3. JWT + Redis-backed limiter

## Decision Outcome

Chosen option: **2**. `internal/coach` owns signup, login, token issuance and `AuthMiddleware`.
The middleware puts the coach id into the Gin context for tenant scoping
(see [0011](0011-tenant-isolation-single-helper.md)). The limiter is keyed by both IP and phone,
and it sweeps expired keys so memory doesn't grow forever.

### Consequences

- Good: no session store or Redis, and it's easy to test.
- Bad: tokens can't be revoked before they expire. The limiter is per-process, so running more
  than one instance needs a shared store.
