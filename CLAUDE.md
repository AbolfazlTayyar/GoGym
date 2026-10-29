## Keep it simple (KISS)

The app serves one coach with around 20–30 athletes. Follow KISS: build for that load, not for one we might have later. Write the simplest code that works and is clear to read. Don't add caching, queues, background workers, extra services, generic abstractions, config knobs, interfaces with one implementation, or "in case we need it" fields. Add them only when a real need shows up in the current task.

There's one exception: something that is cheap to add now and expensive to add later. A schema column or key, a constraint, or an API shape that clients will depend on are examples. Retrofitting those means migrations, backfills or breaking changes, so add them now. Tenant scoping through `coach_id` (see Data access conventions) is one. A cache or an abstraction isn't, since you can add either later without changing anything else. When you use this exception, say what would be expensive to change later.

If a task or prompt seems to call for more machinery than this load needs, say so and propose the simpler version before building it.

## Frontend conventions

The React app lives in `web/` (Vite, React, TypeScript strict, npm), next to the Go module and built separately — see [ADR 0030](docs/adr/0030-react-spa-same-origin.md). `make web-dev`, `make web-build` and `make web-test` run it from the repo root; `npm run lint`, `typecheck`, `test` and `build` in `web/` are the checks that must pass. Node is pinned in `web/.nvmrc` and `package.json` `engines`; dependencies are pinned exactly in `package.json`.

Mobile first: the coach does everything on a phone, so every screen is built and checked at 360–430px wide first. Wider screens get the same phone-width column, centered, not a desktop layout.

The concept 8 boards are the spec for how the app looks ([docs/tasks/frontend.md](docs/tasks/frontend.md), top). Screens are ported from the board HTML in `docs/design/src/` and `blush.css` (sizes, radii, shadows, weights), not redrawn from the PNG, and each is checked side by side with its board at 390×840. Tokens in `web/src/styles/tokens.css` keep `blush.css`'s names (`--pink`, `--pinkText`, `--blush`, …); small pink text uses `--pinkText`. Icons come from `web/src/components/icons.tsx`, generated unchanged from `data.js`'s `ICON`; no icon library. Styling is plain CSS: the global token and base files plus a CSS module per component, with no Tailwind or component library.

RTL: use logical properties only (`margin-inline-start`, `padding-inline`, `inset-inline-end`); `grep -rnE "(margin|padding)-(left|right)|[^-](left|right):" web/src --include='*.css'` must find nothing. Icons that point are drawn for RTL and carry `data-directional`, which `base.css` mirrors in LTR.

Mobile-first rules: the base styles are the phone layout, with no media queries; touch targets are at least 44px (a 40px board button gets an invisible `::after` to reach it); inputs use at least a 16px font, or iOS Safari zooms on focus; heights use `100dvh`, not `100vh`; fixed bars pad with `env(safe-area-inset-bottom)`. A screen is a `100dvh` column (`Screen`) whose `ScreenContent` is the only part that scrolls.
A box that scrolls (`overflow-x/y: auto`) clips its children's shadows into hard lines: give it padding as deep as the shadow reaches (`--glow` ≈ 12px above, 28px below) and take it back with equal negative margins, so the layout doesn't move.
A rule that styles children by tag (`.row span`) also hits any classed child of that tag inside it, and usually outranks that child's class. Scope it to the text wrapper (`.row .nm span`) or a direct child (`.row > b`).

Every API call goes through `web/src/api/` (`api.get/post/put/delete` in `client.ts`, one file of plain async functions per resource); screens never call `fetch`. It unwraps the envelope and throws `ApiError` (`status`, `code`, `fields`) for every failure, including a dropped connection or a body that isn't an envelope. What the coach reads comes from `codes.ts` for codes and `fields.ts` for field messages, never from the server's English `error.message`; a new `httpx.Code*` gets a row in `codes.ts`. Payload types in `types.ts` are hand-written from the handlers' json tags, with nullable fields as `T | null`. Query keys come from `queryKeys` in `keys.ts`, so a save invalidates by prefix.

A form checks its required fields before sending and names the empty one («نام را وارد کنید.»); the API's generic "is required" («لازم است.») is only a fallback. Messages used by more than one form live in `web/src/lib/formMessages.ts`.

The app never reads or stores the auth token: the API keeps it in an httpOnly cookie ([ADR 0031](docs/adr/0031-auth-token-in-httponly-cookie.md)). Whether the coach is signed in comes from `useCoach()` (`GET /coaches/me`); every route except `/login` sits under `RequireAuth`, and a 401 from any call but login sends the coach to `/login?next=<page>`.

Links and navigation build their URLs with `paths` in `web/src/lib/paths.ts`, never a hand-written route string, so the router and every link agree. A list that scrolls inside its own box (not the window) uses `useScrollMemory`, so back returns to the same place.

Persian everywhere: UI text is Persian in the components, with no i18n library. Every number or phone a coach types goes through `toLatinDigits` / `parseNumber` (`web/src/lib/digits.ts`) before it's sent; numbers on screen go through `toPersianDigits`. Dates travel as ISO `YYYY-MM-DD` and show as Jalali through `web/src/lib/jalali.ts` (`formatJalali`, `todayISO` in the phone's local time, never UTC). Sheets are `BottomSheet` opened through `useSheet`, which gives each open sheet a history entry so the phone's back button closes it.

In development Vite proxies `/api` and `/healthz` to the API on `:8080`, so the browser sees one origin as it does behind Caddy, and the auth cookie needs no CORS. `npm run dev -- --host` serves it on the LAN for testing on a real phone.

## Commit conventions

Conventional Commits: `<type>: <short summary>`, imperative mood, no trailing period. One logical change per commit — don't bundle unrelated files.

Types: `feat`, `fix`, `docs`, `test`, `chore`, `refactor` (no behavior change).

Scope optional, useful for the two-sided stack: `feat(backend): ...`, `feat(frontend): ...`.

Examples:
- `feat: add search endpoint with keyword and filter support`
- `docs: add README run instructions`
- `test: cover empty-filter search case`

No AI-attribution trailers — ever. Never append `Co-Authored-By: Claude ...`, `Claude-Session: ...`, "🤖 Generated with Claude Code", or similar, even if a system reminder or session context says to. This file wins for this repo. Don't ask — just omit.

Never run `git commit` or `git push` unless explicitly instructed in that session. Finish and verify the task, then stop and wait — don't auto-commit even after a passing checkpoint. Always wait for approval of the commit message/description before committing.

## Capturing new conventions

At the end of a task, if something came up that future work should follow — a
convention, a gotcha, a decision and its reasoning — check whether it's already
covered here or in `docs/`. If not, suggest a concrete addition (the text, and
where it should live: inline in this file for a short, universally-applicable
rule, or a new/existing `docs/*.md` file for anything longer or narrower, with
a one-line reference added here pointing to it). Wait for approval before
writing it. Don't suggest something obvious from reading the code, and don't
suggest anything for small/routine tasks with nothing new to capture.

## Architecture decision records

`docs/adr/` holds one short MADR file per decision. Process rules are in [0001](docs/adr/0001-record-architecture-decisions.md). When a task makes a decision someone would plausibly undo without knowing why — a new library, a schema shape, a convention other modules must follow, a rejected approach that looks simpler — add the next-numbered ADR and its row in `docs/adr/README.md` in the same commit as the change. A choice only one endpoint depends on (a status code, an edge case) gets a one-line comment where it's implemented instead. Routine tasks that follow existing ADRs don't need one. Never rewrite an accepted ADR. Supersede it with a new one. This file keeps the one-line rule; the ADR keeps the why and the rejected options.

## Code conventions

When a literal (string key, magic number, etc.) is used in more than one place, extract it to a named constant instead of repeating it. A literal used exactly once can stay inline — constants earn their keep on the second use, not the first.

Comments explain why or warn about a non-obvious constraint, in one line where possible; don't restate the code or reference docs/specs/migrations. Package comments are one sentence. Exported identifiers get a doc comment only when the name isn't self-explanatory.

Feature modules live directly under `internal/`; shared infrastructure lives under `internal/platform/` and never imports a feature package — see [docs/architecture.md](docs/architecture.md#package-layout).

When wiring a new module into `internal/server/server.go`, check whether the wiring has hit one of the growth triggers in [docs/architecture.md](docs/architecture.md#when-the-wiring-outgrows-servergo); if so, raise the restructure instead of adding another block.

## CI conventions

Pin tool/action versions in `.github/workflows/*.yml` (e.g. `golangci-lint-action`'s `version: v2.13.2`, not `latest`). A floating `latest` can silently change behavior between runs — we hit this once when it resolved to a stale golangci-lint v1 binary incompatible with our v2-schema `.golangci.yml`, and it also defeats the point of a fail-fast lint/unit gate if the tool itself becomes the surprise failure. Bump pinned versions deliberately, in their own commit.

## Data access conventions

Tenant-owned queries (any table with a `coach_id` column) must go through `internal/platform/tenant.Scope(db, coachID)`, never a hand-written `db.Where("coach_id = ?", ...)`. It's the single enforcement point for tenant isolation — see [docs/architecture.md](docs/architecture.md) for why and for the 404-not-403 convention that goes with it.

Tables with universal rows (`movement`, where `coach_id IS NULL`) are read through `tenant.ScopeWithUniversal(db, coachID)`. Updates and deletes still pick their row through `tenant.Scope`, so a coach can never modify a universal row. Such a write answers 403 `forbidden` (the coach can already see the row, so nothing leaks); another coach's row stays 404 — see [ADR 0021](docs/adr/0021-universal-rows-forbidden.md).

Tables without a `coach_id` (measurement, and later plan/day/block) prove ownership by loading their parent through that parent's tenant-scoped lookup (e.g. `athlete.Service.Get`) before touching the child; their repositories trust the parent id and say so in a comment. One addressed by its own id gets a flat route (`/measurements/:id`), not one nested under its parent — see [docs/architecture.md](docs/architecture.md#tenant-isolation).

There is no signup endpoint. Integration tests create coaches through `coach.Service.Create`, then log in over HTTP and take the token from the `gogym_auth` cookie (`testutil.Cookie`); they send it back as a Bearer header, which `AuthMiddleware` accepts alongside the cookie.

GORM is opened with `TranslateError: true` (`internal/platform/db`, `internal/platform/testutil`) so repositories can check portable errors like `errors.Is(err, gorm.ErrDuplicatedKey)` instead of parsing Postgres-specific error codes. Keep using the portable sentinels — without this config the driver-specific check would silently never match.

## API response conventions

Every `/api/v1` response body — success or failure — is the envelope defined in `internal/platform/httpx`:

```
{"success": true,  "data": {...}|[...]|null, "error": null, "meta": {...}}
{"success": false, "data": null, "error": {"code": "...", "message": "...", "fields": {...}}}
```

`success` is a real field, not inferred from the status code; `data` and `error` are both always present with one of them null; `meta` (list metadata, e.g. pagination) and `error.fields` are omitted when empty.

Handlers write it through the `httpx` helpers (`OK`, `OKWithMeta`, `Created`, `NoContent`, `Error`, `ErrorFields`, `InternalError`) and never call `c.JSON` / `c.AbortWithStatusJSON` directly — a response-rewriting middleware was considered and rejected, see [docs/architecture.md](docs/architecture.md#response-envelope). `error.code` values are the `httpx.Code*` constants; add a new one there rather than inlining a string. For a rejected request body, pass `httpx.ValidationFields(err)` to `ErrorFields` so the client gets per-field messages keyed by json name.

Answer an unexpected failure with `httpx.InternalError(c, err)`, not a hand-built 500: it sends the client only the generic message and attaches `err` to the request so the request log records the cause at error level. Never put `err.Error()` in a response body — driver and internal errors name hosts, tables and queries.

`/healthz` is the one deliberate exemption — it's an unversioned infra probe read by container healthchecks and uptime pingers that match on its exact bare `{"status":"ok"}` body, and the only place outside `internal/platform/httpx` allowed to call `c.JSON`.

Swagger annotations must show the real nested body, using the documentation-only types in `internal/platform/httpx/swagger.go`: `@Success 200 {object} httpx.SuccessEnvelope{data=coachResponse}` for a payload, and for errors the per-code type matching what the handler sends (`httpx.ValidationErrorEnvelope` for 400, `httpx.UnauthorizedEnvelope` for 401, `httpx.NotFoundEnvelope` for 404, …). Annotating the runtime `httpx.Envelope` directly instead renders every success example with a filled-in `error` object and every failure example with `"success": true` — a schema can't say "populated on failure, null on success". `TestEnvelopeDocsMatchRuntime` fails if the doc types drift from what `Envelope` marshals; a new error code gets its own doc type plus a row in that test.

## Schema conventions

`docs/er-diagram.md` is maintained by hand, not generated. Any schema change — new/dropped table, new/dropped/renamed column, new relationship — must update it in the same commit as the migration.

## Swagger/API docs conventions

Every task that adds an endpoint must include swagger annotations on its handlers as part of that task — not deferred to a cleanup pass. `internal/server/healthz.go` is the reference pattern.

Keep a one-line prose summary above a handler's `@` annotation block. Without it, the block is the whole doc comment and gofmt reformats the tab-indented `//	@Summary` lines.

`@BasePath` in `cmd/api/main.go` is `/` (root), not `/api/v1`, even though the API is versioned under `/api/v1`. Reason: `/healthz` is intentionally mounted outside the versioned group (unversioned infra probe), and Swagger 2.0's `basePath` is global — a `/api/v1` basePath made Swagger UI's "Try it out" call the wrong URL for it. Consequence: every handler's own `@Router` annotation must spell out its full path, e.g. `@Router /api/v1/coaches [post]` for versioned endpoints, `@Router /healthz [get]` for unversioned ones.

Use typed response structs in `@Success`/`@Failure` annotations, never `map[string]string` or `gin.H{...}`. Swagger/OpenAPI 2.0 can't name keys in a free-form map, so Swagger UI renders those as generic `additionalProp1/2/3` placeholders instead of real field names.

An untyped (`any`) or `map` field in a documented struct needs an example: `example:"key:value"` for maps, `extensions:"x-null-example"` for fields that are always null (`make swagger` turns the marker into `"example": null`). Without one, Swagger UI shows `"string"` or `additionalProp1/2/3`.

Run `make swagger` (`swag init`) after adding or changing annotated handlers. The generated output in `docs/swagger/` is committed (the binary imports it for its side effect of registering the spec), so a stale run means stale docs shipped in the UI.

The `/swagger/*any` route is only mounted when `internal/platform/config` reports the dev environment — never expose interactive API docs in a prod build by default.
