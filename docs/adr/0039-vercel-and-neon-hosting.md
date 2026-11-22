# 0039 — The app and API are hosted on Vercel, with Postgres on Neon

- Status: accepted
- Date: 2026-11-22
- Deciders: Abolfazl Tayyar
- Supersedes: [0030](0030-react-spa-same-origin.md), for where the app is served only. The React
  SPA itself, and serving it on the API's origin, still stand.

## Context and Problem Statement

Nothing is deployed yet. [0030](0030-react-spa-same-origin.md) and `docs/tasks/deployment.md`
planned a VPS running `docker-compose.yml` with Caddy in front, serving the built app and proxying
`/api` to Go. The domain is ready, and Vercel plus Neon removes running a server: no OS updates,
TLS, Docker or backups to look after for one coach and 20–30 athletes.

## Decision Drivers

- One origin for the app and the API, so the `SameSite=Strict` cookie and no-CORS rule of
  [0031](0031-auth-token-in-httponly-cookie.md) hold unchanged.
- No server to run or patch.
- Moving off Vercel later must not touch feature code.

## Considered Options

1. A VPS with Docker Compose and Caddy, as planned in 0030
2. Vercel for the app only, with the API and Postgres on another host behind a Vercel rewrite
3. One Vercel project for the app and the API (a Go function), with Postgres on Neon

## Decision Outcome

Chosen option: **3**. Option 2 needs two providers and leaves the API reachable at its own
address, which skips the proxy; it also can't pin a proxy IP for `TRUSTED_PROXIES`. Option 1
stays the fallback and is what leaving Vercel means (below).

How it fits together:
- `vercel.json` builds `web/` into the static output and rewrites `/api/*` and `/healthz` to one
  Go function, `api/index.go`; every other path falls back to `index.html`.
- `api/index.go` builds the same router as `cmd/api` through `server.New`, once per function
  instance. Feature code doesn't know it runs on Vercel.
- The API uses Neon's pooled host. Migrations run from a laptop with `make migrate-up`, with the
  `DB_*` variables pointing at Neon's direct (non-pooler) host, before deploying code that needs
  them. The first coach is created the same way with `coachctl` ([0025](0025-coach-accounts-by-cli.md)).
- The function's peer is Vercel's runtime, so [0024](0024-trusted-proxies-explicit.md)'s
  `TRUSTED_PROXIES` can't name it. `api/index.go` sets Gin's `TrustedPlatform` to `X-Real-IP`,
  which Vercel's edge overwrites with the caller's IP. Only that entry point sets it: anywhere
  else a caller could send the header and pick its own rate-limit bucket.
- `SERVER_PORT` is still required by the config but unused on Vercel.

### Consequences

- Good: no server, TLS or Docker to run; deploys are a push.
- Good: the change is two files (`api/index.go`, `vercel.json`); `cmd/api`, the Dockerfile and
  Compose still work for local dev.
- Bad: the auth rate limiter ([0010](0010-coach-jwt-auth.md)) keeps counts in memory, so each
  function instance counts on its own. Add a Vercel Firewall rate-limit rule on `/api/v1/auth/*`.
- Bad: Neon's free tier suspends compute when idle, so the first request after a quiet spell
  takes about a second longer.
- Bad: Vercel's Hobby plan is for non-commercial use; a paying coach means the Pro plan.
- Bad: migrations are a manual step from a laptop, and forgetting one fails at request time.

## Leaving Vercel

Moving to another host (the VPS of option 1, or anything that runs a long-lived process) means
undoing this decision's changes, in a new ADR that supersedes this one:

1. Delete `api/index.go` and `vercel.json`, and the `.vercel/` line in `.gitignore`.
2. Run the API from `cmd/api` (the existing Dockerfile) and serve `web/dist` with `/api/*` and
   `/healthz` proxied to it on the same origin, as 0030 describes for Caddy.
3. Set `TRUSTED_PROXIES` to the proxy's fixed IP (0024). Don't carry over the `X-Real-IP`
   `TrustedPlatform` line: off Vercel nothing overwrites that header.
4. Move the data: `pg_dump` from Neon, restore into the new Postgres, then point `DB_*` at it.
5. Point the domain's DNS at the new host and remove it from the Vercel project.
6. Drop the Vercel Firewall rule; the in-memory rate limiter works again with one process.
