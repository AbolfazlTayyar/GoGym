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

## Code conventions

When a literal (string key, magic number, etc.) is used in more than one place, extract it to a named constant instead of repeating it. A literal used exactly once can stay inline — constants earn their keep on the second use, not the first.

## CI conventions

Pin tool/action versions in `.github/workflows/*.yml` (e.g. `golangci-lint-action`'s `version: v2.13.2`, not `latest`). A floating `latest` can silently change behavior between runs — we hit this once when it resolved to a stale golangci-lint v1 binary incompatible with our v2-schema `.golangci.yml`, and it also defeats the point of a fail-fast lint/unit gate if the tool itself becomes the surprise failure. Bump pinned versions deliberately, in their own commit.

## Schema conventions

`docs/er-diagram.md` is maintained by hand, not generated. Any schema change — new/dropped table, new/dropped/renamed column, new relationship — must update it in the same commit as the migration.
