# 0026 — /healthz caches its database ping instead of hiding behind a secret path

- Status: accepted
- Date: 2026-10-11
- Deciders: Abolfazl Tayyar

## Context and Problem Statement

`/healthz` is unauthenticated and pinged the database on every request, so a flood of it would
compete with real traffic for the connection pool. A planned deployment task would have made
Caddy answer 404 on the plain path and serve the check only on a random path kept in a secret env
var, with a fail-closed compose check for an unset value and a rotation runbook. A flood of 404s
or of `/api/v1/auth/login` requests costs the server about as much as a flood of health checks,
so hiding the path barely changes the exposure.

## Decision Drivers

- A flood of `/healthz` can't multiply database pings.
- No extra secret, Caddy config or runbook to maintain.
- The uptime monitor and container healthcheck keep using the plain path and still notice an
  outage within seconds.

## Considered Options

1. Serve the check only on a secret path through Caddy
2. Leave `/healthz` as it is
3. Cache the ping result for a few seconds inside the handler

## Decision Outcome

Chosen option: **3**. Option 1 adds a secret, a rewrite rule and a rotation process to protect
an endpoint that costs no more to flood than any other path. Option 2 lets every request start a
ping. With option 3, any number of requests within the window share one ping. Concurrent callers
wait on the same ping rather than each starting their own, and the ping is detached from the
caller's context, so a client that hangs up isn't cached as an outage.

### Consequences

- Good: at most one database ping per 5 seconds, however often the endpoint is hit, with no
  infra change.
- Good: a cached failure still attaches its cause, so each 503 logs the real error.
- Bad: the check can report up to 5 seconds late, in both directions. Probes run every 10 seconds
  or more, so this is negligible.
- Bad: the endpoint stays public. Flooding it still uses CPU and bandwidth, as flooding any other
  path does. Rate limiting at the proxy is the answer if that ever matters.
