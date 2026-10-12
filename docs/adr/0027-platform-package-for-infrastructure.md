# 0027 — Shared infrastructure lives under internal/platform

- Status: accepted
- Date: 2026-10-12
- Deciders: Abolfazl Tayyar

## Context and Problem Statement

`internal/` held feature modules (`athlete`, `coach`, …) and shared infrastructure (`config`,
`db`, `httpx`, `search`, `tenant`, `testutil`, `validate`) side by side, so a reader couldn't
tell from the listing which packages were the app and which were plumbing.

## Decision Drivers

- Features are the app. They should be the first thing a reader sees.
- A grouping folder should carry a rule a reviewer can check, not only a label.
- Short import paths for the packages features use most.

## Considered Options

1. Keep `internal/` flat and list features and infrastructure in `architecture.md`
2. Move features into `internal/modules/<module>`
3. Move shared infrastructure into `internal/platform/<pkg>` and keep features at the top

## Decision Outcome

Chosen option: **3**. A `platform` package never imports a feature package. `internal/server`
stays at the top because it imports every feature to wire them together, which would break
that rule. Feature modules keep the `internal/<module>` shape from
[0002](0002-modular-monolith.md).

Option 1 needs no code change, but the split is only documented, not visible in the tree.
Option 2 makes the core of the app the deeper path, and "modules" says nothing about what is
inside.

### Consequences

- Good: the top of `internal/` lists the features plus `server` and `platform`.
- Good: "does this import a feature?" decides where a new shared package goes.
- Bad: imports of shared packages are one segment longer.
- Bad: swag names definitions after the package path, so the generated swagger definition names
  changed from `..._internal_httpx.*` to `..._internal_platform_httpx.*`. The response shapes
  are unchanged.
- Neutral: older ADRs and completed task prompts still name the old paths (`internal/httpx`,
  `internal/tenant`, …). They are records of their time and are not rewritten.
