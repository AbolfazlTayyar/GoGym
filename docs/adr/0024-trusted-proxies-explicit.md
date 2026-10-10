# 0024 — Forwarded client IPs are trusted only from configured proxies

- Status: accepted
- Date: 2026-10-10
- Deciders: Abolfazl Tayyar

## Context and Problem Statement

[0010](0010-coach-jwt-auth.md) rate-limits `/auth/*` per IP using `c.ClientIP()`. Gin's default
trusts `X-Forwarded-For` and `X-Real-IP` from every peer, so any caller can name its own IP and get
a fresh bucket on each request. The VPS deployment will put Caddy in front, and Caddy's IP will be
the peer on every request, so the real client IP has to come from its header there.

## Decision Drivers

- A caller must not be able to choose the IP its requests are counted under.
- Behind Caddy, clients must not all share Caddy's bucket.
- Pointing the app at a proxy shouldn't need a code change.

## Considered Options

1. Keep Gin's default and trust every peer
2. Hard-code `SetTrustedProxies(nil)` and change it in code when Caddy arrives
3. An optional `TRUSTED_PROXIES` env var (IPs/CIDRs), empty meaning trust none

## Decision Outcome

Chosen option: **3**. Option 1 is the bypass. Option 2 is safe today, but the deploy step that
undoes it is easy to forget, and forgetting it fails quietly: every client lands in Caddy's bucket,
so one caller can lock everyone out of login. With option 3 the safe default needs no setup, and
the deploy sets the proxy in the same place as the rest of its config. `server.New` returns an
error for an invalid entry, so a typo stops startup instead of trusting nothing or everything.

### Consequences

- Good: the per-IP auth limit holds whether the API is reached directly or through a proxy.
- Bad: behind a proxy the setting is required, and nothing fails if it's missing. The deploy
  runbook has to set it.
