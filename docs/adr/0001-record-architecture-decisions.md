# 0001 — Record architecture decisions

- Status: accepted
- Date: 2026-10-03
- Deciders: Abolfazl Tayyar

## Context and Problem Statement

GoGym's reasoning has been spread across `CLAUDE.md`, `docs/architecture.md`, migration comments,
task prompts and the archived roadmap. Each one holds part of the story, and none of them
records which options were rejected or when a decision was made. Later work (plan builder,
billing, a possible org layer) needs to know why things are the way they are before changing them.

## Decision Drivers

- Rejected alternatives should be written down, not only the chosen one.
- Decisions should be individually revisable without rewriting one big document.

## Considered Options

1. One ADR file per decision in `docs/adr/` (MADR template)
2. A single `docs/ADR.md` with numbered sections
3. No ADRs; keep growing `docs/architecture.md`

## Decision Outcome

Chosen option: **1 — one MADR file per decision in `docs/adr/`**, indexed by
[README.md](README.md). ADRs 0002 onward were written after the fact, from the commit history and
existing docs. Each one carries the date the decision was originally made.

Rules:
- Files are numbered `NNNN-kebab-title.md` and never renumbered.
- An accepted ADR is not edited in substance. A change of mind is a new ADR that marks the old
  one `superseded by NNNN`.
- `CLAUDE.md` holds the short operational rule. The ADR holds the reasoning and the rejected options.

### Consequences

- Good: each decision records its alternatives and trade-offs, and one decision can be superseded on its own.
- Good: `architecture.md` can stay a description of the current shape and link here for the "why".
- Bad: more files to keep in sync with `CLAUDE.md`.
