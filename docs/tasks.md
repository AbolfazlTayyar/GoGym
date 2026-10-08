# GoGym — MVP Task List v1

Stack locked in for this list: **Go + Gin + GORM + PostgreSQL**, migrations via `golang-migrate`, JWT + bcrypt auth, module-first / clean-architecture-per-module layout (`internal/<module>/{handler,service,repository,model}.go`) per [prestart-roadmap.md](archive/prestart-roadmap.md) step 2. API docs via **swaggo (Swagger UI)** for manual endpoint testing — see the *Swagger / OpenAPI docs setup* task. Frontend stack is **not yet decided** — this list covers scaffolding and backend API only; frontend tasks get written once that's chosen (see *Frontend stack decision* below).

Each task below has: what it is, a plain description, a ready-to-paste prompt for Claude Code, and a checkpoint to verify it's actually done before moving on. Work top to bottom — later tasks assume earlier ones are merged.

> **Amendments folded in:** tasks marked **[A1]**–**[A4]** exist because the goal shifted from an internal tool to something other coaches pay for. They are grouped under those labels because each is close to free while there is no production data and expensive afterwards — do them in the order they appear, not last. The reasoning behind them is in [product-direction.md](product-direction.md); the task text below is self-contained and is what you work from.

---

## Phase 0 — Foundation

### Initialize Go project structure & tooling

✅ **Done**

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
docs/archive/prestart-roadmap.md step 2 — internal/<module>/{handler,service,repository,model}.go
— but don't create any module folders yet, that starts with the coach auth module.

Don't add any dependencies yet beyond what cmd/api/main.go needs to compile and run.
```

**Checkpoint:** `go build ./...` succeeds, `make run` starts and exits cleanly (or serves nothing yet — that's fine), `git status` shows only the new scaffolding files.

---

### Config loader & environment setup

✅ **Done**

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

✅ **Done**

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

✅ **Done**

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
nullable movement.coach_id noted in the ER diagram and spec.md (NULL = system-seeded,
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

✅ **Done**

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

✅ **Done**

**Description:** Decide and wire up how this project actually gets tested — unit tests with no DB, integration tests against a real Postgres — before feature code starts piling up untested.

**Prompt:**
```
Set up the testing harness described in docs/archive/prestart-roadmap.md step 9:
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

✅ **Done**

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

✅ **Done**

**Description:** The actual HTTP server boots, with middleware (logging, recovery, CORS) wired in, and a `/healthz` endpoint to prove it's alive.

**Prompt:**
```
Wire cmd/api/main.go to start a Gin server using internal/config for the port and
internal/db for the DB connection. Add global middleware: gin.Recovery(), a request
logger (structured JSON logs to stdout per docs/archive/prestart-roadmap.md step 10 — use
Gin's logger or a structured logger like zerolog/zap, pick one and tell me which),
and CORS configured permissively for now (frontend origin is still undecided per
spec.md — note a TODO to lock this down once the frontend stack and its origin
are chosen).

Add GET /healthz that checks DB connectivity (a lightweight ping, e.g. sqlDB.PingContext)
and returns 200 with {"status":"ok"} or 503 with the failure reason if the DB is
unreachable. Group all future feature routes under /api/v1 even though no feature
routes exist yet.
```

**Checkpoint:** `docker compose up`, then `curl localhost:<port>/healthz` returns 200 `{"status":"ok"}`; stop the `db` container and re-curl — should return 503 within a couple seconds, not hang.

---

### Swagger / OpenAPI docs setup

✅ **Done**

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

### Schema amendments: coaching fields & soft delete **[A2]**

✅ **Done**

**Description:** A second migration filling the gaps in the v1 schema that block real coaching use — prescription load, per-set variation, movement metadata, athlete status — plus soft delete. These are near-free now and a backfill-plus-downtime once real athlete data exists. This task exists separately because the original schema migration and the GORM models task are already merged; it amends both.

**Prompt:**
```
Write a new golang-migrate migration (and matching down migration) amending the schema
from migrations/000001_create_core_schema and
update the GORM structs in internal/models/ to match in the same commit.

block_movement — the prescription is incomplete:
- Load. There is reps and duration_seconds but no weight. Coaches program load, so a
  plan without it isn't a plan.a numeric kg column is good, and a free-text `load` field may
 serve v1 better than a
  premature numeric model. This is a judgment call — make it explicitly, don't default.

movement — the library is too thin to be fast:
- A media/video URL column. Note this implies object storage (S3/MinIO/a local provider)
  later; adding the column now does NOT commit us to building the upload flow in v1, and
  this task should not build it.

Soft delete — add deleted_at to the coach-owned tables a
coach can delete from the UI — at minimum athlete and movement. Use GORM's
gorm.DeletedAt so the default scope excludes them, and confirm the movement library's
visibility query (own + universal) still behaves correctly with the soft-delete scope
applied.

Per CLAUDE.md, docs/er-diagram.md is maintained by hand and must be updated in the SAME
commit as this migration — every new column, type and nullability, plus a note on the
soft-delete convention.
```

**Checkpoint:** `make migrate-up` applies cleanly on top of the existing schema and `make migrate-down` reverses it without dropping the base tables; rerun both twice to confirm idempotence. `.Find()` on every amended model via `internal/db` still succeeds with no "column not found" error. Soft-deleting an athlete removes it from `GET /athletes` results but the row is still present in a raw `SELECT` with `deleted_at IS NOT NULL`. `docs/er-diagram.md` diffs in the same commit as the migration.

---

### Coach auth: signup, login, JWT middleware **[A1]**

✅ **Done**

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
/auth/* routes themselves. Since v1 is single-coach (per spec.md), signup doesn't
need an admin gate yet.

Tenant isolation — build the enforcement layer here, before any feature module exists to
copy the wrong pattern:

- Every feature task below repeats "scope the query by the authenticated coach's id".
  A rule enforced by discipline across fifteen handlers is one forgotten WHERE away from
  a cross-tenant data leak. Make it structurally hard to bypass instead: a single scoping
  helper in the shared repository layer (a scoped *gorm.DB session or a GORM callback)
  that every tenant-owned query goes through, so constructing an unscoped query for a
  tenant-owned table is visibly wrong in review rather than merely against the rules.
- Keep the 404-not-403 convention the feature tasks already use — don't leak existence.
- Consider Postgres row-level security as defense in depth. I don't need it implemented,
  but record the decision in a comment either way, including if the answer is "not yet".
- Route the scope through one place so that if tenancy ever becomes org_id (a gym with
  several coaches, per docs/product-direction.md section 7) it's one file to change
  rather than every repository. Don't add an organization table now.

Also add rate limiting on the /auth/* routes (a simple in-process limiter keyed by IP
and phone is enough for v1 — no Redis). Unauthenticated endpoints that accept a password
or, after the next task, send an SMS, must not be free to hammer.

Write unit tests for the service layer (password hashing/verification logic, token
generation/validation) using the testify setup from the testing strategy task, one
integration test covering signup → login → hitting a protected route with the
resulting token, and one test proving the scoping helper actually filters — construct a
query through it as coach A and confirm coach B's rows are unreachable.

Add swagger annotations to both handlers (summary, request body shape, success/error
responses) per the pattern from the swagger setup task, and rerun `make swagger`.
```

**Checkpoint:** `curl -X POST .../auth/signup` with valid data returns 201 and no `password_hash` field in the response; login with correct credentials returns a JWT; login with wrong password returns 401 with a generic message; hitting any `/api/v1/*` route without a token (or with a garbage token) returns 401; with a valid token it passes through. Hammering `/auth/login` trips the rate limiter rather than accepting unlimited attempts. In Swagger UI, both endpoints appear under an auth tag, and pasting a login-issued token into "Authorize" lets subsequent "Try it out" calls on protected routes succeed.

---

### API response envelope

✅ **Done**

**Description:** One response shape for every `/api/v1` endpoint — `success`, `data`, `error`, `meta` — plus the helpers that produce it. Today each handler returns its own bare struct (`coachResponse`, `loginResponse`, `errorResponse`), so the shape is whatever each handler decided. This lands here, right after the only module that currently returns JSON, because retrofitting two handlers and a middleware is minutes of work while retrofitting fifteen endpoints across six feature modules — and the frontend already written against them — is not.

**Prompt:**
```
Introduce a single response envelope for the API and retrofit the existing handlers to it.

Shape — every /api/v1 response body, success or failure:

  {
    "success": true,
    "data":    { ... } | [ ... ] | null,
    "error":   null,
    "meta":    { ... }   // omitted entirely when empty
  }

  {
    "success": false,
    "data":    null,
    "error": {
      "code":    "validation_failed",
      "message": "invalid request",
      "fields":  { "phone": "must be an Iranian mobile number" }  // omitted when empty
    }
  }

- success is a real boolean field, not inferred from the status code. data and error are
  both always present on success/failure responses (one of them null).
- meta is reserved for pagination and similar list metadata. Nothing populates it in this
  task — define it, omitempty it, and leave it out of every current response. The athlete
  list task decides whether it paginates; this is the field it fills in when it does.
- error.code is a stable machine-readable string the frontend can branch on; error.message
  is human-readable. error.fields is an optional map of field name -> message, for
  request-validation failures only (the athlete task's "400 with a clear field-level error,
  not a raw DB error" requirement is what it exists for).
- Define the codes as named constants in one place, not as inline string literals per
  handler (per CLAUDE.md's rule on repeated literals). Start with only the codes the
  current handlers actually need — validation_failed, invalid_credentials, unauthorized,
  not_found, conflict, rate_limited, internal_error — and let later tasks add their own.
  Don't invent a code taxonomy for endpoints that don't exist yet.

Enforcement — explicit helpers, not response-rewriting middleware:

- Add a shared package (internal/httpx, or a name you prefer — tell me which and why) with
  the envelope types and helpers: OK / Created / NoContent-equivalent for success, and an
  error helper taking status + code + message, with a variant that carries the fields map.
- Handlers call the helpers and never call c.JSON / c.AbortWithStatusJSON directly.
  A response-rewriting middleware was considered and rejected: it buffers every response
  and desynchronizes the swagger annotations from the real body. Record that decision in a
  comment in the package.
- Gin generates a few responses that never reach a handler: no-route 404, no-method 405,
  and gin.Recovery()'s 500. Wire explicit NoRoute/NoMethod handlers through the helpers so
  those aren't the only unenveloped bodies on the API. For the recovery 500, decide whether
  a custom recovery handler is worth it in v1 — make the call, note it in a comment, flag
  it to me either way.

Retrofit — internal/coach is the only module returning JSON today:
- Signup, Login and Me responses (the coach/login payloads become the data value, unchanged
  in their own shape).
- The 400/401/409/429/500 error paths in the handlers, and the 401s in AuthMiddleware.
- The 429s from the /auth/* rate limiter.
- Fold the existing package-local errorResponse struct and its errMsg* constants into the
  shared package rather than leaving a second error shape behind.

Do NOT envelope /healthz. It's mounted outside /api/v1 on purpose as an unversioned infra
probe, and its {"status":"ok"} body is asserted by the Docker Compose healthcheck and by
earlier tasks' checkpoints in this file. Note the exemption in a comment on the handler so
it reads as deliberate rather than missed.

Swagger, per CLAUDE.md's typed-response rule — the annotations must show the real nested
body, not the bare payload struct:
- swaggo can compose a wrapper with a payload (@Success 200 {object} httpx.Envelope{data=coachResponse}).
  Verify it actually generates the nested schema with the installed swag version before
  committing to it; if it doesn't, fall back to explicit per-payload envelope structs rather
  than shipping annotations that lie about the response.
- Rerun `make swagger` and confirm in Swagger UI that the example bodies show the envelope.

Tests: unit tests for the helpers (each helper produces the documented shape and status,
fields omitted when empty, meta omitted when empty), and update the existing coach unit and
integration tests to assert against the enveloped bodies — a test still passing on a bare
body means something wasn't retrofitted.

Finally, write the convention down so the feature tasks below inherit it: add a short
"API response conventions" section to CLAUDE.md (the shape, the helpers-not-c.JSON rule, the
/healthz exemption) and a one-line note under this file's Phase 1 header alongside the
existing tenant-scoping note.
```

**Checkpoint:** `grep -rn "c.JSON\|AbortWithStatusJSON" internal/` returns hits only inside the envelope package and `internal/server/healthz.go`. `curl -X POST .../auth/signup` with valid data returns `{"success":true,"data":{...},"error":null}` with no `meta` key and still no `password_hash`; with a missing/short password it returns `success:false` and a populated `error.fields`; login with a wrong password returns the `invalid_credentials` code; hitting a protected route with no token returns the enveloped 401; hammering `/auth/login` returns the enveloped 429. `curl .../api/v1/does-not-exist` returns an enveloped 404, not Gin's default `404 page not found` text. `curl .../healthz` still returns bare `{"status":"ok"}` and `docker compose ps` still reports the api container healthy. In Swagger UI the signup/login/me example responses show the envelope with the real payload nested under `data`, not a free-form object.

---

### Coach auth hardening: phone verification & password reset **[A4]**

⬜ **Not started**

**Description:** The two flows missing from the auth task that turn into support tickets the first day someone who isn't you uses this. Split from the task above because OTP needs an SMS provider decision and its own schema, not because it's optional — both should land before anyone else signs up.

**Prompt:**
```
Extend internal/coach with the two auth flows missing from the signup/login task.

1. Phone verification (OTP). phone is both the login identity and a unique key, so an
   unverified typo at signup creates an account nobody can log into, nobody can recover,
   and whose uniqueness constraint now squats on the real number. Add a one-time-code
   flow: issue a short numeric code on signup, verify it against a stored hash with a
   short expiry and an attempt limit, and mark the coach verified.
   - Pick the SMS provider deliberately and tell me the options before wiring one in —
     this is market-specific (see the market assumption in docs/product-direction.md
     section 5) and I may already have a preference.
   - Put it behind an interface with a log-to-stdout implementation for local dev and
     tests, so the suite never needs a real provider or network.
   - Decide and tell me whether an unverified coach can log in at all or is only blocked
     from some actions. Make the call, note it in a comment, flag it as an assumption.

2. Password reset. There is no flow at all today — the only recovery path is a manual
   UPDATE against the database. Build reset-request → code/token → set-new-password on
   the same OTP machinery, with single-use tokens, a short expiry, and no response
   difference between a known and an unknown phone number (don't leak which numbers are
   registered).

Both flows go through the /auth/* rate limiter from the previous task — verify that's
actually applied, since these are the endpoints that cost real money per request when an
SMS provider is wired in.

Add swagger annotations to every new handler and rerun `make swagger`. If this needs new
columns or a table (verification codes, reset tokens), it's a migration, and per CLAUDE.md
docs/er-diagram.md updates in the same commit.
```

**Checkpoint:** Sign up with a phone number, confirm the code is logged by the dev SMS implementation, and verify it — a wrong code and an expired code both fail without verifying, and repeated wrong codes hit the attempt limit. Request a password reset for a registered number and complete it, then confirm the old password no longer works and the reset token can't be reused. Request a reset for an unregistered number and confirm the response is indistinguishable from the registered case. The full suite still passes with no network access.

**Expected output:**
```
$ docker compose logs api | grep otp
{"level":"info","msg":"dev sms: otp sent","phone":"09372144430","code":"483920"}

$ curl -s -X POST localhost:8080/api/v1/auth/verify -d '{"phone":"09372144430","code":"000000"}'
{"success":false,"data":null,"error":{"code":"invalid_code","message":"invalid or expired code"}}

$ curl -s -X POST localhost:8080/api/v1/auth/verify -d '{"phone":"09372144430","code":"483920"}'
{"success":true,"data":{"verified":true},"error":null}

$ curl -s -X POST localhost:8080/api/v1/auth/password-reset/request -d '{"phone":"09120000000"}'
{"success":true,"data":null,"error":null}     # identical for a registered number
```

**In short:** a typo'd phone can no longer lock an account forever, and a coach who forgets their password can get back in without you touching the database.

---

## Phase 1 — Core v1 features (backend API)

> Every endpoint below returns the `internal/httpx` response envelope — `{success, data, error, meta}` — written through that package's helpers, never `c.JSON` directly. See the *API response envelope* task and CLAUDE.md's *API response conventions*.

> Every endpoint below is tenant-scoped through the scoping helper from the coach auth task **[A1]** — don't hand-write a `WHERE coach_id = ?` per handler and don't rely on remembering to. An athlete, plan, or movement belonging to another coach should 404, not 403 (don't leak existence).

### Athlete management (MVP feature 1: dashboard)

✅ **Done**

**Description:** Coach can list their athletes, search by name, and add a new athlete — the coach dashboard's backend.

**Prompt:**
```
Create the internal/athlete module. Implement:
- POST /api/v1/athletes — create an athlete (fields per docs/spec.md's Athlete
  entity), coach_id set from the authenticated coach, not from the request body.
- GET /api/v1/athletes — list the authenticated coach's athletes, with an optional
  ?q= query param that searches first_name/last_name (case-insensitive, partial match).
  Paginate if you judge the list could realistically grow large; otherwise a flat list
  is fine for v1 — use your judgment and note which you chose.
- The home screen/dashboard view of this list should default to `athlete_type = 'private'`
  athletes only (see docs/er-diagram.md) — `public` athletes are one-off plan-link
  deliveries the coach doesn't manage ongoing, so they don't belong in the main working
  list. Support an explicit filter/param to see `public` athletes too rather than hiding
  them entirely.

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

✅ **Done**

**Description:** View an athlete's base info plus their measurement history (the data the progress chart is built from), and log a new measurement.

**Prompt:**
```
Add to internal/athlete (profile read) and create internal/measurement (measurement
CRUD), scoped through athlete ownership (a measurement's athlete must belong to the
authenticated coach — check this via a join/lookup, don't trust an athlete_id in the
request blindly). Implement:
- GET /api/v1/athletes/:id — full athlete profile (base info from spec.md's Athlete
  entity).
- POST /api/v1/athletes/:id/measurements — add a measurement (date, weight, chest,
  waist, arm, thigh, hip per docs/spec.md's AthleteMeasurement entity).
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

**Expected output:**
```
$ curl -s -H "Authorization: Bearer $TOKEN" localhost:8080/api/v1/athletes/$ID/measurements
{"success":true,"data":[
  {"id":"...","date":"2026-08-01","weight":82.5,"chest":104,"waist":90,"arm":36,"thigh":58,"hip":100},
  {"id":"...","date":"2026-09-01","weight":80.1,"chest":103,"waist":87,"arm":36,"thigh":57,"hip":98}
],"error":null}

$ curl -s -X POST -H "Authorization: Bearer $TOKEN_B" localhost:8080/api/v1/athletes/$ID/measurements -d '{...}'
{"success":false,"data":null,"error":{"code":"not_found","message":"athlete not found"}}
```

**In short:** you can open one athlete, see their details, and log body measurements over time — the data the progress chart will draw from.

---

### Athlete's plan list (MVP feature 3)

✅ **Done**

**Description:** List an athlete's plans, with the current one distinguishable from past ones.

**Prompt:**
```
Create the internal/plan module (model.go covers Plan, Day, Block, BlockMovement per
docs/er-diagram.md — they're one module since they're each other's children, not
independent domains). Implement:
- GET /api/v1/athletes/:id/plans — list plans for the athlete, ordered by start_date
  descending. Decide and document how "current plan" is determined (e.g. most recent
  start_date, or most recent with no later plan superseding it) — docs/spec.md
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

**Expected output:**
```
$ curl -s -H "Authorization: Bearer $TOKEN" localhost:8080/api/v1/athletes/$ID/plans
{"success":true,"data":[
  {"id":"...","title":"Cut phase 2","start_date":"2026-11-01","is_current":false},
  {"id":"...","title":"Cut phase 1","start_date":"2026-09-15","is_current":true},
  {"id":"...","title":"Base","start_date":"2026-07-01","is_current":false}
],"error":null}
```

**In short:** you can see every plan an athlete has had, newest first, with the one they're on right now clearly marked.

---

### Plan detail (MVP feature 4)

✅ **Done**

**Description:** Fetch one plan's full nested structure — days, each day's blocks, each block's movements — in one call, ready for a tabbed-by-day UI.

**Prompt:**
```
Add GET /api/v1/plans/:id to internal/plan, returning the plan with its full nested
tree: days (ordered by order_index), each day's blocks (ordered by order_index, with
sets/rest_seconds/notes), each block's movements (ordered by order_in_block, joined
with the movement library for name/category so the frontend doesn't need a second
round-trip per movement). Structure the JSON so blocks with multiple movements are
grouped together (per spec.md: "movements within a block are grouped visually...
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

**Expected output:**
```
$ curl -s -H "Authorization: Bearer $TOKEN" localhost:8080/api/v1/plans/$PLAN_ID
{"success":true,"data":{
  "id":"...","title":"Cut phase 1","start_date":"2026-09-15",
  "days":[
    {"label":"A","order_index":0,"blocks":[
      {"order_index":0,"sets":4,"rest_seconds":90,"notes":null,"movements":[
        {"movement_id":"...","name":"Back squat","category":"strength","reps":8,"order_in_block":0}
      ]},
      {"order_index":1,"sets":3,"rest_seconds":60,"notes":"superset","movements":[
        {"movement_id":"...","name":"Pull-up","category":"strength","reps":10,"order_in_block":0},
        {"movement_id":"...","name":"Push-up","category":"strength","reps":15,"order_in_block":1}
      ]}
    ]}
  ]
},"error":null}

# GORM SQL log for that request: a handful of SELECTs, not one per day/block/movement
```

**In short:** one request returns a whole plan — days, blocks, supersets, movement names — ready for the screen to show day by day.

---

### Plan builder (MVP feature 5) **[A3]**

✅ **Done**

**Description:** The write side of plan-building — add a day to a plan, add a block to a day, add movement(s) to a block — including searching the movement library while building.

**Prompt:**
```
Extend internal/plan with the write endpoints:
- POST /api/v1/plans — create a plan for an athlete (start_date, title/note).
- POST /api/v1/plans/:id/days — add a day (label, order_index — validate label against
  the allowed set from spec.md: A/B/C/D/E/F/G or day1..day7, reject anything else
  with 400). A plan holds at most plan.MaxDaysPerPlan (7) days, enforced by the day
  table's constraints (docs/adr/0017): map gorm.ErrCheckConstraintViolated (order_index
  outside 0-6) and gorm.ErrDuplicatedKey (slot already taken) to 400 validation errors
  on order_index, not 500s.
- POST /api/v1/days/:id/blocks — add a block (order_index, sets, rest_seconds, notes).
- POST /api/v1/blocks/:id/movements — add one or more movements to a block in one call
  (accept an array so a superset can be added atomically, each with movement_id, reps,
  duration_seconds, order_in_block, plus the load/per-set/tempo fields added by the
  schema amendments task) — validate every movement_id exists and is visible to this
  coach (their own or a universal one, per movement.coach_id rules in
  docs/er-diagram.md) before inserting any, and do the insert in a DB transaction so a
  partial superset never gets created on a mid-batch failure.

Every nested create must verify the ownership chain up to the authenticated coach
(block → day → plan → athlete → coach), 404 on any break in that chain — through the
scoping helper from the coach auth task, not a hand-written filter.

Before implementing, record one design decision in a comment on this module. This task
deliberately ships NO logging table — the decision is what matters now:

  docs/er-diagram.md notes that Day, Block and BlockMovement carry created_at only,
  because the plan builder deletes and recreates them rather than editing in place.
  That's a fine v1 simplification, but the moment athlete workout logs exist, it turns
  destructive: a log row that foreign-keys to block_movement_id is orphaned by every
  plan edit, and training history is the hardest-to-replace data in the product. So
  when logging is built, a log must reference movement_id plus a denormalized snapshot
  of what was prescribed at the time (sets/reps/load as programmed), NOT the plan
  structure. Write that down here so whoever implements logging doesn't rediscover it.

Note in the same comment that plan templates and duplication (docs/product-direction.md
section 3) also assume plans are copied and edited in place rather than rebuilt — if that
feature is likely, delete-and-recreate is on borrowed time and updated_at on those three
tables is cheap insurance. Flag it to me rather than changing the approach in this task.

Write integration tests: full build-up of a plan (create plan → add day → add block →
add a 2-movement superset in one call) confirming each step's response and the final
state via the plan detail endpoint; and a failure case (movement_id that doesn't belong
to this coach and isn't universal) confirming 400/404 and that nothing partial was
inserted.

Add swagger annotations to all four handlers and rerun `make swagger`.
```

**Checkpoint:** Build a full plan via curl/Postman end to end (plan → day → block → superset of 2 movements), then GET it via the plan detail endpoint and confirm it matches. Then deliberately POST a movement batch where the 2nd movement_id is invalid — confirm the response is an error and neither movement was inserted (check via a direct query, not just the error response). Repeat the same full build-up flow once from Swagger UI alone (Authorize, then "Try it out" on each endpoint in order) to confirm it's usable without curl.

**Expected output:**
```
$ curl -s -X POST -H "Authorization: Bearer $TOKEN" localhost:8080/api/v1/blocks/$BLOCK_ID/movements \
    -d '[{"movement_id":"<squat>","reps":8,"order_in_block":0},{"movement_id":"<lunge>","reps":10,"order_in_block":1}]'
{"success":true,"data":[{...},{...}],"error":null}

$ curl -s -X POST -H "Authorization: Bearer $TOKEN" localhost:8080/api/v1/plans/$PLAN_ID/days -d '{"label":"Z","order_index":0}'
{"success":false,"data":null,"error":{"code":"validation_failed","message":"invalid request","fields":{"label":"must be one of A-G or day1-day7"}}}

$ curl -s -X POST ... /blocks/$BLOCK_ID/movements -d '[{"movement_id":"<valid>",...},{"movement_id":"<other coach>",...}]'
{"success":false,"data":null,"error":{"code":"validation_failed","message":"invalid request","fields":{"[1].movement_id":"not found"}}}
# and SELECT count(*) FROM block_movement WHERE block_id = '<BLOCK_ID>' is unchanged
# (a 400 naming the bad entry, not a 404 — see docs/adr/0019)
```

**In short:** you can build a full training plan from scratch through the API — days, blocks, supersets — and a bad request never leaves half a plan behind.

---

### Movement library management (MVP feature 6)

✅ **Done**

**Description:** The coach's own movement library — list, search, add, edit, delete — with system-seeded universal movements visible but read-only.

**Prompt:**
```
Create the internal/movement module. Implement:
- GET /api/v1/movements?q= — list movements visible to the coach: their own
  (coach_id = authenticated coach) plus all universal ones (coach_id IS NULL), optional
  name search, plus filters on the muscle group and equipment columns added by the
  schema amendments task — those filters are what make picking a movement fast while
  building a plan, which is the point of the whole builder.
- POST /api/v1/movements — create a movement owned by the authenticated coach
  (coach_id set server-side, never from the request).
- PUT /api/v1/movements/:id — edit a movement; return 403 if it's universal
  (coach_id IS NULL) or owned by a different coach — a coach can only edit their own.
- DELETE /api/v1/movements/:id — same ownership rule as edit. Consider whether a
  movement in use by an existing BlockMovement should block deletion or cascade —
  docs/spec.md doesn't say; make a call (I'd lean toward blocking deletion with a
  409 if it's referenced, to avoid silently breaking existing plans), note it in a
  comment, and flag it to me as an assumption.

Also add a seed migration (separate from the schema migration, per golang-migrate's
convention of one concern per migration) inserting a small starter set of universal
movements (coach_id NULL) — a handful across warmup/strength/cardio categories is
enough for v1, this isn't meant to be exhaustive. Populate the muscle group and
equipment columns on the seeded rows; a seeded library with those left null makes the
new filters look broken on a fresh account, which is exactly the first impression a
new coach gets. Leave the media/video column null — the upload flow isn't in v1.

Write integration tests: coach can edit/delete their own movement; coach gets 403
editing/deleting a universal one; coach gets 403 (or 404, pick consistently with the
other feature endpoints' pattern and justify if you diverge) editing another coach's
movement; universal movements appear in every coach's list.

Add swagger annotations to all four handlers and rerun `make swagger`.
```

**Checkpoint:** As a fresh coach with no movements of their own, `GET /movements` still returns the seeded universal set. Create a custom movement, confirm it appears too. Attempt to `PUT`/`DELETE` a universal movement's id — confirm 403. Attempt the same against another coach's custom movement — confirm it's refused too, as a 404 rather than a 403 (see docs/adr/0021). Confirm all four endpoints are listed and usable from Swagger UI — this is also a convenient spot to eyeball the whole API surface in one place and confirm every earlier task's endpoints are still present in the generated docs.

**Expected output:**
```
$ curl -s -H "Authorization: Bearer $NEW_COACH_TOKEN" "localhost:8080/api/v1/movements?muscle_group=legs"
{"success":true,"data":[
  {"id":"...","coach_id":null,"name":"Back squat","category":"strength","description":null,"muscle_group":"legs","equipment":"barbell","media_url":null,...},
  {"id":"...","coach_id":null,"name":"Bodyweight squat","category":"warmup","description":null,"muscle_group":"legs","equipment":"bodyweight","media_url":null,...},
  ...
],"error":null}
# 6 seeded movements, ordered by name

$ curl -s -X PUT -H "Authorization: Bearer $TOKEN" localhost:8080/api/v1/movements/$UNIVERSAL_ID -d '{"name":"Squat"}'
{"success":false,"data":null,"error":{"code":"forbidden","message":"universal movements can't be edited or deleted"}}

$ curl -s -X PUT -H "Authorization: Bearer $TOKEN" localhost:8080/api/v1/movements/$OTHER_COACH_ID -d '{"name":"Squat"}'
{"success":false,"data":null,"error":{"code":"not_found","message":"not found"}}
# a 404, not a 403, so another coach's movement isn't confirmed to exist — see docs/adr/0021

$ curl -s -X DELETE -H "Authorization: Bearer $TOKEN" localhost:8080/api/v1/movements/$IN_USE_ID
{"success":true,"data":null,"error":null}
# a soft delete, not a 409: the plan still names the movement, it just leaves the library — see docs/adr/0022
```

**In short:** every coach starts with a ready-made exercise library they can search and filter, and can add, edit and remove their own exercises without touching the shared ones.

---

## Phase 2 — Wrap-up

### VPS deployment & production migrations

⬜ **Not started**

**Description:** Nothing runs anywhere but local dev and CI today — there is no deploy target and no defined way to apply a migration against a real production database. This task exists because the *Backup & restore drill* task below assumes a running deployment target and a migration process already exist; right now neither does.

**Prompt:**
```
Stand up the v1 deployment target per docs/archive/prestart-roadmap.md step 7: a VPS
(Hetzner/DigitalOcean or similar) running the existing docker-compose.yml (api + db),
behind a reverse proxy (Caddy is the simplest TLS option) terminating HTTPS. Don't reach
for Kubernetes.

Define and document how migrations reach production — this app has no auto-migrate on
container startup today (internal/db explicitly disables GORM AutoMigrate; migrations/
is the only source of truth, applied via `make migrate-up` wrapping golang-migrate CLI).
Pick one deliberately and tell me why:
- A manual step: SSH to the VPS, run `make migrate-up` (or the raw `migrate` binary)
  against the prod DSN before restarting the api container on each deploy.
- An automated step: a one-shot init container/job in the compose stack that runs
  `migrate ... up` and exits before the api service starts (compose depends_on +
  a healthcheck-style gate, or a deploy script step) — safer against "forgot to migrate"
  but needs care that it never runs concurrently with a second deploy.

Whichever you pick, the DSN must come from the same internal/config-driven env vars used
locally — no separate prod-only connection logic. Add a short runbook (a doc or a section
in this file) covering: how to deploy a new image, how/when migrations run relative to
that, and how to roll back a bad migration (down migration + previous image tag).

Per docs/archive/prestart-roadmap.md step 10, this is also the natural point to wire basic
CD (a GitHub Actions job that builds and pushes the image, then SSHs to the VPS to pull
and restart — no need for anything fancier yet) — call out explicitly if you're deferring
that to a separate task instead of doing it here.
```

**Checkpoint:** A real migration (start with the existing `000001`/`000002` files) has actually been applied against the VPS's Postgres, not just described — confirm via `\dt`/`\d athlete` over SSH. Deploying a trivial code change (e.g. a log line) end-to-end once, following only the written runbook, succeeds without undocumented manual steps. Rolling back one migration via the down file has been exercised at least once, not just assumed to work.

**Expected output:**
```
$ curl -s https://<your-domain>/healthz
{"status":"ok"}

$ ssh vps 'docker compose exec db psql -U gogym -c "\dt"'
 public | athlete              | table | gogym
 public | block                | table | gogym
 ...
 public | schema_migrations    | table | gogym

$ ssh vps 'docker compose exec db psql -U gogym -c "SELECT version, dirty FROM schema_migrations"'
 version | dirty
---------+-------
       2 | f
```

**In short:** the app runs on a real server over HTTPS, and there's a written, tested way to ship a new version and its migrations — and to undo one.

---

### Backup & restore drill **[A4]**

⬜ **Not started**

**Description:** Not a feature — the point at which losing the database stops being an inconvenience and starts being the end of the business. An untested backup is not a backup.

**Prompt:**
```
Set up database backups for the deployment target from docs/archive/prestart-roadmap.md step 7
(VPS + Docker Compose): a scheduled pg_dump of the Postgres volume, retained for a
sensible window, stored somewhere that is NOT the same VPS — a lost disk shouldn't take
the backups with it. Keep it as simple as the rest of the stack; a cron job and an
object-storage bucket is enough, this doesn't need a backup tool.

Then actually perform a restore drill and write down what happened: take a dump, restore
it into a throwaway Postgres (locally or a second container), and confirm the data is
really there. Document the restore procedure as a short runbook — the steps, in order,
that you'd follow at 2am with a dead database. Note how long the restore took.

Don't claim this task is done on the strength of the backup job running. It's done when
a restore has succeeded and the runbook has been followed once, start to finish.
```

**Checkpoint:** A backup exists off the VPS. A restore from that backup into an empty Postgres has actually been run by you — not described — and the restored database serves a working `/healthz` and returns real athlete rows. The runbook exists and someone who isn't you could follow it.

**Expected output:**
```
$ <list the off-VPS bucket>
gogym-2026-10-02T03-00.sql.gz
gogym-2026-10-03T03-00.sql.gz

$ gunzip -c gogym-2026-10-03T03-00.sql.gz | docker compose -f restore.yml exec -T db psql -U gogym
...
$ curl -s localhost:8081/healthz
{"status":"ok"}
$ docker compose -f restore.yml exec db psql -U gogym -c "SELECT count(*) FROM athlete"
 count
-------
    37
```

**In short:** the database is copied off the server every day, and you've actually brought it back from a copy once, so you know it works.

---

### Frontend stack decision

⬜ **Not started**

**Description:** Not a build task — a checkpoint to actually make the frontend call (React SPA vs. Go templates + htmx, per prestart-roadmap.md's open question) now that the backend API shape from the feature tasks above is real and can inform the decision.

**Prompt:**
```
I need to decide the frontend stack for GoGym now that the backend API (the six MVP
feature endpoints) is built and I can see the real shape of the endpoints and response
payloads. Summarize the tradeoff between a React SPA and Go templates + htmx
specifically in light of:
- the mobile-first, dense, touch-heavy UI and measurement trend chart from
  docs/spec.md's design constraints
- the actual JSON shapes now returned by /api/v1 (especially the nested plan detail
  response)
- my stated goals in docs/archive/prestart-roadmap.md (deepen Go vs. broaden into a standard
  two-sided stack)
Don't implement anything yet — just help me decide, then once I confirm, write the
frontend task list as a follow-up to this document.
```

**Checkpoint:** A decision is recorded (append it to this file's header note, replacing "not yet decided"), and a new `docs/tasks-frontend.md` (or an appended section here) exists before any frontend code is written.

**Expected output:**
```
docs/tasks.md header:  "Frontend stack is **<React SPA | Go templates + htmx>** — chosen because ..."
docs/tasks-frontend.md exists, in this file's task format, with no frontend code written yet
```

**In short:** the frontend choice is made and written down, and there's a task list for building the screens.
