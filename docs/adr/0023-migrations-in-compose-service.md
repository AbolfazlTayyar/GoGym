# 0023 — Migrations run in a one-shot Compose service before the API starts

- Status: accepted
- Date: 2026-10-08
- Deciders: Abolfazl Tayyar

## Context and Problem Statement

[0004](0004-sql-migrations-as-schema-source.md) made `migrations/` the only schema source but left
applying it to a manual `make migrate-up`. A fresh `docker compose up` started the API against an
empty database, and the VPS deployment task still had to decide how migrations reach production.

## Decision Drivers

- A fresh checkout or a deploy reaches the latest schema without a step someone can forget.
- Dev and prod apply migrations the same way.
- A failed migration should stop the rollout, not crash-loop the running app.
- The API's database user shouldn't need rights to change the schema.

## Considered Options

1. Manual `make migrate-up` before starting the API
2. The API applies embedded migrations on startup
3. A one-shot `migrate` service in `docker-compose.yml` that the `api` service waits on

## Decision Outcome

Chosen option: **3**. The `migrate` service runs the `migrate/migrate` image (pinned to the
golang-migrate version in `go.mod`) against `db` once it is healthy, applies `up` and exits. `api`
depends on it with `service_completed_successfully`, so Compose won't start the API after a failed
migration. Option 1 is how a fresh database ended up with no tables. Option 2 ties every API start
to the migration: a failure marks the database dirty, the API crash-loops until someone runs
`migrate force`, and the API's user needs DDL rights. `make migrate-up` and `make migrate-down`
stay for `make run` outside Compose and for rolling back.

### Consequences

- Good: `docker compose up --build` on a fresh checkout gives a migrated database and a running API.
- Good: the VPS deployment runs the same Compose file, so its migration step is already decided.
- Bad: migrations reach the container through a bind mount, so the host running Compose needs the
  repo checkout, not only the API image.
- Bad: the image tag and `go.mod` pin golang-migrate separately and must be bumped together.
