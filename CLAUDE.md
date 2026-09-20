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

## Code conventions

When a literal (string key, magic number, etc.) is used in more than one place, extract it to a named constant instead of repeating it. A literal used exactly once can stay inline — constants earn their keep on the second use, not the first.

## CI conventions

Pin tool/action versions in `.github/workflows/*.yml` (e.g. `golangci-lint-action`'s `version: v2.13.2`, not `latest`). A floating `latest` can silently change behavior between runs — we hit this once when it resolved to a stale golangci-lint v1 binary incompatible with our v2-schema `.golangci.yml`, and it also defeats the point of a fail-fast lint/unit gate if the tool itself becomes the surprise failure. Bump pinned versions deliberately, in their own commit.

## Data access conventions

Tenant-owned queries (any table with a `coach_id` column) must go through `internal/tenant.Scope(db, coachID)`, never a hand-written `db.Where("coach_id = ?", ...)`. It's the single enforcement point for tenant isolation — see [docs/architecture.md](docs/architecture.md) for why and for the 404-not-403 convention that goes with it.

GORM is opened with `TranslateError: true` (`internal/db`, `internal/testutil`) so repositories can check portable errors like `errors.Is(err, gorm.ErrDuplicatedKey)` instead of parsing Postgres-specific error codes. Keep using the portable sentinels — without this config the driver-specific check would silently never match.

## Schema conventions

`docs/er-diagram.md` is maintained by hand, not generated. Any schema change — new/dropped table, new/dropped/renamed column, new relationship — must update it in the same commit as the migration.

## Swagger/API docs conventions

Every task that adds an endpoint must include swagger annotations on its handlers as part of that task — not deferred to a cleanup pass. `internal/server/healthz.go` is the reference pattern.

`@BasePath` in `cmd/api/main.go` is `/` (root), not `/api/v1`, even though the API is versioned under `/api/v1`. Reason: `/healthz` is intentionally mounted outside the versioned group (unversioned infra probe), and Swagger 2.0's `basePath` is global — a `/api/v1` basePath made Swagger UI's "Try it out" call the wrong URL for it. Consequence: every handler's own `@Router` annotation must spell out its full path, e.g. `@Router /api/v1/coaches [post]` for versioned endpoints, `@Router /healthz [get]` for unversioned ones.

Use typed response structs in `@Success`/`@Failure` annotations, never `map[string]string` or `gin.H{...}`. Swagger/OpenAPI 2.0 can't name keys in a free-form map, so Swagger UI renders those as generic `additionalProp1/2/3` placeholders instead of real field names.

Run `make swagger` (`swag init`) after adding or changing annotated handlers. The generated output in `docs/swagger/` is committed (the binary imports it for its side effect of registering the spec), so a stale run means stale docs shipped in the UI.

The `/swagger/*any` route is only mounted when `internal/config` reports the dev environment — never expose interactive API docs in a prod build by default.
