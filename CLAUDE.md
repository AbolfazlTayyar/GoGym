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
