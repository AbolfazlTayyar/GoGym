# GoGym — MVP Task List v1

Stack locked in for this list: **Go + Gin + GORM + PostgreSQL**, migrations via `golang-migrate`, JWT + bcrypt auth, module-first / clean-architecture-per-module layout (`internal/<module>/{handler,service,repository,model}.go`) per [prestart-roadmap.md](prestart-roadmap.md) step 2. API docs via **swaggo (Swagger UI)** for manual endpoint testing — see the *Swagger / OpenAPI docs setup* task. Frontend stack is **not yet decided** — this list covers scaffolding and backend API only; frontend tasks get written once that's chosen (see *Frontend stack decision* below).

Each task below has: what it is, a plain description, a ready-to-paste prompt for Claude Code, and a checkpoint to verify it's actually done before moving on. Work top to bottom — later tasks assume earlier ones are merged.

---

## Phase 0 — Foundation

### Initialize Go project structure & tooling

**Description:** Set up the base Go module, standard folder layout, and a Makefile for the commands you'll run constantly (build, run, test, lint).

**Prompt:**
```
Initialize the GoGym Go module at the repo root. Set up:
- go.mod (module path: check with me for the right github.com/<user>/gogym path before creating)
- Standard layout: cmd/api/main.go (entrypoint, currently just starts an empty process),
  internal/ (empty for now, modules land here in later tasks), pkg/ only if something
  is genuinely reusable outside this repo (don't create it speculatively)
- A Makefile with targets: build, run, test, lint, migrate-up, migrate-down (migrate
  targets can be stubs for now, the migrations task fills them in)
- .gitignore for Go (binaries, .env, vendor if not committed)

Follow the module-first, clean-architecture-per-module structure described in
docs/prestart-roadmap.md step 2 — internal/<module>/{handler,service,repository,model}.go
— but don't create any module folders yet, that starts with the coach auth module.

Don't add any dependencies yet beyond what cmd/api/main.go needs to compile and run.
```

**Checkpoint:** `go build ./...` succeeds, `make run` starts and exits cleanly (or serves nothing yet — that's fine), `git status` shows only the new scaffolding files.

---

### Config loader & environment setup

**Description:** A single place that reads config (DB connection string, JWT secret, port, etc.) from environment variables / a `.env` file, so nothing is hardcoded later.

**Prompt:**
```
Add a config package at internal/config that loads app configuration from environment
variables, with a .env file supported for local dev (use github.com/joho/godotenv or
similar — pick one well-maintained option and tell me which).

Required fields for now: DB connection (host, port, user, password, dbname, sslmode),
server port, JWT secret, JWT expiry duration. Fail fast (panic or os.Exit with a clear
message) on startup if a required var is missing — don't silently default secrets.

Add a .env.example (committed) with placeholder values and confirm .env itself is
gitignored. Wire cmd/api/main.go to load config on startup and log (not print the
secret values) that config loaded successfully.
```

**Checkpoint:** Running `make run` without a `.env` file fails with a clear error naming the missing var; copying `.env.example` to `.env` with real values lets it start and log success. `.env` is confirmed absent from `git status`/`git ls-files`.

---

### Docker Compose for local dev

**Description:** One command spins up the Go app and a Postgres instance together, so local dev matches how it'll eventually be deployed.

**Prompt:**
```
Write a docker-compose.yml at the repo root with two services: `api` (this Go app,
built from a Dockerfile you also write) and `db` (postgres:16, with a named volume
for data persistence, exposing the port locally, and env vars matching internal/config
from the config loader task). Use a multi-stage Dockerfile (build stage with the Go
toolchain, minimal runtime stage) for the api service.

Make sure the api service's DB host in its env matches the db service's Docker Compose
service name, and that api waits for db to be healthy before starting (use a healthcheck
on the db service and depends_on with condition: service_healthy).

Update .env.example if the Compose setup needs additional vars.
```

**Checkpoint:** `docker compose up --build` brings up both containers; `docker compose ps` shows both healthy/running; `docker compose logs api` shows the config-loaded-successfully log line with no crash loop.

---

### Database migrations: create schema

**Description:** Turn the ER diagram into the first real migration — every table, column, type, and foreign key from `docs/er-diagram.md`.

**Prompt:**
```
Set up golang-migrate for this project (migration files under a top-level migrations/
directory, sql format, sequential numbered up/down pairs). Wire the Makefile's
migrate-up / migrate-down targets from the scaffolding task to actually run migrate
against the DB connection from internal/config.

Write the first migration creating all tables from docs/er-diagram.md exactly as
specified there: coach, athlete, athlete_measurement, plan, day, block, block_movement,
movement — with their columns, types, primary keys (uuid), foreign keys, and the
nullable movement.coach_id noted in the ER diagram and mvp-spec.md (NULL = system-seeded,
universal movement). Add a not-null + a reasonable default where the spec implies one
(e.g. created_at/updated_at timestamps are reasonable to add even though not listed
explicitly — ask me before adding anything not in the ER diagram if you're unsure it's
warranted).

Per CLAUDE.md, docs/er-diagram.md must be updated in the same commit as any schema
change — since this migration matches the existing diagram exactly, no diagram edit
should be needed; flag it to me if you find a mismatch instead of silently resolving it.
```

**Checkpoint:** `make migrate-up` against the Compose Postgres succeeds; `\dt` in `psql` lists all 8 tables; `make migrate-down` cleanly drops them and `migrate-up` again is idempotent (reruns cleanly from empty).

---

### GORM models & DB connection layer

**Description:** Go structs for every entity, mapped to the migrated schema, plus the shared GORM connection setup other modules will reuse.

**Prompt:**
```
Add a shared internal/db package that opens and returns a *gorm.DB using the config
from internal/config, with sensible connection pool settings (max open/idle conns,
conn max lifetime) and GORM logger set to a level appropriate for dev vs prod (read
from config, default to warn in prod, info in dev).

Define GORM model structs for all 8 entities from docs/er-diagram.md, placed in each
future module's model.go per the module-first layout — but since modules don't exist
yet, put them temporarily in internal/models/ and note in a comment that they'll move
into internal/<module>/model.go as each module is built. Match column names/types/
nullability exactly to the migration from the schema migration task — don't let GORM
auto-migrate diverge from the hand-written migration; disable GORM's AutoMigrate
entirely, migrations are the single source of truth.

Confirm every GORM tag (gorm:"...") matches the actual migrated column, especially the
nullable movement.coach_id (should be a *uuid.UUID or sql.NullString-equivalent, not a
non-nullable field).
```

**Checkpoint:** A small throwaway `go run` script (delete after) can open the DB via `internal/db`, `.Find()` on each model without error against the empty migrated schema, confirming every struct maps to a real table with no "column not found" errors.

---

### Testing strategy setup

**Description:** Decide and wire up how this project actually gets tested — unit tests with no DB, integration tests against a real Postgres — before feature code starts piling up untested.

**Prompt:**
```
Set up the testing harness described in docs/prestart-roadmap.md step 9:
1. Confirm the standard library testing + testify (assert/require) for unit tests —
   add testify as a dependency if not already present.
2. Add integration test support using testcontainers-go to spin up a throwaway Postgres
   per test run, apply migrations from migrations/ against it, and expose a helper
   (e.g. internal/testutil) that other modules' _test.go files can call to get a
   ready-to-use *gorm.DB.
3. Add a `make test-unit` (short, no containers, safe for tight loops) and `make test`
   (full suite including integration) target.
4. Write one smoke integration test using this harness against the migrated schema
   — e.g. insert a coach row, read it back — purely to prove the harness works, not as
   real feature coverage.

Don't write extensive tests yet; this task is the harness, not coverage. Real tests get
added alongside each feature module starting with coach auth.
```

**Checkpoint:** `make test-unit` runs instantly with no Docker dependency. `make test` spins up a real Postgres container, runs migrations, runs the smoke test, and tears the container down — rerun it twice in a row to confirm no leftover state/port conflicts.

---

### CI pipeline

**Description:** GitHub Actions runs vet, tests, and lint on every push, so broken code can't silently sit on `main`.

**Prompt:**
```
Add a GitHub Actions workflow at .github/workflows/ci.yml that runs on push and pull_request
to main. Steps: checkout, set up Go (matching the version in go.mod), go vet ./...,
golangci-lint (add a .golangci.yml with a reasonable default ruleset if one doesn't
exist), make test-unit, and make test (the full integration suite — GitHub Actions runners
support Docker so testcontainers-go from the testing strategy task should work as-is,
but verify and add any needed Docker-in-Docker setup).

Keep the workflow fast: unit tests and lint should run in a job (or step) that fails
fast before the slower integration job runs, so a lint error doesn't wait on a Postgres
container to spin up first.
```

**Checkpoint:** Push a branch with a deliberate `go vet` failure (e.g. unused import) and confirm the Actions run fails at the lint/vet step before integration tests even start; fix it, push again, confirm a full green run.

---

### Gin router skeleton + health check

**Description:** The actual HTTP server boots, with middleware (logging, recovery, CORS) wired in, and a `/healthz` endpoint to prove it's alive.

**Prompt:**
```
Wire cmd/api/main.go to start a Gin server using internal/config for the port and
internal/db for the DB connection. Add global middleware: gin.Recovery(), a request
logger (structured JSON logs to stdout per docs/prestart-roadmap.md step 10 — use
Gin's logger or a structured logger like zerolog/zap, pick one and tell me which),
and CORS configured permissively for now (frontend origin is still undecided per
mvp-spec.md — note a TODO to lock this down once the frontend stack and its origin
are chosen).

Add GET /healthz that checks DB connectivity (a lightweight ping, e.g. sqlDB.PingContext)
and returns 200 with {"status":"ok"} or 503 with the failure reason if the DB is
unreachable. Group all future feature routes under /api/v1 even though no feature
routes exist yet.
```

**Checkpoint:** `docker compose up`, then `curl localhost:<port>/healthz` returns 200 `{"status":"ok"}`; stop the `db` container and re-curl — should return 503 within a couple seconds, not hang.

---

### Swagger / OpenAPI docs setup

**Description:** A live, browsable API doc (Swagger UI) generated from code annotations, so you can manually poke every endpoint from the browser instead of hand-writing curl commands as you build.

**Prompt:**
```
Add Swagger/OpenAPI support to the Gin API using swaggo (github.com/swaggo/swag for
the doc generator, github.com/swaggo/gin-swagger + github.com/swaggo/files to serve
the UI). Set up:
- General API metadata annotations in cmd/api/main.go (title, description, version,
  base path /api/v1, and a bearer/JWT security definition matching the auth scheme
  from the coach auth task so "Authorize" in the UI works once that task lands).
- Wire GET /swagger/*any to serve the generated Swagger UI, only mounted when running
  in dev (read an env flag from internal/config — don't expose interactive API docs
  in a prod build by default).
- A `make swagger` Makefile target that runs `swag init` to (re)generate docs/swagger
  output from handler annotations, and document in the Makefile or a README note that
  this must be rerun after adding/changing annotated handlers.
- Annotate the existing GET /healthz handler as a working example of the annotation
  format (summary, tags, success/failure responses), so later tasks have a real
  pattern to copy.

No feature endpoints exist yet — this task is purely the plumbing. Every task from here
on that adds an endpoint should include swagger annotations on its handlers as part of
that task, not deferred to a cleanup pass later.
```

**Checkpoint:** `make swagger && make run` (or `docker compose up --build`), then open `http://localhost:<port>/swagger/index.html` in a browser — it loads, lists `/healthz`, and "Try it out" against `/healthz` returns the real 200 response.

---

### Coach auth: signup, login, JWT middleware

**Description:** The one login this app has. A coach can be created (however you decide to bootstrap the first one — signup endpoint or seed script) and logs in to get a JWT that gates every other endpoint.

**Prompt:**
```
Create the internal/coach module (handler, service, repository, model.go — model.go
can now absorb the Coach struct that was temporarily in internal/models/; move it,
don't duplicate it). Implement:

- POST /api/v1/auth/signup — creates a coach (first_name, last_name, phone, password),
  hashes the password with bcrypt (never store or log the plaintext), returns the
  created coach (never the password_hash).
- POST /api/v1/auth/login — verifies phone + password against the bcrypt hash, returns
  a signed JWT (claims: coach id, expiry from config) on success, 401 on mismatch —
  don't leak whether the phone or the password was wrong in the error message.
- JWT auth middleware that validates the token on protected routes, rejects with 401
  if missing/invalid/expired, and injects the authenticated coach's id into the Gin
  context for handlers to read.

Apply the middleware to the /api/v1 group from the router skeleton task, except the
/auth/* routes themselves. Since v1 is single-coach (per mvp-spec.md), signup doesn't
need an admin gate yet, but every future module's queries must still filter by the
authenticated coach's id — treat that as a hard rule from here on, not optional.

Write unit tests for the service layer (password hashing/verification logic, token
generation/validation) using the testify setup from the testing strategy task, and one
integration test covering signup → login → hitting a protected route with the
resulting token.

Add swagger annotations to both handlers (summary, request body shape, success/error
responses) per the pattern from the swagger setup task, and rerun `make swagger`.
```

**Checkpoint:** `curl -X POST .../auth/signup` with valid data returns 201 and no `password_hash` field in the response; login with correct credentials returns a JWT; login with wrong password returns 401 with a generic message; hitting any `/api/v1/*` route without a token (or with a garbage token) returns 401; with a valid token it passes through. In Swagger UI, both endpoints appear under an auth tag, and pasting a login-issued token into "Authorize" lets subsequent "Try it out" calls on protected routes succeed.

---

## Phase 1 — Core v1 features (backend API)

> Every endpoint below must scope its query by the authenticated coach's id from the JWT (coach auth task) — an athlete, plan, or movement belonging to another coach should 404, not 403 (don't leak existence).

### Athlete management (MVP feature 1: dashboard)

**Description:** Coach can list their athletes, search by name, and add a new athlete — the coach dashboard's backend.

**Prompt:**
```
Create the internal/athlete module. Implement:
- POST /api/v1/athletes — create an athlete (fields per docs/mvp-spec.md's Athlete
  entity), coach_id set from the authenticated coach, not from the request body.
- GET /api/v1/athletes — list the authenticated coach's athletes, with an optional
  ?q= query param that searches first_name/last_name (case-insensitive, partial match).
  Paginate if you judge the list could realistically grow large; otherwise a flat list
  is fine for v1 — use your judgment and note which you chose.

Validate required fields server-side (don't trust the frontend to enforce them) and
return 400 with a clear field-level error on invalid input, not a raw DB error.

Write service-layer unit tests (validation, search filtering logic) and an integration
test covering: create athlete → appears in list → search matches by partial name →
another coach's athletes never appear in this coach's list.

Add swagger annotations to both handlers, including the bearer security requirement
and the ?q= query param, and rerun `make swagger`.
```

**Checkpoint:** Create two athletes under coach A, one under coach B (via a second signup). `GET /athletes` as coach A returns only their 2. `GET /athletes?q=<partial first name>` returns the matching one only. Confirm coach B's athlete never leaks into coach A's results even with a matching search term. Both endpoints are usable end-to-end from Swagger UI once authorized with coach A's token.

---

### Athlete profile + measurements (MVP feature 2)

**Description:** View an athlete's base info plus their measurement history (the data the progress chart is built from), and log a new measurement.

**Prompt:**
```
Add to internal/athlete (profile read) and create internal/measurement (measurement
CRUD), scoped through athlete ownership (a measurement's athlete must belong to the
authenticated coach — check this via a join/lookup, don't trust an athlete_id in the
request blindly). Implement:
- GET /api/v1/athletes/:id — full athlete profile (base info from mvp-spec.md's Athlete
  entity).
- POST /api/v1/athletes/:id/measurements — add a measurement (date, weight, chest,
  waist, arm, thigh, hip per docs/mvp-spec.md's AthleteMeasurement entity).
- GET /api/v1/athletes/:id/measurements — list measurements for the athlete, ordered
  by date, in a shape ready for a frontend chart to consume directly (an array of
  {date, weight, chest, ...} is fine — don't over-engineer a chart-specific format
  since the frontend isn't chosen yet).

Return 404 (not 403) if :id doesn't belong to the authenticated coach.

Write an integration test: create athlete, add 3 measurements across different dates,
confirm the list returns them ordered by date; confirm a measurement POST against
another coach's athlete id returns 404.

Add swagger annotations to all three handlers and rerun `make swagger`.
```

**Checkpoint:** Add measurements out of chronological order (e.g. POST a later date first, then an earlier one) and confirm `GET .../measurements` still returns them sorted by date, not insertion order. Confirm cross-coach 404 behavior manually with curl. Confirm all three endpoints are listed and usable from Swagger UI.

---

### Athlete's plan list (MVP feature 3)

**Description:** List an athlete's plans, with the current one distinguishable from past ones.

**Prompt:**
```
Create the internal/plan module (model.go covers Plan, Day, Block, BlockMovement per
docs/er-diagram.md — they're one module since they're each other's children, not
independent domains). Implement:
- GET /api/v1/athletes/:id/plans — list plans for the athlete, ordered by start_date
  descending. Decide and document how "current plan" is determined (e.g. most recent
  start_date, or most recent with no later plan superseding it) — docs/mvp-spec.md
  doesn't define this precisely, so make a reasonable call, note it in a code comment,
  and flag it to me as an assumption worth confirming.
- Each plan in the list response includes an is_current boolean computed per that rule,
  rather than making the frontend re-derive it.

Athlete ownership check (404 on mismatch) applies here too, same as the measurements task.

Write an integration test with 3 plans at different start_dates confirming exactly one
is marked is_current and it's the expected one per your chosen rule.

Add a swagger annotation to the handler and rerun `make swagger`.
```

**Checkpoint:** Create 3 plans for one athlete with different `start_date`s (including one with a future date, if your rule needs to handle that) and confirm the list response marks the right one `is_current` and the rest `false`. Confirm the endpoint is listed and usable from Swagger UI.

---

### Plan detail (MVP feature 4)

**Description:** Fetch one plan's full nested structure — days, each day's blocks, each block's movements — in one call, ready for a tabbed-by-day UI.

**Prompt:**
```
Add GET /api/v1/plans/:id to internal/plan, returning the plan with its full nested
tree: days (ordered by order_index), each day's blocks (ordered by order_index, with
sets/rest_seconds/notes), each block's movements (ordered by order_in_block, joined
with the movement library for name/category so the frontend doesn't need a second
round-trip per movement). Structure the JSON so blocks with multiple movements are
grouped together (per mvp-spec.md: "movements within a block are grouped visually...
supersets read as one unit") — a block's "movements" array naturally gives you that
grouping, confirm the response shape makes a single-movement block and a
multi-movement (superset) block look structurally identical, just with array length
1 vs N.

Ownership check: the plan's athlete must belong to the authenticated coach, 404 otherwise.
Use GORM preloading (Preload chains) to avoid an N+1 query per day/block/movement —
confirm this with the query log from the GORM setup task, not just by assumption.

Write an integration test building a plan with 2 days, each with 2 blocks, one block
having a 2-movement superset, confirming the nested response shape and ordering.

Add a swagger annotation to the handler, including the nested response schema, and
rerun `make swagger`.
```

**Checkpoint:** Enable GORM's SQL logging (dev log level) and hit this endpoint for a plan with several days/blocks/movements — count the queries. It should be a small constant number (Preload-driven), not one query per row. Confirm the JSON's day → block → movement nesting and ordering matches what you created. Confirm the endpoint is listed and usable from Swagger UI.

---

### Plan builder (MVP feature 5)

**Description:** The write side of plan-building — add a day to a plan, add a block to a day, add movement(s) to a block — including searching the movement library while building.

**Prompt:**
```
Extend internal/plan with the write endpoints:
- POST /api/v1/plans — create a plan for an athlete (start_date, title/note).
- POST /api/v1/plans/:id/days — add a day (label, order_index — validate label against
  the allowed set from mvp-spec.md: A/B/C/D/E/F or day1..day6, reject anything else
  with 400).
- POST /api/v1/days/:id/blocks — add a block (order_index, sets, rest_seconds, notes).
- POST /api/v1/blocks/:id/movements — add one or more movements to a block in one call
  (accept an array so a superset can be added atomically, each with movement_id, reps,
  duration_seconds, order_in_block) — validate every movement_id exists and is visible
  to this coach (their own or a universal one, per movement.coach_id rules in
  docs/er-diagram.md) before inserting any, and do the insert in a DB transaction so a
  partial superset never gets created on a mid-batch failure.

Every nested create must verify the ownership chain up to the authenticated coach
(block → day → plan → athlete → coach), 404 on any break in that chain.

Write integration tests: full build-up of a plan (create plan → add day → add block →
add a 2-movement superset in one call) confirming each step's response and the final
state via the plan detail endpoint; and a failure case (movement_id that doesn't belong
to this coach and isn't universal) confirming 400/404 and that nothing partial was
inserted.

Add swagger annotations to all four handlers and rerun `make swagger`.
```

**Checkpoint:** Build a full plan via curl/Postman end to end (plan → day → block → superset of 2 movements), then GET it via the plan detail endpoint and confirm it matches. Then deliberately POST a movement batch where the 2nd movement_id is invalid — confirm the response is an error and neither movement was inserted (check via a direct query, not just the error response). Repeat the same full build-up flow once from Swagger UI alone (Authorize, then "Try it out" on each endpoint in order) to confirm it's usable without curl.

---

### Movement library management (MVP feature 6)

**Description:** The coach's own movement library — list, search, add, edit, delete — with system-seeded universal movements visible but read-only.

**Prompt:**
```
Create the internal/movement module. Implement:
- GET /api/v1/movements?q= — list movements visible to the coach: their own
  (coach_id = authenticated coach) plus all universal ones (coach_id IS NULL), optional
  name search.
- POST /api/v1/movements — create a movement owned by the authenticated coach
  (coach_id set server-side, never from the request).
- PUT /api/v1/movements/:id — edit a movement; return 403 if it's universal
  (coach_id IS NULL) or owned by a different coach — a coach can only edit their own.
- DELETE /api/v1/movements/:id — same ownership rule as edit. Consider whether a
  movement in use by an existing BlockMovement should block deletion or cascade —
  docs/mvp-spec.md doesn't say; make a call (I'd lean toward blocking deletion with a
  409 if it's referenced, to avoid silently breaking existing plans), note it in a
  comment, and flag it to me as an assumption.

Also add a seed migration (separate from the schema migration, per golang-migrate's
convention of one concern per migration) inserting a small starter set of universal
movements (coach_id NULL) — a handful across warmup/strength/cardio categories is
enough for v1, this isn't meant to be exhaustive.

Write integration tests: coach can edit/delete their own movement; coach gets 403
editing/deleting a universal one; coach gets 403 (or 404, pick consistently with the
other feature endpoints' pattern and justify if you diverge) editing another coach's
movement; universal movements appear in every coach's list.

Add swagger annotations to all four handlers and rerun `make swagger`.
```

**Checkpoint:** As a fresh coach with no movements of their own, `GET /movements` still returns the seeded universal set. Create a custom movement, confirm it appears too. Attempt to `PUT`/`DELETE` a universal movement's id — confirm 403. Attempt the same against another coach's custom movement — confirm the same guard applies. Confirm all four endpoints are listed and usable from Swagger UI — this is also a convenient spot to eyeball the whole API surface in one place and confirm every earlier task's endpoints are still present in the generated docs.

---

## Phase 2 — Wrap-up

### Frontend stack decision

**Description:** Not a build task — a checkpoint to actually make the frontend call (React SPA vs. Go templates + htmx, per prestart-roadmap.md's open question) now that the backend API shape from the feature tasks above is real and can inform the decision.

**Prompt:**
```
I need to decide the frontend stack for GoGym now that the backend API (the six MVP
feature endpoints) is built and I can see the real shape of the endpoints and response
payloads. Summarize the tradeoff between a React SPA and Go templates + htmx
specifically in light of:
- the mobile-first, dense, touch-heavy UI and measurement trend chart from
  docs/mvp-spec.md's design constraints
- the actual JSON shapes now returned by /api/v1 (especially the nested plan detail
  response)
- my stated goals in docs/prestart-roadmap.md (deepen Go vs. broaden into a standard
  two-sided stack)
Don't implement anything yet — just help me decide, then once I confirm, write the
frontend task list as a follow-up to this document.
```

**Checkpoint:** A decision is recorded (append it to this file's header note, replacing "not yet decided"), and a new `docs/mvp-tasks-frontend.md` (or an appended section here) exists before any frontend code is written.
