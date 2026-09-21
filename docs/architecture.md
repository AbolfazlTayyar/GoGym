# Architecture

How the codebase is shaped, and the precedent later feature modules should follow. This
documents decisions already made in code — if it and the code disagree, the code wins; fix
this doc.

## Module layout

Each feature lives in its own `internal/<module>` package, split by responsibility:

| File | Responsibility |
|---|---|
| `model.go` | GORM struct(s) for the module's table(s), tagged to match `migrations/*.sql` exactly. |
| `repository.go` | Persistence — the only place that talks to `*gorm.DB` for this module's tables. |
| `service.go` | Business logic: validation, orchestration, anything that isn't "shape an HTTP response" or "run a query." |
| `handler.go` | Gin handlers, request/response DTOs, and their swagger annotations (see [CLAUDE.md](../CLAUDE.md)). |
| `routes.go` | A `RegisterRoutes` function that mounts the module's endpoints onto the router groups it's given. |

`internal/coach` is the reference implementation of this shape (added in the coach-auth task,
`docs/tasks.md` **[A1]**) — copy its structure rather than inventing a new one.

## Wiring a module in

`internal/server/server.go` is the only place that constructs modules and mounts routes. The
pattern, per module:

```go
repo := <module>.NewRepository(gormDB)
svc := <module>.NewService(repo, ...)
handler := <module>.NewHandler(svc, repo)
<module>.RegisterRoutes(v1, protected, handler)
```

`v1` is the `/api/v1` group; `protected` is the same group with `coach.AuthMiddleware` applied,
for every route that isn't `/auth/*`. Register unauthenticated routes on `v1` directly and
everything else on `protected`.

## Tenant isolation

Every coach-owned table carries `coach_id`, and `internal/tenant` is the single place that
scoping decision is made. The reason is leak-prevention, not future-proofing: a rule enforced
by discipline across a dozen handlers is one forgotten `WHERE` away from serving one coach's
athletes to another.

- `tenant.Scope(db, coachID)` returns `db.Where("coach_id = ?", coachID)`. Every repository
  query against a tenant-owned table must be built from this, never a hand-written
  `.Where("coach_id = ?", ...)` — see [CLAUDE.md](../CLAUDE.md#data-access-conventions).
  Constructing an unscoped query for a tenant-owned table should look visibly wrong in review.
- `tenant.SetCoachID` / `tenant.CoachIDFromContext` move the authenticated coach's id through
  the Gin context; `coach.AuthMiddleware` sets it after validating the JWT, handlers/services
  read it from there.
- A lookup that's out of scope for the authenticated coach (another coach's row) must 404, not
  403 — don't leak that the row exists at all.
- Postgres row-level security was considered as defense in depth and deliberately deferred
  (see the doc comment on `internal/tenant/scope.go`): v1 has exactly one enforcement path —
  this package — so a second, DB-level copy of the same rule is redundant until something else
  (a background job, an admin console) can construct a query against tenant-owned tables
  outside this path.
- An organization layer above coach is **not planned** — the app stays per-manager (decided
  2026-09-20, superseding the hedge in `docs/product-direction.md` §7). The single choke point
  earns its keep on leak-prevention alone; that it would also make an `org_id` switch a
  one-file change is a side benefit, not a roadmap item. Don't design around it.

## Auth

`internal/coach` owns the one login this app has: signup, login, JWT issuance/validation, and
`AuthMiddleware`. Passwords are bcrypt-hashed, never logged or stored in plaintext. `/auth/*`
is rate-limited in-process, keyed by both IP and phone (`internal/coach/ratelimit.go`) — a
simple sliding-window limiter, sufficient for v1's single-instance deployment; a
multi-instance deployment would need a shared store instead.
