# GoGym — Docs

Start here. Every document below has one job; if two of them seem to disagree, the one listed
higher wins.

## Working documents

| Document | What it is | When you touch it |
|---|---|---|
| [tasks.md](tasks.md) | The build checklist — one task per feature, each with a paste-ready prompt and a checkpoint. Tracks progress with ✅ / ⬜. | Constantly. This is the doc you work from. |
| [spec.md](spec.md) | What v1 is and, just as importantly, what it is not. Users, entities, the six MVP features, out-of-scope list. | When scope changes. Rare on purpose. |
| [er-diagram.md](er-diagram.md) | The schema. Hand-maintained, **not** generated. | In the **same commit** as any migration — see [CLAUDE.md](../CLAUDE.md). |

## Product thinking

This exists because the goal shifted from "a tool for one coach" to "something other coaches pay
for." It is not committed scope.

| Document | What it is |
|---|---|
| [product-direction.md](product-direction.md) | The post-v1 hypothesis: why v1 alone is hard to sell, what is missing, and in what order to fix it. Explicitly unvalidated — it is reasoning from these docs, not from users. The parts of it that must happen *during* v1 are already folded into [tasks.md](tasks.md) as the tasks marked `[A1]`–`[A4]`. |

If it ever conflicts with [tasks.md](tasks.md), tasks.md is what gets built — fix the task text,
don't work from the rationale.

## Generated — do not edit by hand

| Path | Notes |
|---|---|
| [swagger/](swagger/) | Swagger/OpenAPI output from `swag init`. Regenerate with `make swagger` after changing any annotated handler. It is committed because the binary imports it for its registration side effect, so a stale run ships stale docs. The import path is `docs/swagger` — **this directory cannot be moved or renamed** without updating [internal/server/server.go](../internal/server/server.go). |

## Archive

| Path | Notes |
|---|---|
| [archive/prestart-roadmap.md](archive/prestart-roadmap.md) | The original pre-code planning doc — stack choices, architecture shape, the eleven SDLC steps. Its decisions have all been made and now live in the docs above, but the reasoning is still worth having. Kept for history; don't plan from it. |

## Conventions

- Docs are hand-written and reviewed like code. Anything generated goes in its own directory and
  says so at the top.
- A doc that has been superseded moves to `archive/` rather than being deleted — the reasoning
  behind a decision outlives the decision.
- Cross-references use relative links so they work on GitHub and in an editor alike.
